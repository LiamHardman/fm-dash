'use strict'

const { execSync } = require('node:child_process')
const path = require('node:path')

const apiDir = path.resolve(__dirname, '..', 'src', 'api')

try {
  execSync('golangci-lint run', { cwd: apiDir, stdio: 'inherit' })
} catch (err) {
  process.exit(err.status ?? 1)
}
