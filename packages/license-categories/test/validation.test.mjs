import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import test from 'node:test'

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
