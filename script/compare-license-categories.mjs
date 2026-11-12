import {execFileSync} from 'node:child_process'
import {readFileSync} from 'node:fs'
import {join} from 'node:path'
import {fileURLToPath} from 'node:url'

function withoutProvenance(source) {
  return source.replace(/^(?:\/\/[^\n]*\n)+/, header =>
    header.replace(/^\/\/ (?:go-spdx|spdx-license-list-version|scancode-licensedb commit):[^\n]*\n/gm, ''),
  )
}

export async function generatedDataChanged(
  packageDir = fileURLToPath(new URL('../packages/license-categories/', import.meta.url)),
  registryFetch = fetch,
) {
  const {name} = JSON.parse(readFileSync(join(packageDir, 'package.json'), 'utf8'))
  const local = readFileSync(join(packageDir, 'dist/generated.js'), 'utf8')
  const response = await registryFetch(`https://registry.npmjs.org/${encodeURIComponent(name)}`)
  if (response.status === 404) {
    console.log(`${name} has not been published; publishing initial generated data.`)
    return true
  }
  if (!response.ok) throw new Error(`npm metadata request failed: HTTP ${response.status}`)
  const metadata = await response.json()
  const versions = Object.keys(metadata.versions ?? {}).map(version => {
    const time = Date.parse(metadata.time?.[version])
    if (!Number.isFinite(time)) throw new Error(`Missing publication timestamp for ${name}@${version}`)
    return {version, time}
  })
  versions.sort((a, b) => b.time - a.time)
  if (!versions.length) throw new Error(`No published versions found for ${name}`)
  const {version} = versions[0]
  const tarball = await registryFetch(metadata.versions[version].dist.tarball)
  if (!tarball.ok) throw new Error(`npm tarball request failed for ${name}@${version}: HTTP ${tarball.status}`)
  const published = execFileSync('tar', ['-xzOf', '-', 'package/dist/generated.js'], {
    input: Buffer.from(await tarball.arrayBuffer()),
    encoding: 'utf8',
  })
  const changed = withoutProvenance(local) !== withoutProvenance(published)
  console.log(`Generated data ${changed ? 'changed; publishing' : 'unchanged; skipping publication'} compared with ${name}@${version}.`)
  return changed
}
