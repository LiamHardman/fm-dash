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
