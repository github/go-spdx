import {execFileSync} from 'node:child_process'
import {readFileSync, writeFileSync} from 'node:fs'
import {resolve} from 'node:path'
import {fileURLToPath} from 'node:url'

function gitRevision(ref) {
  return execFileSync('git', ['rev-parse', '--verify', '--end-of-options', `${ref}^{commit}`], {
    encoding: 'utf8',
  }).trim()
}

export function publicationProvenance(eventName, event, revision = gitRevision) {
  const head = revision('HEAD')
  if (!/^[a-f0-9]{40}$/.test(head)) throw new Error('Expected the full checked-out Git commit SHA.')
  if (eventName === 'workflow_dispatch') return `${head} (dispatch)`
  if (eventName === 'release' && event.action === 'published') {
    const tag = event.release?.tag_name
    if (typeof tag !== 'string' || !tag || /[\r\n]/.test(tag)) {
      throw new Error('Published release must identify a single-line Git tag.')
    }
    if (revision(`refs/tags/${tag}`) !== head) {
      throw new Error('Checked-out commit does not match the published release tag.')
    }
    return `${tag} (release)`
  }
  throw new Error('Provenance stamping requires a published release or manual dispatch.')
}

export function stampSource(source, provenance) {
  const header = /^\/\/ go-spdx: [^\r\n]*$/gm
  if ([...source.matchAll(header)].length !== 1) {
    throw new Error('Expected exactly one go-spdx provenance header in generated source.')
  }
  return source.replace(header, () => `// go-spdx: ${provenance}`)
}

function main() {
  if (process.env.GITHUB_ACTIONS !== 'true') throw new Error('Publication stamping is only supported in GitHub Actions.')
  const event = JSON.parse(readFileSync(process.env.GITHUB_EVENT_PATH, 'utf8'))
  const provenance = publicationProvenance(process.env.GITHUB_EVENT_NAME, event)
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
