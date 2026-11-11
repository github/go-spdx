import {categoryDescriptions, licensesById} from './generated.js'

export {categoryDescriptions, licensesById}

export function categoryForLicense(spdxId: string): string | undefined {
  return licensesById[spdxId]?.category
}

export function categoriesWithLicenses(): Array<{
  category: string
  description: string
  licenses: Array<{id: string; name: string}>
}> {
  const grouped: Record<string, Array<{id: string; name: string}>> = {}
  for (const [id, info] of Object.entries(licensesById)) {
    ;(grouped[info.category] ??= []).push({id, name: info.name})
  }
  return Object.keys(grouped)
    .sort((a, z) => a.localeCompare(z))
    .map(category => ({
      category,
      description: categoryDescriptions[category] ?? '',
      licenses: (grouped[category] ?? []).sort((a, z) => a.name.localeCompare(z.name) || a.id.localeCompare(z.id)),
    }))
}