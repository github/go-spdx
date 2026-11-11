import assert from 'node:assert/strict'
import {spawnSync} from 'node:child_process'
import {createHash} from 'node:crypto'
import {existsSync, mkdtempSync, mkdirSync, readFileSync, readdirSync, rmSync, writeFileSync} from 'node:fs'
import {tmpdir} from 'node:os'
import {join} from 'node:path'
import test from 'node:test'
import {gzipSync} from 'node:zlib'
import {create, Header, Pax} from 'tar'
import {
  latestPublication, packedFiles, preparePublication, publicationRef,
  publishedTarball, registryMetadata, samePackageContents, validateVersion,
} from '../script/prepare-publication.mjs'

const name = '@github/license-categories'
const manifest = version => ({name, version, type: 'module', files: ['dist', 'README.md', 'LICENSE']})
const lock = version => ({name, version, packages: {'': {name, version}}})
const metadata = version => ({
  name, versions: {[version]: {name, version}}, time: {[version]: '2026-01-01T00:00:00Z'},
})
const jsonResponse = (body, status = 200) => new Response(JSON.stringify(body), {status})
const files = (version, extra = {}) => new Map([
  ['package.json', Buffer.from(JSON.stringify({...manifest(version), ...extra}))],
  ['dist/index.js', Buffer.from('export const answer = 42\n')],
  ['dist/generated.js', Buffer.from('// go-spdx: (devel)\nexport const data = {}\n')],
])

function writeEmptyArchive(destination, paths, type = 'File') {
  const blocks = paths.flatMap(path => {
    const header = new Header({type, size: 0, mode: 0o644, ...(typeof path === 'string' ? {path} : path)})
    const block = Buffer.alloc(512)
    return header.encode(block) ? [new Pax({path}).encode(), block] : [block]
  })
  writeFileSync(destination, gzipSync(Buffer.concat([...blocks, Buffer.alloc(1024)])))
}

test('packed archives bound entry count and UTF-8 path bytes, including directories', async t => {
  const temporary = mkdtempSync(join(tmpdir(), 'license-categories-archive-bounds-'))
  t.after(() => rmSync(temporary, {recursive: true, force: true}))
  const directoryPath = `package/${'d'.repeat(60)}/${'é'.repeat(8157)}/`
  assert.equal(Buffer.byteLength(directoryPath), 16_384)
  const cases = [
    ['file-count-boundary', Array.from({length: 100}, (_, index) => `package/${index}`), 'File', 100],
    ['file-count-exceeded', Array.from({length: 101}, (_, index) => `package/${index}`), 'File', /entry count exceeds 100/],
    ['directory-count-exceeded', Array(101).fill('package/'), 'Directory', /entry count exceeds 100/],
    ['path-bytes-boundary', Array(64).fill(directoryPath), 'Directory', 0],
    ['path-bytes-exceeded', Array(65).fill(directoryPath), 'Directory', /path bytes exceed 1 MiB/],
  ]
  for (const [label, paths, type, expected] of cases) {
    const archive = join(temporary, `${label}.tgz`)
    writeEmptyArchive(archive, paths, type)
    if (typeof expected === 'number') {
      assert.equal((await packedFiles(archive)).size, expected, label)
    } else {
      const result = spawnSync(process.execPath, ['--input-type=module', '-e',
        `import {packedFiles} from ${JSON.stringify(new URL('../script/prepare-publication.mjs', import.meta.url).href)};
        await packedFiles(process.argv[1]);`, archive], {encoding: 'utf8', timeout: 10_000})
      assert.ifError(result.error)
      assert.equal(result.status, 1, label)
      assert.match(result.stderr, expected, label)
    }
  }
})

test('invalid published archives reject normally, clean up, and do not write a candidate', async t => {
  const temporary = mkdtempSync(join(tmpdir(), 'license-categories-archive-errors-'))
  t.after(() => rmSync(temporary, {recursive: true, force: true}))
  const directory = join(temporary, 'package')
  mkdirSync(directory)
  writeFileSync(join(directory, 'package.json'), JSON.stringify(manifest('1.0.1')))
  writeFileSync(join(directory, 'package-lock.json'), JSON.stringify(lock('1.0.1')))
  const cases = [
    ['symlink', [{path: 'package/link', type: 'SymbolicLink', linkpath: 'target'}], /Unsupported package archive entry/],
    ['outside', ['outside/file'], /Unsupported package archive entry/],
    ['traversal', ['package/../file'], /Invalid or duplicate/],
    ['dot', ['package/./file'], /Invalid or duplicate/],
    ['empty-segment', ['package/nested//file'], /Invalid or duplicate/],
    ['backslash', ['package/nested\\file'], /Invalid or duplicate/],
    ['duplicate', ['package/file', 'package/file'], /Invalid or duplicate/],
    ['oversized', [{path: 'package/large', size: 100 * 1024 * 1024 + 1}], /contents exceed 100 MiB/],
    ['entry-count', Array(101).fill('package/'), /entry count exceeds 100/],
    ['path-bytes', Array(65).fill(`package/${'d'.repeat(60)}/${'é'.repeat(8157)}/`), /path bytes exceed 1 MiB/],
  ]
  for (const [label, paths, expected] of cases) {
    const archive = join(temporary, `${label}.tgz`)
    const scratch = join(temporary, `${label}-scratch`)
    const destination = join(temporary, `${label}-candidate.tgz`)
    mkdirSync(scratch)
    writeEmptyArchive(archive, paths)
    const result = spawnSync(process.execPath, ['--input-type=module', '-e', `
      import {readFileSync} from 'node:fs';
      import {createHash} from 'node:crypto';
      import {preparePublication} from ${JSON.stringify(new URL('../script/prepare-publication.mjs', import.meta.url).href)};
      const bytes = readFileSync(process.argv[1]);
      const metadata = ${JSON.stringify(metadata('1.0.0'))};
      metadata.versions['1.0.0'].dist = {
        tarball: 'https://registry.npmjs.org/fixture.tgz',
        integrity: 'sha512-' + createHash('sha512').update(bytes).digest('base64'),
      };
      try {
        await preparePublication(process.argv[2], process.argv[3], async url =>
          String(url).endsWith('.tgz') ? new Response(bytes) : new Response(JSON.stringify(metadata)));
        throw new Error('Unexpected successful preparation');
      } catch (error) {
        console.log('caught: ' + error.message);
      }
    `, archive, directory, destination], {
      encoding: 'utf8', timeout: 10_000,
      env: {...process.env, TMPDIR: scratch, NODE_DISABLE_COMPILE_CACHE: '1'},
    })
    assert.ifError(result.error)
    assert.equal(result.status, 0, `${label}: ${result.stderr}`)
    assert.match(result.stdout, expected, label)
    assert.deepEqual(readdirSync(scratch), [], `${label}: temporary files remain`)
    assert.equal(existsSync(destination), false, `${label}: candidate was written`)
  }
})

test('release tags and selected dispatch refs are used literally', () => {
  assert.equal(publicationRef('release', {action: 'published', release: {tag_name: 'v2.9.0'}}, 'refs/heads/main'), 'refs/tags/v2.9.0')
  assert.equal(publicationRef('workflow_dispatch', {}, 'refs/heads/fix'), 'refs/heads/fix')
  assert.equal(publicationRef('workflow_dispatch', {}, 'refs/tags/v2.9.0'), 'refs/tags/v2.9.0')
  assert.equal(publicationRef('release', {action: 'published', release: {tag_name: '$(touch bad)'}}, ''), 'refs/tags/$(touch bad)')
  assert.throws(() => publicationRef('push', {}, 'refs/heads/main'), /requires/)
  assert.throws(() => publicationRef('release', {action: 'drafted'}, ''), /requires/)
})

test('comparison ignores only manifest version, npm bookkeeping and top-level JSON key order', () => {
  assert.ok(samePackageContents(files('1.0.0'), files('1.1.0', {_id: `${name}@1.1.0`, _npmVersion: '11', gitHead: 'abc'})))
  const reordered = files('1.1.0')
  reordered.set('package.json', Buffer.from(JSON.stringify({
    files: ['dist', 'README.md', 'LICENSE'], type: 'module', version: '1.1.0', name,
  })))
  assert.ok(samePackageContents(files('1.0.0'), reordered))
  assert.throws(() => samePackageContents(files('1.0.0'), new Map()), /missing package.json/)
  assert.throws(() => samePackageContents(files('1.0.0'), files('1.0.0', {name: 'unexpected'})), /manifest name/)
  const malformed = files('1.0.0')
  malformed.set('package.json', Buffer.from('{'))
  assert.throws(() => samePackageContents(files('1.0.0'), malformed), SyntaxError)
})

for (const [label, before, after] of [
  [
    'nested exports conditions',
    {exports: {'.': {node: {import: './node.js', default: './fallback.js'}}}},
    {exports: {'.': {node: {default: './fallback.js', import: './node.js'}}}},
  ],
  [
    'imports conditions',
    {imports: {'#internal': {node: './node.js', default: './fallback.js'}}},
    {imports: {'#internal': {default: './fallback.js', node: './node.js'}}},
  ],
  [
    'overlapping typesVersions ranges',
    {typesVersions: {'>=5.0': {'*': ['modern.d.ts']}, '>=4.0': {'*': ['legacy.d.ts']}}},
    {typesVersions: {'>=4.0': {'*': ['legacy.d.ts']}, '>=5.0': {'*': ['modern.d.ts']}}},
  ],
]) {
  test(`comparison preserves ${label} order`, () => {
    const original = files('1.0.0', before)
    const reordered = files('1.1.0', after)
    assert.ok(samePackageContents(original, files('1.1.0', before)))
    assert.ok(!samePackageContents(original, reordered))
    assert.ok(!samePackageContents(reordered, original))
  })
}

test('every packaged file, provenance and consumer metadata remain significant', () => {
  for (const extra of [{license: 'MIT'}, {description: 'changed'}, {exports: './new.js'}, {engines: {node: '>=24'}}]) {
    assert.ok(!samePackageContents(files('1.0.0'), files('1.0.0', extra)))
  }
  for (const path of ['dist/index.js', 'dist/index.d.ts', 'dist/generated.js', 'dist/generated.d.ts', 'README.md', 'LICENSE', 'THIRD_PARTY_NOTICE.md']) {
    const changed = files('1.0.0')
    changed.set(path, Buffer.from('changed'))
    assert.ok(!samePackageContents(files('1.0.0'), changed), path)
    if (files('1.0.0').has(path)) {
      changed.delete(path)
      assert.ok(!samePackageContents(files('1.0.0'), changed), `removed ${path}`)
    }
  }
  const renamed = files('1.0.0')
  renamed.set('dist/renamed.js', renamed.get('dist/index.js'))
  renamed.delete('dist/index.js')
  assert.ok(!samePackageContents(files('1.0.0'), renamed))
  const stamped = files('1.0.0')
  stamped.set('dist/generated.js', Buffer.from('// go-spdx: v2.9.0 (release)\nexport const data = {}\n'))
  assert.ok(!samePackageContents(files('1.0.0'), stamped))
})

test('eligible publications require consistent, unused, semantically newer versions without auto-bumping', () => {
  validateVersion(manifest('0.1.2'), lock('0.1.2'), null)
  validateVersion(manifest('1.0.1'), lock('1.0.1'), metadata('1.0.0'))
  validateVersion(manifest('1.0.0-beta.2'), lock('1.0.0-beta.2'), metadata('1.0.0-beta.1'))
  validateVersion(manifest('1.0.0'), lock('1.0.0'), metadata('1.0.0-beta.2'))
  for (const version of ['1.0.0', '0.9.0', '1.0.0-beta.1', '1.0.0+build']) {
    assert.throws(() => validateVersion(manifest(version), lock(version), metadata('1.0.0')), /already published|semantically newer/)
  }
  for (const version of ['', 'v1.1.0', '=1.1.0', '1.1', 'bad']) {
    assert.throws(() => validateVersion(manifest(version), lock(version), null), /Invalid npm semantic version/)
  }
  assert.throws(() => validateVersion(manifest('1.1.0'), lock('1.0.0'), metadata('1.0.0')), /must agree/)
  assert.throws(() => validateVersion(manifest('1.1.0'), {...lock('1.1.0'), name: 'wrong'}, null), /must agree/)
  const occupied = metadata('1.0.0')
  occupied.versions['1.1.0'] = {name, version: '1.1.0'}
  assert.throws(() => validateVersion(manifest('1.1.0'), lock('1.1.0'), occupied), /already published/)
  assert.throws(() => validateVersion(manifest('1.0.1'), lock('1.0.1'), occupied), /semantically newer/)
})

test('baseline is most recently published tarball, not a Go tag or latest dist-tag', () => {
  const published = metadata('1.0.0')
  published.versions['1.1.0-beta.1'] = {name, version: '1.1.0-beta.1'}
  published.time['1.1.0-beta.1'] = '2026-02-01T00:00:00Z'
  published['dist-tags'] = {latest: '1.0.0'}
  assert.equal(latestPublication(published), '1.1.0-beta.1')
  assert.throws(() => latestPublication({...published, time: {}}), /publication time/)
  assert.throws(() => latestPublication({...published, versions: {}}), /no published versions/)
})

test('only explicit npm package absence counts as first publication', async () => {
  assert.equal(await registryMetadata(async () => jsonResponse({error: 'Not found'}, 404)), null)
  assert.deepEqual(await registryMetadata(async () => jsonResponse(metadata('1.0.0'))), metadata('1.0.0'))
  for (const status of [401, 403, 429, 500]) {
    await assert.rejects(registryMetadata(async () => jsonResponse({error: 'failure'}, status)), new RegExp(`HTTP ${status}`))
  }
  await assert.rejects(registryMetadata(async () => jsonResponse({error: 'permission denied'}, 404)), /inconclusive/)
  await assert.rejects(registryMetadata(async () => new Response('<html>proxy failure</html>', {status: 404})), SyntaxError)
  await assert.rejects(registryMetadata(async () => jsonResponse({})), /invalid package metadata/)
  await assert.rejects(registryMetadata(async () => { throw new Error('network failure') }), /network failure/)
})

test('tarball download validates origin, status, and integrity', async t => {
  const temporary = mkdtempSync(join(tmpdir(), 'license-categories-download-'))
  t.after(() => rmSync(temporary, {recursive: true, force: true}))
  const bytes = Buffer.from('fixture')
  const dist = {
    tarball: 'https://registry.npmjs.org/@github/license-categories/-/license-categories-1.0.0.tgz',
    integrity: `sha512-${createHash('sha512').update(bytes).digest('base64')}`,
  }
  const destination = join(temporary, 'published.tgz')
  await publishedTarball(dist, destination, async () => new Response(bytes))
  assert.deepEqual(readFileSync(destination), bytes)
  await publishedTarball({...dist, integrity: undefined, shasum: createHash('sha1').update(bytes).digest('hex')},
    destination, async () => new Response(bytes))
  await assert.rejects(publishedTarball({...dist, tarball: 'https://evil.example/archive'}, destination), /unexpected tarball URL/)
  await assert.rejects(publishedTarball(dist, destination, async () => new Response('', {status: 403})), /HTTP 403/)
  await assert.rejects(publishedTarball(dist, destination, async () => new Response('wrong')), /integrity mismatch/)
  await assert.rejects(publishedTarball({...dist, integrity: undefined}, destination, async () => new Response(bytes)), /integrity hash/)
})

test('packed fixtures cover whole-package gates, versions, excluded docs, failures and reruns', async t => {
  const temporary = mkdtempSync(join(tmpdir(), 'license-categories-release-'))
  t.after(() => rmSync(temporary, {recursive: true, force: true}))
  const directory = join(temporary, 'package')
  mkdirSync(join(directory, 'dist'), {recursive: true})
  mkdirSync(join(directory, 'docs'))
  const writeVersion = version => {
    writeFileSync(join(directory, 'package.json'), JSON.stringify(manifest(version)))
    writeFileSync(join(directory, 'package-lock.json'), JSON.stringify(lock(version)))
  }
  writeVersion('1.0.0')
  const baselineSource = '// go-spdx: (devel)\nexport const answer = 42\n'
  writeFileSync(join(directory, 'dist/index.js'), baselineSource)
  writeFileSync(join(directory, 'README.md'), 'consumer guide')
  writeFileSync(join(directory, 'LICENSE'), 'MIT fixture')
  writeFileSync(join(directory, 'docs/operations.md'), 'excluded maintainer guide')
  const baseline = join(temporary, 'baseline.tgz')
  await create({cwd: temporary, file: baseline, gzip: true}, ['package/package.json', 'package/dist', 'package/README.md', 'package/LICENSE'])
  const bytes = readFileSync(baseline)
  const published = metadata('1.0.0')
  published.versions['1.0.0'].dist = {
    tarball: 'https://registry.npmjs.org/fixture.tgz',
    integrity: `sha512-${createHash('sha512').update(bytes).digest('base64')}`,
  }
  const fetcher = async url => String(url).endsWith('.tgz') ? new Response(bytes) : jsonResponse(published)
  const destination = join(temporary, 'candidate.tgz')
  const first = await preparePublication(directory, destination, async () => jsonResponse({error: 'Not found'}, 404))
  assert.equal(first.changed, true)
  const candidate = await packedFiles(destination)
  assert.ok(candidate.has('dist/index.js'))
  assert.ok(!candidate.has('docs/operations.md'))
  assert.equal((await preparePublication(directory, destination, fetcher)).changed, false)
  writeFileSync(join(directory, 'docs/operations.md'), 'changed excluded documentation')
  writeVersion('1.0.1')
  assert.equal((await preparePublication(directory, destination, fetcher)).changed, false)
  // Dispatch uses this same content gate; no forced-publication option exists.
  writeFileSync(join(directory, 'dist/index.js'), baselineSource.replace('(devel)', 'v2.9.0 (release)'))
  assert.equal((await preparePublication(directory, destination, fetcher)).changed, true)
  writeVersion('1.0.0')
  await assert.rejects(preparePublication(directory, destination, fetcher), /already published/)
  writeVersion('0.9.0')
  await assert.rejects(preparePublication(directory, destination, fetcher), /semantically newer/)
  writeVersion('1.0.1')
  writeFileSync(join(directory, 'package-lock.json'), JSON.stringify(lock('1.0.0')))
  await assert.rejects(preparePublication(directory, destination, fetcher), /must agree/)
  writeVersion('1.0.1')
  writeFileSync(join(directory, 'dist/index.js'), baselineSource)
  for (const path of ['README.md', 'LICENSE', 'dist/index.js']) {
    const original = readFileSync(join(directory, path))
    writeFileSync(join(directory, path), 'changed packaged content')
    assert.equal((await preparePublication(directory, destination, fetcher)).changed, true, path)
    writeFileSync(join(directory, path), original)
  }
  writeFileSync(join(directory, 'dist/index.js'), 'export const answer = 43\n')
  assert.equal((await preparePublication(directory, destination, fetcher)).changed, true)
  assert.equal((await preparePublication(directory, destination, fetcher)).changed, true)
  const uploaded = readFileSync(destination)
  assert.equal((await packedFiles(destination)).get('dist/index.js').toString(), 'export const answer = 43\n')
  const nowPublished = metadata('1.0.1')
  nowPublished.versions['1.0.1'].dist = {
    tarball: 'https://registry.npmjs.org/uploaded.tgz',
    integrity: `sha512-${createHash('sha512').update(uploaded).digest('base64')}`,
  }
  assert.equal((await preparePublication(directory, destination, async url =>
    String(url).endsWith('.tgz') ? new Response(uploaded) : jsonResponse(nowPublished))).changed, false)
  await assert.rejects(preparePublication(directory, destination, async () => { throw new Error('offline') }), /offline/)
  // First publication still enforces the manifest/lockfile contract.
  writeFileSync(join(directory, 'package-lock.json'), JSON.stringify(lock('1.0.0')))
  await assert.rejects(preparePublication(directory, destination, async () => jsonResponse({error: 'Not found'}, 404)), /must agree/)
})
