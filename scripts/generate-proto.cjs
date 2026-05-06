/**
 * generate-proto.cjs
 *
 * Regenerates src/utils/proto-descriptor.js from the canonical .proto sources:
 *   src/api/proto/player.proto
 *   src/api/proto/api_responses.proto
 *
 * Usage:
 *   node scripts/generate-proto.cjs
 *   -- or --
 *   npm run generate:proto
 *
 * Requires protobufjs-cli (pbjs) to be installed:
 *   npm install  (it is in devDependencies)
 */

'use strict'

const { execSync } = require('node:child_process')
const fs = require('node:fs')
const path = require('node:path')

const root = path.resolve(__dirname, '..')
const protoFiles = [
  'src/api/proto/player.proto',
  'src/api/proto/api_responses.proto',
]
  .map((f) => path.join(root, f))
  .join(' ')
const protoIncludePath = path.join(root, 'src', 'api')

const outFile = path.join(root, 'src', 'utils', 'proto-descriptor.js')

console.log('Generating proto descriptor from:')
protoFiles.split(' ').forEach((f) => console.log(' ', f))

let jsonOutput
try {
  jsonOutput = execSync(`npx pbjs -t json --path "${protoIncludePath}" ${protoFiles}`, {
    cwd: root,
  }).toString()
} catch (err) {
  console.error('pbjs failed:', err.message)
  process.exit(1)
}

const descriptor = JSON.parse(jsonOutput)

const fileContent = `/**
 * Protobuf JSON descriptor for the fmdash API.
 *
 * AUTO-GENERATED — do not edit by hand.
 * Regenerate with: npm run generate:proto
 *
 * Sources:
 *   src/api/proto/player.proto
 *   src/api/proto/api_responses.proto
 */

const descriptor = ${JSON.stringify(descriptor, null, 2)}

export default descriptor
`

fs.writeFileSync(outFile, fileContent, 'utf8')
console.log(`Written: ${path.relative(root, outFile)}`)
