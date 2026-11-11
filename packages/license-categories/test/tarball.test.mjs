import assert from 'node:assert/strict'
import {execFileSync} from 'node:child_process'
import {readFileSync} from 'node:fs'
import test from 'node:test'

test('targets public npm distribution under MIT', () => {
  const manifest = JSON.parse(readFileSync(new URL('../package.json', import.meta.url), 'utf8'))
  const lock = JSON.parse(readFileSync(new URL('../package-lock.json', import.meta.url), 'utf8'))

  assert.equal(manifest.name, '@github/license-categories')
  assert.deepEqual(manifest.publishConfig, {access: 'public', registry: 'https://registry.npmjs.org'})
  assert.equal(manifest.license, 'MIT')
  assert.equal(lock.name, manifest.name)
  assert.equal(lock.packages[''].name, manifest.name)
  assert.equal(lock.version, manifest.version)
  assert.equal(lock.packages[''].version, manifest.version)
})

test('publishes only compiled JavaScript, declarations, and package metadata', () => {
  const output = execFileSync('npm', ['pack', '--dry-run', '--json'], {encoding: 'utf8'})
  const [{files}] = JSON.parse(output)

  assert.deepEqual(
    files.map(({path}) => path).sort(),
    ['LICENSE', 'README.md', 'THIRD_PARTY_NOTICE.md', 'dist/generated.d.ts', 'dist/generated.js', 'dist/index.d.ts', 'dist/index.js', 'package.json'],
  )
})