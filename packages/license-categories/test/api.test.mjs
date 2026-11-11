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