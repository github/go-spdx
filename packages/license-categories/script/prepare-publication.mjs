import {execFileSync} from 'node:child_process'
import {createHash} from 'node:crypto'
import {appendFileSync, mkdtempSync, readFileSync, rmSync, writeFileSync} from 'node:fs'
import {tmpdir} from 'node:os'
import {join, resolve} from 'node:path'
import {fileURLToPath} from 'node:url'
import semver from 'semver'
import {list} from 'tar'

const packageName = '@github/license-categories'
const registry = 'https://registry.npmjs.org'
const maximumBytes = 100 * 1024 * 1024
const maximumEntries = 100
const maximumPathBytes = 1024 * 1024

export function publicationRef(eventName, event, selectedRef) {
  if (eventName === 'release' && event.action === 'published' && event.release?.tag_name) {
    return `refs/tags/${event.release.tag_name}`
  }
  if (eventName === 'workflow_dispatch' && selectedRef) return selectedRef
  throw new Error('Publication requires a published release tag or a workflow_dispatch selected ref.')
}

export async function packedFiles(path) {
  const files = new Map()
  let total = 0
  let entries = 0
  let pathBytes = 0
  let failure
  const invalidate = message => { failure = new Error(message) }
  await list({
    file: path,
    strict: true,
    onReadEntry(entry) {
      if (failure) return
      entries++
      pathBytes += Buffer.byteLength(entry.path)
      if (entries > maximumEntries) return invalidate('Package archive entry count exceeds 100.')
      if (pathBytes > maximumPathBytes) return invalidate('Package archive path bytes exceed 1 MiB.')
      if (entry.type === 'Directory') return
      if (entry.type !== 'File' || !entry.path.startsWith('package/')) {
        return invalidate(`Unsupported package archive entry: ${entry.path} (${entry.type})`)
      }
      const path = entry.path.slice('package/'.length)
      if (!path || path.split('/').some(part => part === '..' || part === '' || part === '.')
          || path.includes('\\') || files.has(path)) {
        return invalidate(`Invalid or duplicate package archive path: ${entry.path}`)
      }
      total += entry.size
      if (total > maximumBytes) return invalidate('Package contents exceed 100 MiB.')
      const chunks = []
      files.set(path, chunks)
      entry.on('data', chunk => {
        if (!failure) chunks.push(chunk)
      })
    },
  }).catch(error => {
    throw failure ?? error
  })
  // tar drains entries after the callback; reject here so callers can clean up.
  if (failure) throw failure
  return new Map([...files].map(([path, chunks]) => [path, Buffer.concat(chunks)]))
}

export function samePackageContents(candidate, published) {
  const normalize = files => {
    const manifestBytes = files.get('package.json')
    if (!manifestBytes) throw new Error('Package is missing package.json; cannot compare package contents.')
    const manifest = JSON.parse(manifestBytes.toString('utf8'))
    if (manifest.name !== packageName) throw new Error('Unexpected packaged manifest name.')
    // npm may add bookkeeping to the registry tarball. Consumer metadata stays significant.
    for (const key of Object.keys(manifest)) {
      if (key === 'version' || key === 'gitHead' || key.startsWith('_')) delete manifest[key]
    }
    // Nested key order can affect exports, imports, and typesVersions resolution.
    const ordered = Object.fromEntries(Object.keys(manifest).sort().map(key => [key, manifest[key]]))
    return new Map([...files, ['package.json', Buffer.from(JSON.stringify(ordered))]])
  }
  const left = normalize(candidate)
  const right = normalize(published)
  return left.size === right.size && [...left].every(([path, bytes]) => right.has(path) && bytes.equals(right.get(path)))
}

export function latestPublication(metadata) {
  if (metadata.name !== packageName || !metadata.versions || !metadata.time) {
    throw new Error('Registry returned invalid package metadata (name, versions, or publication times).')
  }
  const versions = Object.keys(metadata.versions)
  if (!versions.length || versions.some(version => !semver.valid(version))) {
    throw new Error('Registry returned no published versions or an invalid semantic version.')
  }
  for (const version of versions) {
    if (!Number.isFinite(Date.parse(metadata.time[version]))) {
      throw new Error(`Registry omitted the publication time for ${version}.`)
    }
  }
  return versions.sort((a, b) => Date.parse(metadata.time[b]) - Date.parse(metadata.time[a]) || semver.rcompare(a, b))[0]
}

export function validateVersion(manifest, lock, metadata) {
  const version = manifest.version
  if (manifest.name !== packageName || lock.name !== packageName || lock.packages?.['']?.name !== packageName
      || lock.version !== version || lock.packages[''].version !== version) {
    throw new Error('Manifest and lockfile names/versions must agree. Update package.json and package-lock.json together.')
  }
  const parsed = semver.parse(version)
  const normalized = parsed && parsed.version + (parsed.build.length ? `+${parsed.build.join('.')}` : '')
  if (!parsed || normalized !== version) {
    throw new Error(`Invalid npm semantic version ${JSON.stringify(version)}. Set an explicit valid version in both package files.`)
  }
  if (!metadata) return
  if (Object.hasOwn(metadata.versions, version)) {
    throw new Error(`Version ${version} is already published. Bump package.json and package-lock.json; npm versions cannot be reused.`)
  }
  const newest = Object.keys(metadata.versions).sort(semver.rcompare)[0]
  if (!semver.gt(version, newest)) {
    throw new Error(`Version ${version} must be semantically newer than published version ${newest}. Bump both package files; no automatic bump is performed.`)
  }
}

async function responseBytes(response) {
  const chunks = []
  let size = 0
  for await (const chunk of response.body) {
    size += chunk.length
    if (size > maximumBytes) throw new Error('Registry response exceeds 100 MiB.')
    chunks.push(chunk)
  }
  return Buffer.concat(chunks)
}

export async function registryMetadata(fetcher = fetch) {
  const response = await fetcher(`${registry}/${encodeURIComponent(packageName)}`, {
    signal: AbortSignal.timeout(60_000),
    redirect: 'error',
    headers: {accept: 'application/json'},
  })
  const bytes = await responseBytes(response)
  if (response.status === 404) {
    const body = JSON.parse(bytes.toString('utf8'))
    if (body.error === 'Not found') return null
    throw new Error('Registry returned an inconclusive 404; refusing to assume first publication.')
  }
  if (!response.ok) throw new Error(`Registry lookup failed: HTTP ${response.status}. Check registry access before retrying.`)
  const metadata = JSON.parse(bytes.toString('utf8'))
  latestPublication(metadata)
  return metadata
}

export async function publishedTarball(dist, destination, fetcher = fetch) {
  const url = new URL(dist?.tarball)
  if (url.origin !== registry || url.username || url.password) {
    throw new Error('Registry returned an unexpected tarball URL; only HTTPS registry.npmjs.org is allowed.')
  }
  const response = await fetcher(url, {signal: AbortSignal.timeout(60_000), redirect: 'error'})
  if (!response.ok) throw new Error(`Published tarball download failed: HTTP ${response.status}.`)
  const bytes = await responseBytes(response)
  const integrity = dist.integrity
  if (typeof integrity === 'string' && /^sha512-[A-Za-z0-9+/]+=*$/.test(integrity)) {
    if (`sha512-${createHash('sha512').update(bytes).digest('base64')}` !== integrity) {
      throw new Error('Published tarball integrity mismatch.')
    }
  } else if (typeof dist.shasum === 'string' && /^[a-f0-9]{40}$/.test(dist.shasum)) {
    if (createHash('sha1').update(bytes).digest('hex') !== dist.shasum) {
      throw new Error('Published tarball checksum mismatch.')
    }
  } else {
    throw new Error('Registry omitted a supported tarball integrity hash.')
  }
  writeFileSync(destination, bytes)
}

export async function preparePublication(directory, destination, fetcher = fetch) {
  const temporary = mkdtempSync(join(tmpdir(), 'license-categories-publication-'))
  try {
    const manifest = JSON.parse(readFileSync(join(directory, 'package.json'), 'utf8'))
    const lock = JSON.parse(readFileSync(join(directory, 'package-lock.json'), 'utf8'))
    // Compare and publish the same already-built artifact without lifecycle scripts.
    const output = execFileSync('npm', ['pack', '--json', '--ignore-scripts', '--pack-destination', temporary], {
      cwd: directory, encoding: 'utf8',
    })
    const [{filename}] = JSON.parse(output)
    const candidatePath = join(temporary, filename)
    const candidate = await packedFiles(candidatePath)
    const metadata = await registryMetadata(fetcher)
    if (metadata) {
      const latest = latestPublication(metadata)
      const publishedPath = join(temporary, 'published.tgz')
      await publishedTarball(metadata.versions[latest].dist, publishedPath, fetcher)
      if (samePackageContents(candidate, await packedFiles(publishedPath))) {
        return {changed: false, message: `Skipping publication: package contents are unchanged from ${packageName}@${latest}.`}
      }
    }
    validateVersion(manifest, lock, metadata)
    writeFileSync(destination, readFileSync(candidatePath))
    const reason = metadata ? 'Package contents changed' : 'First publication'
    return {changed: true, message: `${reason}; ${packageName}@${manifest.version} is eligible for publication.`}
  } finally {
    rmSync(temporary, {recursive: true, force: true})
  }
}

async function main() {
  const destination = process.argv[2]
  if (!destination) throw new Error('Usage: node script/prepare-publication.mjs /absolute/path/candidate.tgz')
  if (process.env.GITHUB_ACTIONS === 'true') {
    const event = JSON.parse(readFileSync(process.env.GITHUB_EVENT_PATH, 'utf8'))
    const ref = publicationRef(process.env.GITHUB_EVENT_NAME, event, process.env.GITHUB_REF)
    const expected = execFileSync('git', ['rev-parse', '--verify', '--end-of-options', `${ref}^{commit}`], {encoding: 'utf8'}).trim()
    const actual = execFileSync('git', ['rev-parse', 'HEAD'], {encoding: 'utf8'}).trim()
    if (expected !== actual) throw new Error('Checked-out commit does not match the release tag or selected dispatch ref.')
  }
  const result = await preparePublication(fileURLToPath(new URL('../', import.meta.url)), resolve(destination))
  console.log(result.message)
  if (process.env.GITHUB_OUTPUT) appendFileSync(process.env.GITHUB_OUTPUT, `changed=${result.changed}\n`)
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main().catch(error => {
    console.error('Publication preparation failed:', error)
    process.exitCode = 1
  })
}
