/** @type {import('@rtk-query/codegen-openapi').ConfigFile} */
const config = {
  schemaFile: '../backend/api/openapi.yml',
  apiFile: './src/store/api.ts',
  apiImport: 'api',
  outputFile: './src/api/client.ts',
  hooks: true,
  tag: true,
}

module.exports = config
