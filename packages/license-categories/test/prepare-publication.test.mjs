import assert from 'node:assert/strict'
import {createHash} from 'node:crypto'
import {mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync} from 'node:fs'
import {tmpdir} from 'node:os'
import {join} from 'node:path'
import test from 'node:test'
import {create} from 'tar'
import {
  latestPublication, packedFiles, preparePublication, publicationRef,
  publishedTarball, registryMetadata, samePayload, validateVersion,
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
])

test('release tags and selected dispatch refs are used literally', () => {
  assert.equal(publicationRef('release', {action: 'published', release: {tag_name: 'v2.9.0'}}, 'refs/heads/main'), 'refs/tags/v2.9.0')
  assert.equal(publicationRef('workflow_dispatch', {}, 'refs/heads/fix'), 'refs/heads/fix')
  assert.equal(publicationRef('workflow_dispatch', {}, 'refs/tags/v2.9.0'), 'refs/tags/v2.9.0')
  assert.equal(publicationRef('release', {action: 'published', release: {tag_name: '$(touch bad)'}}, ''), 'refs/tags/$(touch bad)')
  assert.throws(() => publicationRef('push', {}, 'refs/heads/main'), /requires/)
  assert.throws(() => publicationRef('release', {action: 'drafted'}, ''), /requires/)
})

test('payload comparison ignores version and npm bookkeeping but preserves meaningful metadata', () => {
  assert.ok(samePayload(files('1.0.0'), files('1.1.0', {_id: `${name}@1.1.0`, _npmVersion: '11', gitHead: 'abc'})))
  for (const extra of [{license: 'MIT'}, {description: 'changed'}, {exports: './new.js'}, {engines: {node: '>=24'}}]) {
    assert.ok(!samePayload(files('1.0.0'), files('1.0.0', extra)))
  }
  for (const path of ['dist/index.js', 'dist/index.d.ts', 'README.md', 'LICENSE', 'THIRD_PARTY_NOTICE.md']) {
    const changed = files('1.0.0')
    changed.set(path, Buffer.from('changed'))
    assert.ok(!samePayload(files('1.0.0'), changed), path)
  }
  assert.ok(!samePayload(files('1.0.0'), new Map()))
})

test('changed payload requires consistent, unused, semantically newer versions without auto-bumping', () => {
  validateVersion(manifest('0.1.2'), lock('0.1.2'), null)
  validateVersion(manifest('1.0.1'), lock('1.0.1'), metadata('1.0.0'))
  validateVersion(manifest('1.0.0-beta.2'), lock('1.0.0-beta.2'), metadata('1.0.0-beta.1'))
  validateVersion(manifest('1.0.0'), lock('1.0.0'), metadata('1.0.0-beta.2'))
  for (const version of ['1.0.0', '0.9.0', '1.0.0-beta.1', '1.0.0+build']) {
    assert.throws(() => validateVersion(manifest(version), lock(version), metadata('1.0.0')), /already published|semantically newer/)
  }
  assert.throws(() => validateVersion(manifest('1.1.0'), lock('1.0.0'), metadata('1.0.0')), /must agree/)
  assert.throws(() => validateVersion(manifest('v1.1.0'), lock('v1.1.0'), null), /Invalid npm semantic version/)
  const occupied = metadata('1.0.0')
  occupied.versions['1.1.0'] = {name, version: '1.1.0'}
  assert.throws(() => validateVersion(manifest('1.1.0'), lock('1.1.0'), occupied), /already published/)
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
  await assert.rejects(publishedTarball({...dist, tarball: 'https://evil.example/archive'}, destination), /unexpected tarball URL/)
  await assert.rejects(publishedTarball(dist, destination, async () => new Response('', {status: 403})), /HTTP 403/)
  await assert.rejects(publishedTarball(dist, destination, async () => new Response('wrong')), /integrity mismatch/)
  await assert.rejects(publishedTarball({...dist, integrity: undefined}, destination, async () => new Response(bytes)), /integrity hash/)
})

test('offline packed fixtures cover first release, skipped releases, docs-only changes, failures and reruns', async t => {
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
  writeFileSync(join(directory, 'dist/index.js'), 'export const answer = 42\n')
  writeFileSync(join(directory, 'README.md'), 'consumer guide')
  writeFileSync(join(directory, 'LICENSE'), 'MIT fixture')
  writeFileSync(join(directory, 'docs/operations.md'), 'excluded maintainer guide')
  const baseline = join(temporary, 'baseline.tgz')
  await create({cwd: temporary, file: baseline, gzip: true}, ['package/package.json', 'package/dist', 'package/README.md', 'package/LICENSE'])
  const baselineFiles = await packedFiles(baseline)
  assert.ok(!baselineFiles.has('docs/operations.md'))
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
  assert.ok((await packedFiles(destination)).has('dist/index.js'))
  assert.equal((await preparePublication(directory, destination, fetcher)).changed, false)
  writeFileSync(join(directory, 'docs/operations.md'), 'changed excluded documentation')
  writeVersion('1.0.1')
  assert.equal((await preparePublication(directory, destination, fetcher)).changed, false)
  writeFileSync(join(directory, 'dist/index.js'), 'export const answer = 43\n')
  writeVersion('1.0.0')
  await assert.rejects(preparePublication(directory, destination, fetcher), /already published/)
  writeVersion('1.0.2')
  assert.equal((await preparePublication(directory, destination, fetcher)).changed, true)
  // A failed attempt can be retried while npm still serves the old baseline.
  assert.equal((await preparePublication(directory, destination, fetcher)).changed, true)
  const uploaded = readFileSync(destination)
  const nowPublished = metadata('1.0.2')
  nowPublished.versions['1.0.2'].dist = {
    tarball: 'https://registry.npmjs.org/uploaded.tgz',
    integrity: `sha512-${createHash('sha512').update(uploaded).digest('base64')}`,
  }
  assert.equal((await preparePublication(directory, destination, async url =>
    String(url).endsWith('.tgz') ? new Response(uploaded) : jsonResponse(nowPublished))).changed, false)
  await assert.rejects(preparePublication(directory, destination, async () => { throw new Error('offline') }), /offline/)
})
