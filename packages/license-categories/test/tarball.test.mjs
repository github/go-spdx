import assert from 'node:assert/strict'
import {execFileSync} from 'node:child_process'
import {cpSync, mkdirSync, mkdtempSync, readFileSync, rmSync, symlinkSync, writeFileSync} from 'node:fs'
import {tmpdir} from 'node:os'
import {join} from 'node:path'
import {fileURLToPath} from 'node:url'
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

for (const event of ['release', 'workflow_dispatch']) {
  test(`preserves exact ${event} provenance when packing`, () => {
    const root = mkdtempSync(join(tmpdir(), 'license-categories-provenance-'))
    try {
      const packageDir = join(root, 'packages/license-categories')
      mkdirSync(packageDir, {recursive: true})
      for (const path of ['src', 'package.json', 'tsconfig.json']) {
        cpSync(new URL(`../${path}`, import.meta.url), join(packageDir, path), {recursive: true})
      }
      symlinkSync(fileURLToPath(new URL('../node_modules', import.meta.url)), join(packageDir, 'node_modules'), 'dir')
      execFileSync('git', ['init', '--quiet', root])
      execFileSync('git', ['-c', 'user.name=Test', '-c', 'user.email=test@example.com', '-c', 'commit.gpgsign=false', 'commit', '--quiet', '--allow-empty', '-m', 'Fixture'], {cwd: root})
      const sha = execFileSync('git', ['rev-parse', 'HEAD'], {cwd: root, encoding: 'utf8'}).trim()
      assert.match(sha, /^[0-9a-f]{40}$/)

      const workflow = readFileSync(new URL('../../../.github/workflows/publish-license-categories.yaml', import.meta.url), 'utf8')
      assert.ok(workflow.indexOf('name: Verify generated license categories') < workflow.indexOf('name: Stamp go-spdx provenance'))
      assert.ok(workflow.indexOf('name: Stamp go-spdx provenance') < workflow.indexOf('name: Test package'))
      const script = workflow.split('      - name: Stamp go-spdx provenance\n')[1]
        .split('        run: |\n')[1].split('\n      - name:')[0].replace(/^          /gm, '')
      const generated = join(packageDir, 'src/generated.ts')
      const before = readFileSync(generated, 'utf8')
      execFileSync('bash', ['-e', '-c', script], {
        cwd: root,
        env: {...process.env, EVENT_NAME: event, RELEASE_TAG: 'v2.0.1', GITHUB_SHA: '0'.repeat(40)},
      })
      const header = `// go-spdx: ${event === 'release' ? 'v2.0.1' : `${sha} (dispatch)`}`
      const expected = before.replace(/^\/\/ go-spdx: .+$/m, header)
      assert.equal(readFileSync(generated, 'utf8'), expected)

      const [{filename, files}] = JSON.parse(execFileSync('npm', ['pack', '--json'], {cwd: packageDir, encoding: 'utf8'}))
      assert.ok(files.some(({path}) => path === 'dist/generated.js'))
      const packed = execFileSync('tar', ['-xOf', join(packageDir, filename), 'package/dist/generated.js'], {encoding: 'utf8'})
      assert.equal(packed.split('\n')[1], header)
    } finally {
      rmSync(root, {recursive: true, force: true})
    }
  })
}

test('rejects release tags containing JavaScript line terminators before writing', () => {
  const root = mkdtempSync(join(tmpdir(), 'license-categories-release-tag-'))
  try {
    mkdirSync(join(root, 'packages/license-categories/src'), {recursive: true})
    const generated = join(root, 'packages/license-categories/src/generated.ts')
    const source = readFileSync(new URL('../src/generated.ts', import.meta.url), 'utf8')
    writeFileSync(generated, source)
    const workflow = readFileSync(new URL('../../../.github/workflows/publish-license-categories.yaml', import.meta.url), 'utf8')
    const script = workflow.split('      - name: Stamp go-spdx provenance\n')[1]
      .split('        run: |\n')[1].split('\n      - name:')[0].replace(/^          /gm, '')

    for (const separator of ['\n', '\r', '\u2028', '\u2029']) {
      assert.throws(() => execFileSync('bash', ['-e', '-c', script], {
        cwd: root,
        env: {...process.env, EVENT_NAME: 'release', RELEASE_TAG: `v2.0.1${separator}throw new Error('injected')`},
        stdio: 'pipe',
      }), error => error.status !== 0 && /provenance must not contain line terminators/.test(error.stderr.toString()))
      assert.equal(readFileSync(generated, 'utf8'), source)
    }
  } finally {
    rmSync(root, {recursive: true, force: true})
  }
})

test('generation verification ignores only go-spdx provenance changes', () => {
  const root = mkdtempSync(join(tmpdir(), 'license-categories-verify-'))
  try {
    mkdirSync(join(root, 'packages/license-categories/src'), {recursive: true})
    const source = readFileSync(new URL('../src/generated.ts', import.meta.url), 'utf8')
    const generated = join(root, 'packages/license-categories/src/generated.ts')
    const tmp = join(root, 'regenerated.ts')
    writeFileSync(tmp, source)
    const makefile = readFileSync(new URL('../../../Makefile', import.meta.url), 'utf8')
    const recipe = makefile.split('verify-license-categories:\n')[1].split('\n\n')[0]
    const command = recipe.slice(recipe.indexOf('\tnode -e'))
    const verify = () => {
      try {
        return execFileSync('make', ['--no-print-directory', '-f', '-', 'verify'], {
          cwd: root, env: {...process.env, tmp}, input: `verify:\n${command}\n`, stdio: 'pipe',
        })
      } finally {
        rmSync(`${tmp}.checked`, {force: true})
      }
    }

    writeFileSync(generated, source.replace(/^\/\/ go-spdx: .+$/m, '// go-spdx: v2.0.1'))
    verify()
    const expanded = execFileSync('make', ['--no-print-directory', '-n', '-f', '-', 'verify'], {
      cwd: root, env: {...process.env, tmp}, input: `verify:\n${command}\n`, encoding: 'utf8',
    })
    try {
      execFileSync('bash', ['-e', '-c', expanded], {cwd: root, env: {...process.env, tmp}, stdio: 'pipe'})
    } finally {
      rmSync(`${tmp}.checked`, {force: true})
    }
    for (const [from, to] of [
      ['// spdx-license-list-version:', '// changed-spdx-license-list-version:'],
      ['// scancode-licensedb commit:', '// changed-scancode-licensedb commit:'],
      ["'Safe for most'", "'Changed category'"],
    ]) {
      assert.ok(source.includes(from))
      writeFileSync(generated, source.replace(from, to))
      assert.throws(verify, {status: 2})
    }
    for (const invalid of [
      `${source}// go-spdx: extra\n`,
      ...['\r', '\u2028', '\u2029'].map(separator => `${source}// go-spdx: extra${separator}throw new Error('injected')\n`),
      ...['\r', '\u2028', '\u2029'].map(separator => source.replace(/^\/\/ go-spdx: [^\n]*$/m, `// go-spdx: v2.0.1${separator}throw new Error('injected')`)),
      source.replace(/^\/\/ go-spdx: .+\n/m, ''),
    ]) {
      writeFileSync(generated, invalid)
      assert.throws(verify, error => error.status !== 0 && /expected exactly one go-spdx provenance header/.test(error.stderr.toString()))
      assert.equal(readFileSync(generated, 'utf8'), invalid)
    }
    writeFileSync(generated, source)
    const target = join(root, 'symlink-target')
    writeFileSync(target, 'must not be overwritten')
    symlinkSync(target, `${tmp}.checked`)
    assert.throws(verify, error => error.status !== 0 && /EEXIST/.test(error.stderr.toString()))
    assert.equal(readFileSync(target, 'utf8'), 'must not be overwritten')
  } finally {
    rmSync(root, {recursive: true, force: true})
  }
})