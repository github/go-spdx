import {execFileSync} from 'node:child_process'
import {readFileSync, writeFileSync} from 'node:fs'
import {resolve} from 'node:path'
import {fileURLToPath} from 'node:url'

const commentTerminator = /[\r\n\u2028\u2029]/

function gitRevision(ref) {
  return execFileSync('git', ['rev-parse', '--verify', '--end-of-options', `${ref}^{commit}`], {
    encoding: 'utf8',
  }).trim()
}

export function publicationProvenance(eventName, event, eventSHA, revision = gitRevision) {
  if (typeof eventSHA !== 'string' || !/^[a-f0-9]{40}$/.test(eventSHA)) {
    throw new Error('Expected GITHUB_SHA to identify the full event commit SHA.')
  }
  const head = revision('HEAD')
  if (!/^[a-f0-9]{40}$/.test(head)) throw new Error('Expected the full checked-out Git commit SHA.')
  if (head !== eventSHA) {
    throw new Error('Checked-out commit does not match GITHUB_SHA; refusing to publish mismatched npm provenance.')
  }
  if (eventName === 'workflow_dispatch') return `${head} (dispatch)`
  if (eventName === 'release' && event.action === 'released') {
    const tag = event.release?.tag_name
    if (typeof tag !== 'string' || !tag || commentTerminator.test(tag)) {
      throw new Error('Released event must identify a single-line Git tag.')
    }
    if (revision(`refs/tags/${tag}`) !== head) {
      throw new Error('Checked-out commit does not match the released tag.')
    }
    return `${tag} (release)`
  }
  throw new Error('Provenance stamping requires a released event or manual dispatch.')
}

export function stampSource(source, provenance) {
  if (commentTerminator.test(provenance)) {
    throw new Error('Publication provenance must be a single-line TypeScript comment value.')
  }
  const header = /^\/\/ go-spdx: [^\r\n]*$/gm
  if ([...source.matchAll(header)].length !== 1) {
    throw new Error('Expected exactly one go-spdx provenance header in generated source.')
  }
  return source.replace(header, () => `// go-spdx: ${provenance}`)
}

function main() {
  if (process.env.GITHUB_ACTIONS !== 'true') throw new Error('Publication stamping is only supported in GitHub Actions.')
  const event = JSON.parse(readFileSync(process.env.GITHUB_EVENT_PATH, 'utf8'))
  const provenance = publicationProvenance(process.env.GITHUB_EVENT_NAME, event, process.env.GITHUB_SHA)
  const path = new URL('../src/generated.ts', import.meta.url)
  const source = readFileSync(path, 'utf8')
  writeFileSync(path, stampSource(source, provenance))
  console.log(`Stamped publication provenance: ${provenance}`)
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    main()
  } catch (error) {
    console.error('Publication provenance stamping failed:', error)
    process.exitCode = 1
  }
}
