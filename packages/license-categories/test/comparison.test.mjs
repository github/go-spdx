import assert from 'node:assert/strict'
import {execFileSync} from 'node:child_process'
import {mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync} from 'node:fs'
import {tmpdir} from 'node:os'
import {join} from 'node:path'
import test from 'node:test'
import {generatedDataChanged} from '../../../script/compare-license-categories.mjs'

test('publication uses the comparison output after testing for either trigger', () => {
  const workflow = readFileSync(new URL('../../../.github/workflows/publish-license-categories.yaml', import.meta.url), 'utf8')
  const comparison = workflow.indexOf('      - name: Compare published generated data\n')
  const publish = workflow.indexOf('      - name: Publish to npm using OIDC\n')
  assert.ok(comparison > workflow.indexOf('      - name: Test package tarball\n'))
  assert.ok(publish > comparison)
  assert.match(workflow.slice(comparison, publish), /appendFileSync\(process\.env\.GITHUB_OUTPUT, `changed=\$\{changed\}\\n`\)/)
  assert.match(workflow.slice(publish), /if: steps\.comparison\.outputs\.changed == 'true'/)
})

test('compares only compiled generated data against the last publication', async t => {
  const root = mkdtempSync(join(tmpdir(), 'license-categories-comparison-'))
  t.after(() => rmSync(root, {recursive: true, force: true}))
  mkdirSync(join(root, 'dist'))
  mkdirSync(join(root, 'package/dist'), {recursive: true})
  const generated = readFileSync(new URL('../dist/generated.js', import.meta.url), 'utf8')
  const local = join(root, 'dist/generated.js')
  writeFileSync(local, generated)
  writeFileSync(join(root, 'package.json'), '{"name":"@github/license-categories","version":"0.1.2"}')
  const archive = source => {
    writeFileSync(join(root, 'package/dist/generated.js'), source)
    return execFileSync('tar', ['-czf', '-', '-C', root, 'package/dist/generated.js'])
  }
  const metadata = {
    'dist-tags': {latest: '2.0.0'},
    versions: {
      '2.0.0': {dist: {tarball: 'https://registry.npmjs.org/older.tgz'}},
      '1.0.1': {dist: {tarball: 'https://registry.npmjs.org/newest.tgz'}},
    },
    time: {'2.0.0': '2026-01-01T00:00:00Z', '1.0.1': '2026-02-01T00:00:00Z'},
  }
  const compare = (published, selectedMetadata = metadata) => {
    const tarball = archive(published)
    return generatedDataChanged(root, async url => {
      if (url === 'https://registry.npmjs.org/%40github%2Flicense-categories') {
        return Response.json(selectedMetadata)
      }
      assert.equal(url, 'https://registry.npmjs.org/newest.tgz')
      return new Response(tarball)
    })
  }

  await t.test('first publication is eligible', async () => {
    assert.equal(await generatedDataChanged(root, async () => new Response('', {status: 404})), true)
  })
  await t.test('identical compiled data skips', async () => {
    assert.equal(await compare(generated), false)
  })
  await t.test('release, dispatch, and bootstrap provenance do not affect the decision', async () => {
    for (const provenance of ['v2.0.1', `${'a'.repeat(40)} (dispatch)`, 'manual-bootstrap']) {
      const published = generated
        .replace(/^\/\/ go-spdx: .+$/m, `// go-spdx: ${provenance}`)
        .replace(/^\/\/ spdx-license-list-version: .+$/m, '// spdx-license-list-version: older')
        .replace(/^\/\/ scancode-licensedb commit: .+$/m, '// scancode-licensedb commit: older')
      assert.notEqual(published, generated)
      assert.equal(await compare(published), false)
    }
  })
  await t.test('changed generated data publishes', async () => {
    assert.ok(generated.includes('Safe for most'))
    assert.equal(await compare(generated.replace('Safe for most', 'Changed category')), true)
  })
  await t.test('unrelated package files do not affect the decision', async () => {
    for (const file of ['README.md', 'dist/index.js', 'THIRD_PARTY_NOTICE.md']) {
      writeFileSync(join(root, file), 'unrelated changes')
    }
    writeFileSync(join(root, 'package.json'), '{"name":"@github/license-categories","version":"9.0.0","description":"changed"}')
    assert.equal(await compare(generated), false)
  })
  await t.test('the entire leading comment header is ignored on both files', async () => {
    assert.equal(await compare(generated.replace('DO NOT EDIT.', 'Changed warning.')), false)
    assert.equal(await compare(generated.replace(/^(?:\/\/[^\n]*\n)+/, '// Completely different provenance\n// Additional metadata\n')), false)
    assert.equal(await compare(generated.replace(/^(?:\/\/[^\n]*\n)+/, '')), false)
    writeFileSync(local, generated.replace(/^(?:\/\/[^\n]*\n)+/, '// Local provenance\n'))
    try {
      assert.equal(await compare(generated), false)
      assert.equal(await compare(generated.replace('Safe for most', 'Changed category')), true)
    } finally {
      writeFileSync(local, generated)
    }
  })
  await t.test('comments after the leading header remain significant', async () => {
    assert.equal(await compare(`${generated}// go-spdx: not a leading provenance header\n`), true)
  })
  await t.test('registry failures are not treated as first publication', async () => {
    await assert.rejects(generatedDataChanged(root, async () => new Response('', {status: 503})), /metadata request failed: HTTP 503/)
    await assert.rejects(generatedDataChanged(root, async () => { throw new Error('Network unavailable') }), /Network unavailable/)
    await assert.rejects(compare(generated, {versions: {}}), /No published versions/)
    await assert.rejects(compare(generated, {...metadata, time: {}}), /Missing publication timestamp/)
  })
  await t.test('tarball errors and missing generated data fail explicitly', async () => {
    await assert.rejects(generatedDataChanged(root, async url =>
      url.endsWith('.tgz') ? new Response('', {status: 404}) : Response.json(metadata),
    ), /tarball request failed.*HTTP 404/)
    const incomplete = execFileSync('tar', ['-czf', '-', '-C', root, 'package.json'])
    await assert.rejects(generatedDataChanged(root, async url =>
      url.endsWith('.tgz') ? new Response(incomplete) : Response.json(metadata),
    ), /Command failed: tar/)
  })
})
