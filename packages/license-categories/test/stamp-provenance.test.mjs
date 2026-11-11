import assert from 'node:assert/strict'
import {execFileSync} from 'node:child_process'
import {mkdtempSync, readFileSync, rmSync, writeFileSync} from 'node:fs'
import {tmpdir} from 'node:os'
import {join} from 'node:path'
import test from 'node:test'
import {publicationProvenance, stampSource} from '../script/stamp-provenance.mjs'
import {samePackageContents} from '../script/prepare-publication.mjs'

const head = 'a'.repeat(40)

test('release provenance uses the actual checked-out release tag', () => {
  const refs = []
  const provenance = publicationProvenance('release', {
    action: 'published', release: {tag_name: 'v2.9.0'},
  }, ref => { refs.push(ref); return head })
  assert.equal(provenance, 'v2.9.0 (release)')
  assert.deepEqual(refs, ['HEAD', 'refs/tags/v2.9.0'])
  assert.throws(() => publicationProvenance('release', {
    action: 'published', release: {tag_name: 'v2.9.0'},
  }, ref => ref === 'HEAD' ? head : 'b'.repeat(40)), /does not match/)
})

test('dispatch provenance reads full checked-out HEAD, never a moving selected ref', () => {
  const refs = []
  assert.equal(publicationProvenance('workflow_dispatch', {ref: 'refs/heads/main'}, ref => {
    refs.push(ref)
    return head
  }), `${head} (dispatch)`)
  assert.deepEqual(refs, ['HEAD'])
  assert.throws(() => publicationProvenance('workflow_dispatch', {}, () => 'abcdef0'), /full checked-out/)
  assert.throws(() => publicationProvenance('push', {}, () => head), /requires/)
  const actual = execFileSync('git', ['rev-parse', 'HEAD'], {encoding: 'utf8'}).trim()
  assert.equal(publicationProvenance('workflow_dispatch', {}), `${actual} (dispatch)`)
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
    action: 'published', release: {tag_name: 'v2.9.0\ninjected'},
  }, () => head), /single-line/)
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

test('publication stamps count as package changes at this stack layer', () => {
  const generated = readFileSync(new URL('../dist/generated.js', import.meta.url), 'utf8')
  const packed = source => new Map([
    ['package.json', Buffer.from(JSON.stringify({name: '@github/license-categories', version: '0.1.2'}))],
    ['dist/generated.js', Buffer.from(source)],
  ])
  const baseline = packed(stampSource(generated, `${head} (local)`))
  assert.ok(samePackageContents(packed(stampSource(generated, `${head} (local)`)), baseline))
  for (const provenance of ['v2.9.0 (release)', `${'b'.repeat(40)} (dispatch)`]) {
    assert.ok(!samePackageContents(packed(stampSource(generated, provenance)), baseline))
  }
})

test('publication verifies first, then stamps, builds/tests, prepares, and publishes', () => {
  const workflow = readFileSync(new URL('../../../.github/workflows/publish-license-categories.yaml', import.meta.url), 'utf8')
  const commands = [
    'run: make verify-license-categories',
    'run: node script/stamp-provenance.mjs',
    'run: npm test',
    'run: node script/prepare-publication.mjs',
    'run: npm publish',
  ]
  let previous = -1
  for (const command of commands) {
    const index = workflow.indexOf(command)
    assert.ok(index > previous, `${command} is missing or out of order`)
    previous = index
  }
})

test('publication retains release/dispatch gates, exact refs, serialization and OIDC safeguards', () => {
  const workflow = readFileSync(new URL('../../../.github/workflows/publish-license-categories.yaml', import.meta.url), 'utf8')
  assert.match(workflow, /release:\s+types: \[published\]/)
  assert.match(workflow, /workflow_dispatch:/)
  assert.doesNotMatch(workflow, /^\s+(push|pull_request):/m)
  assert.match(workflow, /LICENSE_CATEGORIES_NPM_PUBLISH_ENABLED == 'true'/)
  assert.match(workflow, /github\.event_name != 'release' \|\| github\.event\.release\.prerelease == false/)
  assert.match(workflow, /group: publish-license-categories\s+cancel-in-progress: false/)
  assert.match(workflow, /format\('refs\/tags\/\{0\}', github\.event\.release\.tag_name\) \|\| github\.ref/)
  assert.match(workflow, /persist-credentials: false/)
  assert.match(workflow, /id-token: write/)
  assert.doesNotMatch(workflow, /contents: write|NPM_TOKEN|git push|git commit/)
  assert.match(workflow, /node-version: 24/)
  assert.match(workflow, /npm install -g npm@11\.20\.0/)
  assert.match(workflow, /cache: npm\s+cache-dependency-path: packages\/license-categories\/package-lock\.json/)
  assert.match(workflow, /if: steps\.prepare\.outputs\.changed == 'true'/)
  assert.match(workflow, /npm publish "\$RUNNER_TEMP\/license-categories\.tgz" --ignore-scripts --access public --provenance --registry=https:\/\/registry\.npmjs\.org/)
  assert.ok([...workflow.matchAll(/uses: [^\n]+/g)].every(([action]) => /@[a-f0-9]{40}\b/.test(action)))
})

test('separate read-only validation covers relevant PR and main paths without publishing or duplicating Go tests', () => {
  const workflow = readFileSync(new URL('../../../.github/workflows/validate-license-categories.yaml', import.meta.url), 'utf8')
  assert.match(workflow, /pull_request:\s+paths:/)
  assert.match(workflow, /push:\s+branches: \[main\]\s+paths:/)
  for (const path of [
    'cmd/license-categories/**', 'spdxexp/spdxlicenses/**', 'packages/license-categories/**',
    'go.mod', 'go.sum', 'Makefile', '.github/workflows/validate-license-categories.yaml',
    '.github/workflows/publish-license-categories.yaml',
  ]) {
    assert.equal(workflow.split(`- ${path}\n`).length - 1, 2, path)
  }
  assert.match(workflow, /contents: read/)
  assert.match(workflow, /run: make verify-license-categories/)
  assert.match(workflow, /run: npm ci/)
  assert.match(workflow, /run: npm test/)
  assert.doesNotMatch(workflow, /npm publish|id-token: write|contents: write|go test/)
})
