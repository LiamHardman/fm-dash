## [1.1.4](https://git.liamhardman.com/liam/fm24-golang/compare/v1.1.3...v1.1.4) (2025-06-02)

### 🐛 Bug Fixes

* even less vertical scrolling ([943c291](https://git.liamhardman.com/liam/fm24-golang/commit/943c2918721b67a4148f464f82bb7698c34d68f7))

## [1.1.3](https://git.liamhardman.com/liam/fm24-golang/compare/v1.1.2...v1.1.3) (2025-06-02)

### 🐛 Bug Fixes

* reducing vertical scrolling on detail dialog ([bdac16a](https://git.liamhardman.com/liam/fm24-golang/commit/bdac16a65242fdf5aa5e06032b6b61a6ecc971f2))

## [1.1.2](https://git.liamhardman.com/liam/fm24-golang/compare/v1.1.1...v1.1.2) (2025-06-02)

### 🐛 Bug Fixes

* salary/value colour fix ([d540757](https://git.liamhardman.com/liam/fm24-golang/commit/d5407579124d4e25e07d24a5553b930d0a9a8e1e))

## [1.1.1](https://git.liamhardman.com/liam/fm24-golang/compare/v1.1.0...v1.1.1) (2025-06-02)

### 🐛 Bug Fixes

* styling for bargainhunter ([752ab66](https://git.liamhardman.com/liam/fm24-golang/commit/752ab66995ffd67e2ace1145c2052073e71f4e7c))

## [1.1.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.0.2...v1.1.0) (2025-06-02)

### 🚀 Features

* removed echart and added sort options in the playerdatatable ([6ff243a](https://git.liamhardman.com/liam/fm24-golang/commit/6ff243a7e65ec7350c70da08c9b6d84fae981811))

## [1.0.2](https://git.liamhardman.com/liam/fm24-golang/compare/v1.0.1...v1.0.2) (2025-06-02)

### 🐛 Bug Fixes

* removing graphql confirm ([a90f055](https://git.liamhardman.com/liam/fm24-golang/commit/a90f05570740a9fdb7f5ea987c0609eb403ff621))

## [1.0.1](https://git.liamhardman.com/liam/fm24-golang/compare/v1.0.0...v1.0.1) (2025-06-02)

### 🐛 Bug Fixes

* removed assets folder from sem release ([8cd20c6](https://git.liamhardman.com/liam/fm24-golang/commit/8cd20c6c67c61873d1b5840d9103ac7a4612bcce))

## 1.0.0 (2025-06-02)

### 🚀 Features

* add pre-commit hooks with Husky and lint-staged ([f3c3177](https://git.liamhardman.com/liam/fm24-golang/commit/f3c317795d81032c1f9d5a56d41d69af67cdfbb9))

### 🐛 Bug Fixes

* add timeout protection to Go tests in pre-push hook - prevents hanging on TestAsyncProcessingCorrectness test with 30s timeout ([8fe9a7d](https://git.liamhardman.com/liam/fm24-golang/commit/8fe9a7d066b95ba68806cf1c24faaf718ad746c2))
* attempted fix wiht semantic release ([34bd280](https://git.liamhardman.com/liam/fm24-golang/commit/34bd28087fdfc154d8b222c590a402ebc8ba0f03))
* disable husky in semantic ([bde0895](https://git.liamhardman.com/liam/fm24-golang/commit/bde0895db057149e53c06b467b23989ce036faf4))
* resolve async processing test hanging - fix double worker initialization, channel race conditions, and add test timeout protection ([554b530](https://git.liamhardman.com/liam/fm24-golang/commit/554b5302d1b8c923fdd67fcb4934bd411d662c30))
* resolve pre-commit hook issues - increase Biome maxSize to 3MB, exclude teams_data.json from formatting, remove unused Go function, add optional frontend tests ([646c181](https://git.liamhardman.com/liam/fm24-golang/commit/646c1817ccd1b289ff4cdce82d203fb242c67953))
* testing new release schedule thingy ([4d23469](https://git.liamhardman.com/liam/fm24-golang/commit/4d2346935ff2b3153290e8756313ec10734f9439))
