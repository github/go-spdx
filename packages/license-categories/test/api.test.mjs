import assert from 'node:assert/strict'
import test from 'node:test'

import * as licenseCategories from '../dist/index.js'

test('exports the existing license category API', () => {
  assert.deepEqual(Object.keys(licenseCategories).sort(), [
    'categoriesWithLicenses',
    'categoryDescriptions',
    'categoryForLicense',
    'licensesById',
  ])
})

test('looks up and groups generated license data', () => {
  assert.equal(licenseCategories.categoryForLicense('MIT'), 'Permissive')
  assert.equal(licenseCategories.categoryForLicense('not-an-spdx-id'), undefined)

  const permissive = licenseCategories.categoriesWithLicenses().find(({category}) => category === 'Permissive')
  assert.ok(permissive)
  assert.equal(permissive.description, 'Safe for most')
  assert.ok(permissive.licenses.some(({id, name}) => id === 'MIT' && name === 'MIT License'))
})

test('groups every license once with the existing descriptions and order', () => {
  const groups = licenseCategories.categoriesWithLicenses()
  assert.deepEqual(groups.map(({category}) => category), groups.map(({category}) => category).sort((a, z) => a.localeCompare(z)))

  const ids = []
  for (const {category, description, licenses} of groups) {
    assert.equal(description, licenseCategories.categoryDescriptions[category])
    assert.deepEqual(licenses, [...licenses].sort((a, z) => a.name.localeCompare(z.name) || a.id.localeCompare(z.id)))
    for (const {id, name} of licenses) {
      assert.deepEqual(licenseCategories.licensesById[id], {name, category})
      assert.equal(licenseCategories.categoryForLicense(id), category)
      ids.push(id)
    }
  }
  assert.deepEqual(ids.sort(), Object.keys(licenseCategories.licensesById).sort())
})