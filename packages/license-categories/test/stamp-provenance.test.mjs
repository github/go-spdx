import assert from 'node:assert/strict'
import {execFileSync, spawnSync} from 'node:child_process'
import {mkdirSync, mkdtempSync, readFileSync, realpathSync, rmSync, writeFileSync} from 'node:fs'
import {tmpdir} from 'node:os'
import {join} from 'node:path'
import test from 'node:test'
import {publicationProvenance, stampSource} from '../script/stamp-provenance.mjs'
import {sameLicenseData} from '../script/prepare-publication.mjs'

const head = 'a'.repeat(40)

test('stamping requires the event commit and leaves source untouched on failure', t => {
  const directory = realpathSync(mkdtempSync(join(tmpdir(), 'license-categories-event-sha-')))
  t.after(() => rmSync(directory, {recursive: true, force: true}))
  mkdirSync(join(directory, 'script'))
  mkdirSync(join(directory, 'src'))
  const script = join(directory, 'script', 'stamp-provenance.mjs')
  const generated = join(directory, 'src', 'generated.ts')
  const eventPath = join(directory, 'event.json')
  const source = '// go-spdx: original (local)\nexport const fixture = true;\n'
  writeFileSync(script, readFileSync(new URL('../script/stamp-provenance.mjs', import.meta.url)))
  const git = (...args) => execFileSync('git', args, {cwd: directory, encoding: 'utf8'}).trim()
  git('init', '--quiet')
  const commit = () => git('-c', 'user.name=Test', '-c', 'user.email=test@example.com',
    '-c', 'commit.gpgsign=false', '-c', `core.hooksPath=${join(directory, 'hooks')}`,
    'commit', '--quiet', '--allow-empty', '-m', 'Test event provenance')
  commit()
  const eventCommit = git('rev-parse', 'HEAD')
  commit()
  const checkoutCommit = git('rev-parse', 'HEAD')
  git('tag', 'v2.9.0')

  for (const [eventName, event, expected] of [
    ['workflow_dispatch', {}, `${checkoutCommit} (dispatch)`],
    ['release', {action: 'released', release: {tag_name: 'v2.9.0'}}, 'v2.9.0 (release)'],
  ]) {
    writeFileSync(eventPath, JSON.stringify(event))
    const run = eventSHA => {
      writeFileSync(generated, source)
      const env = {...process.env, GITHUB_ACTIONS: 'true', GITHUB_EVENT_NAME: eventName, GITHUB_EVENT_PATH: eventPath}
      if (eventSHA === undefined) delete env.GITHUB_SHA
      else env.GITHUB_SHA = eventSHA
      const result = spawnSync(process.execPath, [script], {cwd: directory, env, encoding: 'utf8', timeout: 10_000})
      assert.ifError(result.error)
      return result
    }

    for (const eventSHA of [undefined, '', 'abcdef0', 'g'.repeat(40), eventCommit]) {
      const result = run(eventSHA)
      assert.equal(result.status, 1, `${eventName}: ${eventSHA}\n${result.stderr}`)
      assert.match(result.stderr, /GITHUB_SHA/)
      assert.equal(result.stdout, '')
      assert.equal(readFileSync(generated, 'utf8'), source)
    }
    const result = run(checkoutCommit)
    assert.equal(result.status, 0, result.stderr)
    assert.equal(readFileSync(generated, 'utf8'), stampSource(source, expected))
  }
})

test('release provenance uses the actual checked-out release tag', () => {
  const refs = []
  const provenance = publicationProvenance('release', {
    action: 'released', release: {tag_name: 'v2.9.0'},
  }, head, ref => { refs.push(ref); return head })
  assert.equal(provenance, 'v2.9.0 (release)')
  assert.deepEqual(refs, ['HEAD', 'refs/tags/v2.9.0'])
  assert.throws(() => publicationProvenance('release', {
    action: 'released', release: {tag_name: 'v2.9.0'},
  }, head, ref => ref === 'HEAD' ? head : 'b'.repeat(40)), /does not match/)
})

test('only released events stamp initial stable releases and prerelease promotions', () => {
  for (const tag of ['v2.9.0', 'v2.9.1-rc.1']) {
    assert.equal(publicationProvenance('release', {
      action: 'released', release: {tag_name: tag, prerelease: false},
    }, head, () => head), `${tag} (release)`)
  }
  for (const action of ['published', 'prereleased', 'created', 'edited']) {
    assert.throws(() => publicationProvenance('release', {
      action, release: {tag_name: 'v2.9.0', prerelease: false},
    }, head, () => head), /requires a released event/)
  }
})

test('dispatch provenance reads full checked-out HEAD, never a moving selected ref', () => {
  const refs = []
  assert.equal(publicationProvenance('workflow_dispatch', {ref: 'refs/heads/main'}, head, ref => {
    refs.push(ref)
    return head
  }), `${head} (dispatch)`)
  assert.deepEqual(refs, ['HEAD'])
  assert.throws(() => publicationProvenance('workflow_dispatch', {}, head, () => 'abcdef0'), /full checked-out/)
  assert.throws(() => publicationProvenance('push', {}, head, () => head), /requires/)
  const actual = execFileSync('git', ['rev-parse', 'HEAD'], {encoding: 'utf8'}).trim()
  assert.equal(publicationProvenance('workflow_dispatch', {}, actual), `${actual} (dispatch)`)
})

test('stamping modifies only the go-spdx header and preserves literal tag text', () => {
  const source = readFileSync(new URL('../src/generated.ts', import.meta.url), 'utf8')
  const provenance = 'v2.9.0-$& (release)'
  const stamped = stampSource(source, provenance)
  assert.ok(stamped.includes(`// go-spdx: ${provenance}\n`))
  const withoutHeader = value => value.replace(/^\/\/ go-spdx: .*$/m, '')
  assert.equal(withoutHeader(stamped), withoutHeader(source))
  assert.throws(() => stampSource('no header', provenance), /exactly one/)
  assert.throws(() => stampSource('// go-spdx: one\n// go-spdx: two', provenance), /exactly one/)
  assert.throws(() => publicationProvenance('release', {
    action: 'released', release: {tag_name: 'v2.9.0\ninjected'},
  }, head, () => head), /single-line/)
})

test('release tags reject every TypeScript comment terminator', () => {
  for (const terminator of ['\r', '\n', '\u2028', '\u2029']) {
    const tag = `v2.9.0${terminator}export const injected = true;`
    const refs = []
    assert.throws(() => publicationProvenance('release', {
      action: 'released', release: {tag_name: tag},
    }, head, ref => { refs.push(ref); return head }), /single-line/)
    assert.deepEqual(refs, ['HEAD'])
  }
})

test('stamp values reject every TypeScript comment terminator', () => {
  const source = readFileSync(new URL('../src/generated.ts', import.meta.url), 'utf8')
  for (const terminator of ['\r', '\n', '\u2028', '\u2029']) {
    const provenance = `v2.9.0${terminator}export const injected = true; (release)`
    assert.throws(() => stampSource(source, provenance), /single-line/)
  }
})

test('the compiled publication artifact retains the stamped header', t => {
  const directory = mkdtempSync(join(tmpdir(), 'license-categories-stamp-'))
  t.after(() => rmSync(directory, {recursive: true, force: true}))
  const source = readFileSync(new URL('../src/generated.ts', import.meta.url), 'utf8')
  writeFileSync(join(directory, 'generated.ts'), stampSource(source, `${head} (dispatch)`))
  execFileSync(process.execPath, [
    new URL('../node_modules/typescript/bin/tsc', import.meta.url).pathname,
    join(directory, 'generated.ts'), '--target', 'ES2022', '--module', 'ES2022',
  ])
  assert.ok(readFileSync(join(directory, 'generated.js'), 'utf8').includes(`// go-spdx: ${head} (dispatch)`))
})

test('real built local, release and dispatch stamps are equivalent but data changes remain significant', t => {
  const directory = mkdtempSync(join(tmpdir(), 'license-categories-data-stamps-'))
  t.after(() => rmSync(directory, {recursive: true, force: true}))
  const source = readFileSync(new URL('../src/generated.ts', import.meta.url), 'utf8')
  const provenances = [
    `${head} (local)`,
    publicationProvenance('release', {action: 'released', release: {tag_name: 'v2.9.0'}}, head, () => head),
    publicationProvenance('workflow_dispatch', {}, 'b'.repeat(40), () => 'b'.repeat(40)),
  ]
  const inputs = provenances.map((provenance, index) => {
    const path = join(directory, `generated-${index}.ts`)
    writeFileSync(path, stampSource(source, provenance)
      .replace(/^(\/\/ spdx-license-list-version: ).*$/m, `$1version-${index}`)
      .replace(/^(\/\/ scancode-licensedb commit: ).*$/m, `$1${String(index).repeat(40)}`))
    return path
  })
  const changedPath = join(directory, 'changed.ts')
  assert.ok(source.includes("'Safe for most'"))
  writeFileSync(changedPath, stampSource(source.replace("'Safe for most'", "'Changed description'"), provenances[2]))
  execFileSync(process.execPath, [
    new URL('../node_modules/typescript/bin/tsc', import.meta.url).pathname,
    ...inputs, changedPath, '--target', 'ES2022', '--module', 'ES2022',
  ])
  const packed = path => new Map([
    ['package.json', Buffer.from(JSON.stringify({name: '@github/license-categories'}))],
    ['dist/generated.js', readFileSync(path)],
  ])
  const variants = inputs.map((path, index) => {
    const js = path.replace(/\.ts$/, '.js')
    assert.ok(readFileSync(js, 'utf8').includes(`// go-spdx: ${provenances[index]}`))
    return packed(js)
  })
  const changed = packed(join(directory, 'changed.js'))
  for (const left of variants) {
    for (const right of variants) assert.ok(sameLicenseData(left, right))
    assert.ok(!sameLicenseData(left, changed))
    assert.ok(!sameLicenseData(changed, left))
  }
})
