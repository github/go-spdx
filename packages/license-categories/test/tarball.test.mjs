import assert from 'node:assert/strict'
import {execFileSync} from 'node:child_process'
import {mkdtempSync, readFileSync, rmSync} from 'node:fs'
import {tmpdir} from 'node:os'
import {join} from 'node:path'
import test from 'node:test'
import {pathToFileURL} from 'node:url'

test('targets public npm distribution under MIT', () => {
  const manifest = JSON.parse(readFileSync(new URL('../package.json', import.meta.url), 'utf8'))
  const lock = JSON.parse(readFileSync(new URL('../package-lock.json', import.meta.url), 'utf8'))

  assert.equal(manifest.name, '@github/license-categories')
  assert.deepEqual(manifest.publishConfig, {access: 'public', registry: 'https://registry.npmjs.org'})
  assert.equal(manifest.license, 'MIT')
  assert.deepEqual(manifest.repository, {
    type: 'git',
    url: 'git+https://github.com/github/go-spdx.git',
    directory: 'packages/license-categories',
  })
  assert.equal(lock.name, manifest.name)
  assert.equal(lock.packages[''].name, manifest.name)
  assert.equal(lock.version, manifest.version)
  assert.equal(lock.packages[''].version, manifest.version)
  assert.equal(lock.packages[''].license, manifest.license)
  assert.equal(
    readFileSync(new URL('../LICENSE', import.meta.url), 'utf8'),
    readFileSync(new URL('../../../LICENSE', import.meta.url), 'utf8'),
  )
})

test('publishes only compiled JavaScript, declarations, and package metadata', () => {
  const output = execFileSync('npm', ['pack', '--dry-run', '--json', '--ignore-scripts'], {encoding: 'utf8'})
  const [{files}] = JSON.parse(output)

  assert.deepEqual(
    files.map(({path}) => path).sort(),
    ['LICENSE', 'README.md', 'THIRD_PARTY_NOTICE.md', 'dist/generated.d.ts', 'dist/generated.js', 'dist/index.d.ts', 'dist/index.js', 'package.json'],
  )
})

test('packed package works independently of the source tree', async () => {
  const directory = mkdtempSync(join(tmpdir(), 'license-categories-'))
  try {
    const output = execFileSync('npm', ['pack', '--json', '--ignore-scripts', '--pack-destination', directory], {encoding: 'utf8'})
    const [{filename}] = JSON.parse(output)
    execFileSync('tar', ['-xzf', join(directory, filename), '-C', directory])
    const manifest = JSON.parse(readFileSync(join(directory, 'package/package.json'), 'utf8'))
    assert.equal(manifest.name, '@github/license-categories')
    const packed = await import(pathToFileURL(join(directory, 'package/dist/index.js')).href)
    assert.equal(packed.categoryForLicense('MIT'), 'Permissive')
    assert.equal(packed.categoryForLicense('not-an-spdx-id'), undefined)
    assert.ok(packed.categoriesWithLicenses().length > 0)
    const notice = readFileSync(join(directory, 'package/THIRD_PARTY_NOTICE.md'), 'utf8')
    assert.ok(notice.includes('CC-BY-4.0'))
    assert.ok(notice.includes('Copyright (c) nexB Inc. and others.'))
    assert.ok(notice.includes('https://creativecommons.org/licenses/by/4.0/'))
    assert.ok(notice.includes('https://github.com/aboutcode-org/scancode-licensedb/blob/main/NOTICE'))
    assert.ok(notice.includes('dist/generated.js'))
    const generated = readFileSync(join(directory, 'package/dist/generated.js'), 'utf8')
    assert.match(generated, /^\/\/ scancode-licensedb commit: [a-f0-9]{40}$/m)
    assert.ok(notice.includes("These modifications do not change\nthe upstream data's license."))
    assert.ok(notice.includes('Some licenses, such as the GNU GENERAL PUBLIC LICENSE, are subject to other licenses.'))
    assert.ok(notice.includes('Consult an attorney for any legal advice.'))
  } finally {
    rmSync(directory, {recursive: true, force: true})
  }
})