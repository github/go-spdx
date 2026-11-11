import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import test from 'node:test'

const workflow = readFileSync(new URL('../../../.github/workflows/publish-license-categories.yaml', import.meta.url), 'utf8')

test('publication verifies, stamps, builds/tests, prepares, and publishes the compared artifact in order', () => {
  const commands = [
    'run: make verify-license-categories',
    'run: npm ci',
    'run: node script/stamp-provenance.mjs',
    'run: npm test',
    'run: node script/prepare-publication.mjs "$RUNNER_TEMP/license-categories.tgz"',
    'run: npm publish "$RUNNER_TEMP/license-categories.tgz" --ignore-scripts --access public --provenance --registry=https://registry.npmjs.org',
  ]
  let previous = -1
  for (const command of commands) {
    const index = workflow.indexOf(command)
    assert.ok(index > previous, `${command} is missing or out of order`)
    previous = index
  }
  assert.match(workflow, /id: prepare/)
  assert.match(workflow, /if: steps\.prepare\.outputs\.changed == 'true'/)
  assert.doesNotMatch(workflow, /run: npm run test:tarball/)
  assert.doesNotMatch(workflow, /NPM_CONFIG_PROVENANCE: 'false'/)
})

test('publication retains release/dispatch gates, exact refs, serialization and OIDC safeguards', () => {
  assert.match(workflow, /release:\s+types: \[released\]/)
  assert.match(workflow, /workflow_dispatch:/)
  assert.doesNotMatch(workflow, /^\s+(push|pull_request):|^run-name:/m)
  assert.match(workflow, /LICENSE_CATEGORIES_NPM_PUBLISH_ENABLED == 'true'/)
  assert.match(workflow, /github\.event_name != 'release' \|\| github\.event\.release\.prerelease == false/)
  assert.doesNotMatch(workflow, /default_branch|Validate default branch/)
  assert.match(workflow, /group: publish-license-categories\s+cancel-in-progress: false/)
  assert.match(workflow, /ref: \$\{\{ github\.event_name == 'release' && format\('refs\/tags\/\{0\}', github\.event\.release\.tag_name\) \|\| github\.ref \}\}/)
  assert.match(workflow, /persist-credentials: false/)
  assert.match(workflow, /^permissions:\n  contents: read$/m)
  assert.match(workflow, /    permissions:\n      contents: read\n      id-token: write/)
  assert.doesNotMatch(workflow, /contents: write|NPM_TOKEN|git push|git commit/)
  assert.match(workflow, /node-version: 24/)
  assert.match(workflow, /npm install -g npm@11\.20\.0/)
  assert.match(workflow, /cache: npm\s+cache-dependency-path: packages\/license-categories\/package-lock\.json/)
  assert.ok([...workflow.matchAll(/uses: [^\n]+/g)].every(([action]) => /@[a-f0-9]{40}\b/.test(action)))
})
