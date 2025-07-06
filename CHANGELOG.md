## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-07-04)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* more filtering on the performance page ([4f19675](https://git.liamhardman.com/liam/fm24-golang/commit/4f19675dad324d9d3ffe96c4f83e5b8b25db6646))
* moved config to its own config file ([dc65025](https://git.liamhardman.com/liam/fm24-golang/commit/dc6502590c15678b9616adbf689d61d1a235374e))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))
* team ID to name is now running on backend ([e139c3a](https://git.liamhardman.com/liam/fm24-golang/commit/e139c3ada40cc0fd346c1f140beb96e68608a720))
* wonderkid filter improvements ([b657d7e](https://git.liamhardman.com/liam/fm24-golang/commit/b657d7e72c3916a3a94b561661b4261467ba5f7d))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted kubernetes multipel replica crash fix ([dff5633](https://git.liamhardman.com/liam/fm24-golang/commit/dff5633ec450fd0e2c53d10b19d9660c280eefa8))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* code linting issue fixes ([b9de2d5](https://git.liamhardman.com/liam/fm24-golang/commit/b9de2d5dea1a34304d8bf6e3d8db9976af77000b))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* fixed percentile values being 0 or - counting ([d30d46e](https://git.liamhardman.com/liam/fm24-golang/commit/d30d46e5ad691eb87591661e1e90d15f824051f8))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented runtime assessment of otel usage ([6d84928](https://git.liamhardman.com/liam/fm24-golang/commit/6d849285cec0bd25ef281d4bfca6e07e0aba973a))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* more backend optimization ([c7b8f17](https://git.liamhardman.com/liam/fm24-golang/commit/c7b8f178624e05c3dee68a35fbe44315540e7016))
* more code fixes from precommit ([0ba5133](https://git.liamhardman.com/liam/fm24-golang/commit/0ba513369a51ed330eb0b1f5f1f3118aa59454a8))
* new demo dataset ID ([61e5613](https://git.liamhardman.com/liam/fm24-golang/commit/61e5613b88bb275ee471476e04e36c4a879a3120))
* only logging non-200 requests ([24cc1d9](https://git.liamhardman.com/liam/fm24-golang/commit/24cc1d96c53d44b0220b8253ffefb92cd1665b4a))
* only logging non-200 requests ([6d50bde](https://git.liamhardman.com/liam/fm24-golang/commit/6d50bde4b3f17ef80556f413199ff01701092d2d))
* optimizations ([3051728](https://git.liamhardman.com/liam/fm24-golang/commit/3051728e3746add5aee8b4d65f57728ab864d755))
* optimizations ([b30c77a](https://git.liamhardman.com/liam/fm24-golang/commit/b30c77a84a58617696ad4329cf35f21fa1c262d7))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog in the bargainhunterdialog ([d760c1e](https://git.liamhardman.com/liam/fm24-golang/commit/d760c1effd69e3fe8472dd1da1e2bc1be1ac682c))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* some changes to team handlers which i can't remember, and readme change ([407b862](https://git.liamhardman.com/liam/fm24-golang/commit/407b862cac6072dd6b83e6d3ba86d4209748f70d))
* some fixes for before github release ([2f54fe9](https://git.liamhardman.com/liam/fm24-golang/commit/2f54fe90ff0f3c42b5055ec0c38a02c8b9124aa4))
* some linting fixes again ([96f4d02](https://git.liamhardman.com/liam/fm24-golang/commit/96f4d02197aa981c7dd8e33bd209f03e773c8806))
* team logo util fix ([7041a75](https://git.liamhardman.com/liam/fm24-golang/commit/7041a758ccdf2f54ec14f04c07da1f0151f7f005))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* changed some comments and removed development.md ([841ff03](https://git.liamhardman.com/liam/fm24-golang/commit/841ff03e33f66b66d8a01f9f031ff0df2c806f54))
* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))
* removed comments that weren't necessary ([26d4ed4](https://git.liamhardman.com/liam/fm24-golang/commit/26d4ed45cbc6c66bcd9f7e5d39570cff9ca1cfb8))
* slight tweak to roadmap ([036c0e1](https://git.liamhardman.com/liam/fm24-golang/commit/036c0e12f73d4634cb5d9b8b741554c3054c1849))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-07-01)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* more filtering on the performance page ([4f19675](https://git.liamhardman.com/liam/fm24-golang/commit/4f19675dad324d9d3ffe96c4f83e5b8b25db6646))
* moved config to its own config file ([dc65025](https://git.liamhardman.com/liam/fm24-golang/commit/dc6502590c15678b9616adbf689d61d1a235374e))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))
* team ID to name is now running on backend ([e139c3a](https://git.liamhardman.com/liam/fm24-golang/commit/e139c3ada40cc0fd346c1f140beb96e68608a720))
* wonderkid filter improvements ([b657d7e](https://git.liamhardman.com/liam/fm24-golang/commit/b657d7e72c3916a3a94b561661b4261467ba5f7d))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted kubernetes multipel replica crash fix ([dff5633](https://git.liamhardman.com/liam/fm24-golang/commit/dff5633ec450fd0e2c53d10b19d9660c280eefa8))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* code linting issue fixes ([b9de2d5](https://git.liamhardman.com/liam/fm24-golang/commit/b9de2d5dea1a34304d8bf6e3d8db9976af77000b))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* fixed percentile values being 0 or - counting ([d30d46e](https://git.liamhardman.com/liam/fm24-golang/commit/d30d46e5ad691eb87591661e1e90d15f824051f8))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented runtime assessment of otel usage ([6d84928](https://git.liamhardman.com/liam/fm24-golang/commit/6d849285cec0bd25ef281d4bfca6e07e0aba973a))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* more backend optimization ([c7b8f17](https://git.liamhardman.com/liam/fm24-golang/commit/c7b8f178624e05c3dee68a35fbe44315540e7016))
* more code fixes from precommit ([0ba5133](https://git.liamhardman.com/liam/fm24-golang/commit/0ba513369a51ed330eb0b1f5f1f3118aa59454a8))
* new demo dataset ID ([61e5613](https://git.liamhardman.com/liam/fm24-golang/commit/61e5613b88bb275ee471476e04e36c4a879a3120))
* only logging non-200 requests ([24cc1d9](https://git.liamhardman.com/liam/fm24-golang/commit/24cc1d96c53d44b0220b8253ffefb92cd1665b4a))
* only logging non-200 requests ([6d50bde](https://git.liamhardman.com/liam/fm24-golang/commit/6d50bde4b3f17ef80556f413199ff01701092d2d))
* optimizations ([3051728](https://git.liamhardman.com/liam/fm24-golang/commit/3051728e3746add5aee8b4d65f57728ab864d755))
* optimizations ([b30c77a](https://git.liamhardman.com/liam/fm24-golang/commit/b30c77a84a58617696ad4329cf35f21fa1c262d7))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* some changes to team handlers which i can't remember, and readme change ([407b862](https://git.liamhardman.com/liam/fm24-golang/commit/407b862cac6072dd6b83e6d3ba86d4209748f70d))
* some fixes for before github release ([2f54fe9](https://git.liamhardman.com/liam/fm24-golang/commit/2f54fe90ff0f3c42b5055ec0c38a02c8b9124aa4))
* some linting fixes again ([96f4d02](https://git.liamhardman.com/liam/fm24-golang/commit/96f4d02197aa981c7dd8e33bd209f03e773c8806))
* team logo util fix ([7041a75](https://git.liamhardman.com/liam/fm24-golang/commit/7041a758ccdf2f54ec14f04c07da1f0151f7f005))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* changed some comments and removed development.md ([841ff03](https://git.liamhardman.com/liam/fm24-golang/commit/841ff03e33f66b66d8a01f9f031ff0df2c806f54))
* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))
* removed comments that weren't necessary ([26d4ed4](https://git.liamhardman.com/liam/fm24-golang/commit/26d4ed45cbc6c66bcd9f7e5d39570cff9ca1cfb8))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-30)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* more filtering on the performance page ([4f19675](https://git.liamhardman.com/liam/fm24-golang/commit/4f19675dad324d9d3ffe96c4f83e5b8b25db6646))
* moved config to its own config file ([dc65025](https://git.liamhardman.com/liam/fm24-golang/commit/dc6502590c15678b9616adbf689d61d1a235374e))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))
* team ID to name is now running on backend ([e139c3a](https://git.liamhardman.com/liam/fm24-golang/commit/e139c3ada40cc0fd346c1f140beb96e68608a720))
* wonderkid filter improvements ([b657d7e](https://git.liamhardman.com/liam/fm24-golang/commit/b657d7e72c3916a3a94b561661b4261467ba5f7d))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted kubernetes multipel replica crash fix ([dff5633](https://git.liamhardman.com/liam/fm24-golang/commit/dff5633ec450fd0e2c53d10b19d9660c280eefa8))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* code linting issue fixes ([b9de2d5](https://git.liamhardman.com/liam/fm24-golang/commit/b9de2d5dea1a34304d8bf6e3d8db9976af77000b))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* fixed percentile values being 0 or - counting ([d30d46e](https://git.liamhardman.com/liam/fm24-golang/commit/d30d46e5ad691eb87591661e1e90d15f824051f8))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* more backend optimization ([c7b8f17](https://git.liamhardman.com/liam/fm24-golang/commit/c7b8f178624e05c3dee68a35fbe44315540e7016))
* more code fixes from precommit ([0ba5133](https://git.liamhardman.com/liam/fm24-golang/commit/0ba513369a51ed330eb0b1f5f1f3118aa59454a8))
* new demo dataset ID ([61e5613](https://git.liamhardman.com/liam/fm24-golang/commit/61e5613b88bb275ee471476e04e36c4a879a3120))
* optimizations ([3051728](https://git.liamhardman.com/liam/fm24-golang/commit/3051728e3746add5aee8b4d65f57728ab864d755))
* optimizations ([b30c77a](https://git.liamhardman.com/liam/fm24-golang/commit/b30c77a84a58617696ad4329cf35f21fa1c262d7))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* some changes to team handlers which i can't remember, and readme change ([407b862](https://git.liamhardman.com/liam/fm24-golang/commit/407b862cac6072dd6b83e6d3ba86d4209748f70d))
* some linting fixes again ([96f4d02](https://git.liamhardman.com/liam/fm24-golang/commit/96f4d02197aa981c7dd8e33bd209f03e773c8806))
* team logo util fix ([7041a75](https://git.liamhardman.com/liam/fm24-golang/commit/7041a758ccdf2f54ec14f04c07da1f0151f7f005))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))
* removed comments that weren't necessary ([26d4ed4](https://git.liamhardman.com/liam/fm24-golang/commit/26d4ed45cbc6c66bcd9f7e5d39570cff9ca1cfb8))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-27)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))
* team ID to name is now running on backend ([e139c3a](https://git.liamhardman.com/liam/fm24-golang/commit/e139c3ada40cc0fd346c1f140beb96e68608a720))
* wonderkid filter improvements ([b657d7e](https://git.liamhardman.com/liam/fm24-golang/commit/b657d7e72c3916a3a94b561661b4261467ba5f7d))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted kubernetes multipel replica crash fix ([dff5633](https://git.liamhardman.com/liam/fm24-golang/commit/dff5633ec450fd0e2c53d10b19d9660c280eefa8))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* code linting issue fixes ([b9de2d5](https://git.liamhardman.com/liam/fm24-golang/commit/b9de2d5dea1a34304d8bf6e3d8db9976af77000b))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* fixed percentile values being 0 or - counting ([d30d46e](https://git.liamhardman.com/liam/fm24-golang/commit/d30d46e5ad691eb87591661e1e90d15f824051f8))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* more backend optimization ([c7b8f17](https://git.liamhardman.com/liam/fm24-golang/commit/c7b8f178624e05c3dee68a35fbe44315540e7016))
* optimizations ([3051728](https://git.liamhardman.com/liam/fm24-golang/commit/3051728e3746add5aee8b4d65f57728ab864d755))
* optimizations ([b30c77a](https://git.liamhardman.com/liam/fm24-golang/commit/b30c77a84a58617696ad4329cf35f21fa1c262d7))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* some linting fixes again ([96f4d02](https://git.liamhardman.com/liam/fm24-golang/commit/96f4d02197aa981c7dd8e33bd209f03e773c8806))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))
* removed comments that weren't necessary ([26d4ed4](https://git.liamhardman.com/liam/fm24-golang/commit/26d4ed45cbc6c66bcd9f7e5d39570cff9ca1cfb8))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-27)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))
* wonderkid filter improvements ([b657d7e](https://git.liamhardman.com/liam/fm24-golang/commit/b657d7e72c3916a3a94b561661b4261467ba5f7d))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted kubernetes multipel replica crash fix ([dff5633](https://git.liamhardman.com/liam/fm24-golang/commit/dff5633ec450fd0e2c53d10b19d9660c280eefa8))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* more backend optimization ([c7b8f17](https://git.liamhardman.com/liam/fm24-golang/commit/c7b8f178624e05c3dee68a35fbe44315540e7016))
* optimizations ([b30c77a](https://git.liamhardman.com/liam/fm24-golang/commit/b30c77a84a58617696ad4329cf35f21fa1c262d7))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))
* removed comments that weren't necessary ([26d4ed4](https://git.liamhardman.com/liam/fm24-golang/commit/26d4ed45cbc6c66bcd9f7e5d39570cff9ca1cfb8))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-27)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))
* wonderkid filter improvements ([b657d7e](https://git.liamhardman.com/liam/fm24-golang/commit/b657d7e72c3916a3a94b561661b4261467ba5f7d))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted kubernetes multipel replica crash fix ([dff5633](https://git.liamhardman.com/liam/fm24-golang/commit/dff5633ec450fd0e2c53d10b19d9660c280eefa8))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* more backend optimization ([c7b8f17](https://git.liamhardman.com/liam/fm24-golang/commit/c7b8f178624e05c3dee68a35fbe44315540e7016))
* optimizations ([b30c77a](https://git.liamhardman.com/liam/fm24-golang/commit/b30c77a84a58617696ad4329cf35f21fa1c262d7))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted kubernetes multipel replica crash fix ([dff5633](https://git.liamhardman.com/liam/fm24-golang/commit/dff5633ec450fd0e2c53d10b19d9660c280eefa8))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* more backend optimization ([c7b8f17](https://git.liamhardman.com/liam/fm24-golang/commit/c7b8f178624e05c3dee68a35fbe44315540e7016))
* optimizations ([b30c77a](https://git.liamhardman.com/liam/fm24-golang/commit/b30c77a84a58617696ad4329cf35f21fa1c262d7))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted kubernetes multipel replica crash fix ([dff5633](https://git.liamhardman.com/liam/fm24-golang/commit/dff5633ec450fd0e2c53d10b19d9660c280eefa8))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* optimizations ([b30c77a](https://git.liamhardman.com/liam/fm24-golang/commit/b30c77a84a58617696ad4329cf35f21fa1c262d7))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted kubernetes multipel replica crash fix ([dff5633](https://git.liamhardman.com/liam/fm24-golang/commit/dff5633ec450fd0e2c53d10b19d9660c280eefa8))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* optimizations ([b30c77a](https://git.liamhardman.com/liam/fm24-golang/commit/b30c77a84a58617696ad4329cf35f21fa1c262d7))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted kubernetes multipel replica crash fix ([dff5633](https://git.liamhardman.com/liam/fm24-golang/commit/dff5633ec450fd0e2c53d10b19d9660c280eefa8))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted kubernetes multipel replica crash fix ([dff5633](https://git.liamhardman.com/liam/fm24-golang/commit/dff5633ec450fd0e2c53d10b19d9660c280eefa8))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted kubernetes multipel replica crash fix ([dff5633](https://git.liamhardman.com/liam/fm24-golang/commit/dff5633ec450fd0e2c53d10b19d9660c280eefa8))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* hopeful frontend upload fix ([2812494](https://git.liamhardman.com/liam/fm24-golang/commit/28124949c06144eb242104ff1530c2cab4c57651))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* max frontend upload size ([828c87c](https://git.liamhardman.com/liam/fm24-golang/commit/828c87c1665a06f021e9db55bad5a39f1814ac7f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* api route fix ([cec3602](https://git.liamhardman.com/liam/fm24-golang/commit/cec36023090131322370a30b408bad1accd0110a))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel fixes for backend ([ba10975](https://git.liamhardman.com/liam/fm24-golang/commit/ba109754f2f9229b80bc4d3682e1d9d4016d6ce0))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* phase 2 of backend otel ([905f321](https://git.liamhardman.com/liam/fm24-golang/commit/905f3216931c86226e96a3582e1d729c1f67f1fc))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* examplars in metrics and fixed missing import ([730c871](https://git.liamhardman.com/liam/fm24-golang/commit/730c871c3d946df97e065e03e78d30fbb5511a6a))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* phase 1 of backend otel improvements ([f8b39c2](https://git.liamhardman.com/liam/fm24-golang/commit/f8b39c28c28d01fc232a707f1d8507a64ed2be02))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* config.js fix ([13c2e92](https://git.liamhardman.com/liam/fm24-golang/commit/13c2e92b8ba1e8bfb3c6dd3262a8db0bf988669f))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-26)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))
* removed frontend tracing ([0105bc8](https://git.liamhardman.com/liam/fm24-golang/commit/0105bc8b612da67c9f7da0b124b63ea22247a26f))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* config injection fix ([29902f5](https://git.liamhardman.com/liam/fm24-golang/commit/29902f5aea75cee76532595610b5e606983f9bde))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-25)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* https:// ([834b8de](https://git.liamhardman.com/liam/fm24-golang/commit/834b8de4e52addada9e4481d981a5e8965b42adc))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-25)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))
* url fix ([0fb9694](https://git.liamhardman.com/liam/fm24-golang/commit/0fb969469bed92b58ef849b719b7f84673e41d39))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-25)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* telemetry routing ([690513d](https://git.liamhardman.com/liam/fm24-golang/commit/690513d1882c79d06956831cc620e1d8efb631fc))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-25)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** implement fallback methods for WebTracerProvider span processor ([2f5a299](https://git.liamhardman.com/liam/fm24-golang/commit/2f5a299af4fd52a7cde566234237c6f50f26d8dd))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-25)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-25)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* **frontend:** resolve resourceFromAttributes variable scoping issue ([cc464e8](https://git.liamhardman.com/liam/fm24-golang/commit/cc464e80b51721de1deda318210ebc3ecf363508))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-25)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* **frontend:** resolve Resource constructor verification error ([963d5d3](https://git.liamhardman.com/liam/fm24-golang/commit/963d5d3c03d24b50227dcc93b6ec9c015ca3d0ea))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-25)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* **frontend:** resolve OpenTelemetry Resource import issue ([cc249a2](https://git.liamhardman.com/liam/fm24-golang/commit/cc249a2534c9c5674cb4cb610f49ada7560ed61c))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-25)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** improved OpenTelemetry module imports to resolve constructor errors ([8653786](https://git.liamhardman.com/liam/fm24-golang/commit/865378648afb4dbcdf7e2128849795c174403143))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-25)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* **frontend:** resolve OpenTelemetry require() error with improved bundling and error handling ([0c4dff0](https://git.liamhardman.com/liam/fm24-golang/commit/0c4dff03557317d45de291ec7733d8a6ad61adb1))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-25)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* config ([2b2afb0](https://git.liamhardman.com/liam/fm24-golang/commit/2b2afb088efc66754d30f375f5ac13723c2d3cc6))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* otel var fix ([5347c8b](https://git.liamhardman.com/liam/fm24-golang/commit/5347c8bd0d7d4a62d7b749b59a2d02a2e0867eae))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-24)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* idk tbh ([4ca79d3](https://git.liamhardman.com/liam/fm24-golang/commit/4ca79d37377f1887678d6795f21182f52322ddb0))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-24)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* added frontend stuff to config injection ([addd0a6](https://git.liamhardman.com/liam/fm24-golang/commit/addd0a660c472352f887adabc5532836da749545))
* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.13.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.13.0) (2025-06-24)

### 🚀 Features

* frontend otel ([2eac60b](https://git.liamhardman.com/liam/fm24-golang/commit/2eac60bdcb5cf98794908b9279dad5d182bdd831))

### 🐛 Bug Fixes

* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.12.11](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.12.11) (2025-06-24)

### 🐛 Bug Fixes

* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* attempted otel frontend fix ([56b034f](https://git.liamhardman.com/liam/fm24-golang/commit/56b034fc2472513a401117849ac297b650b2023f))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.12.11](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.12.11) (2025-06-24)

### 🐛 Bug Fixes

* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* backend tracing and logging improvements ([fccb3e5](https://git.liamhardman.com/liam/fm24-golang/commit/fccb3e5da092085009c8b4c89a6022b10bd6c1e1))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.12.11](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.12.11) (2025-06-24)

### 🐛 Bug Fixes

* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))
* test change ([f41edec](https://git.liamhardman.com/liam/fm24-golang/commit/f41edec58e4610163601be13fa8e15fc5f6b3ccd))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.12.11](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.12.11) (2025-06-24)

### 🐛 Bug Fixes

* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* optimizations for frontend ([7f8b862](https://git.liamhardman.com/liam/fm24-golang/commit/7f8b86235d9b6a127d2913f67d4419544e2216f7))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.12.11](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.12.11) (2025-06-24)

### 🐛 Bug Fixes

* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* frontend optimization ([7132272](https://git.liamhardman.com/liam/fm24-golang/commit/713227216d5a64878230fb91aacc16d1bb21541b))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.12.11](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.12.11) (2025-06-24)

### 🐛 Bug Fixes

* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* league filter cachin stuffs ([b9d466d](https://git.liamhardman.com/liam/fm24-golang/commit/b9d466dcf1eeeb7328b3baabe9d79163dd77992f))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.12.11](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.12.11) (2025-06-23)

### 🐛 Bug Fixes

* attempted caching optimization ([150823d](https://git.liamhardman.com/liam/fm24-golang/commit/150823d41a9c24ebdacaf5a7e0039180e5821366))
* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.12.11](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.12.11) (2025-06-21)

### 🐛 Bug Fixes

* cache implemented for team logos ([428f0a7](https://git.liamhardman.com/liam/fm24-golang/commit/428f0a7a21d6827d1a1e9236de07edb1b7650b43))
* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.12.11](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.12.11) (2025-06-21)

### 🐛 Bug Fixes

* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* instantly opening playerdetaildialog ([821c5b1](https://git.liamhardman.com/liam/fm24-golang/commit/821c5b169ff469342033219c26cb77315cbc5e41))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.12.11](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.12.11) (2025-06-21)

### 🐛 Bug Fixes

* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))
* playerdetaildialog optimization ([7f78200](https://git.liamhardman.com/liam/fm24-golang/commit/7f782000f9c71926ae8018a33351bd8760f3b859))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.12.11](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.10...v1.12.11) (2025-06-20)

### 🐛 Bug Fixes

* implemented some weighting for lower ID teams ([8403115](https://git.liamhardman.com/liam/fm24-golang/commit/84031155745fc4ac8732010c3e82340f1850a6f2))

### 📚 Documentation

* docs update ([532b5f1](https://git.liamhardman.com/liam/fm24-golang/commit/532b5f11a39ac5e679411e44af1394efee0e84d7))

## [1.12.10](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.9...v1.12.10) (2025-06-15)

### 🐛 Bug Fixes

* fixes for useTeamLogos for matching ([4de53b8](https://git.liamhardman.com/liam/fm24-golang/commit/4de53b8141a53089669f8057e0f08140f41125cb))

## [1.12.9](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.8...v1.12.9) (2025-06-15)

### 🐛 Bug Fixes

* select all divisions ([b297dde](https://git.liamhardman.com/liam/fm24-golang/commit/b297dde2cf7051d4bc677a57617484bf2230087f))

## [1.12.8](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.7...v1.12.8) (2025-06-15)

### 🐛 Bug Fixes

* using mean of 7 role ratings instead of all for overall calc ([b6b7001](https://git.liamhardman.com/liam/fm24-golang/commit/b6b7001a7bad47034a27c5834e6bb3c39ef96e91))

## [1.12.7](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.6...v1.12.7) (2025-06-15)

### 🐛 Bug Fixes

* attempted release for both repos ([c362843](https://git.liamhardman.com/liam/fm24-golang/commit/c362843900bb834e53e8722b0b05722612f04368))
* attempting to have both gitea and github releases ([f24c0e2](https://git.liamhardman.com/liam/fm24-golang/commit/f24c0e2e32bf6b26e3007dcc5d94944a15787a05))
* first gh-ready release ([b85b53d](https://git.liamhardman.com/liam/fm24-golang/commit/b85b53da1874aef6618b00906b619baf58a4becf))
* fixed --config var ([ef2673d](https://git.liamhardman.com/liam/fm24-golang/commit/ef2673d689edb71d2d6be2803d74af9e2b1e1019))
* fixed --config var ([216e645](https://git.liamhardman.com/liam/fm24-golang/commit/216e64545e7103bdb814950c2c457de157539550))
* only github 2 electric boogaloo ([27d8e12](https://git.liamhardman.com/liam/fm24-golang/commit/27d8e123c77379faa7c7ce09f287b324a66a3619))
* only github 2 electric boogaloo ([e99c160](https://git.liamhardman.com/liam/fm24-golang/commit/e99c160a64bf588258e8b97a6555c3599267066c))
* only github 2 electric boogaloo ([abd55c2](https://git.liamhardman.com/liam/fm24-golang/commit/abd55c2b353a9c06e55c77875923d4a6da236dad))
* only github 2 electric boogaloo ([add4002](https://git.liamhardman.com/liam/fm24-golang/commit/add400276fd3e798ee5e329e45c9252164b03602))
* only github 2 electric boogaloo ([56194c7](https://git.liamhardman.com/liam/fm24-golang/commit/56194c7f06f050837d9f3b6c3c801d72106159c8))
* only github 2 electric boogaloo ([79c9f7d](https://git.liamhardman.com/liam/fm24-golang/commit/79c9f7d13ba393d29bf5578a888e652cb6005903))
* only github 2 electric boogaloo ([d70bf4b](https://git.liamhardman.com/liam/fm24-golang/commit/d70bf4b5e6e61d07b783e7bc032512289b63f14c))
* only github release ([c5a3719](https://git.liamhardman.com/liam/fm24-golang/commit/c5a37199431f846325559aa47411d30c2fae53a3))

## [1.12.6](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.5...v1.12.6) (2025-06-14)

### 🐛 Bug Fixes

* fixed passing volume graph ([7ac8e59](https://git.liamhardman.com/liam/fm24-golang/commit/7ac8e5939f9dd56a0340c455aeefb822efe0aa4d))

## [1.12.5](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.4...v1.12.5) (2025-06-14)

### 🐛 Bug Fixes

* tabs in tabs for performance page for better visibility ([6393f1a](https://git.liamhardman.com/liam/fm24-golang/commit/6393f1a59d49f286f454ca871db53ec3db878ee6))

## [1.12.4](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.3...v1.12.4) (2025-06-13)

### 🐛 Bug Fixes

* default closed dropdowns on settings modal and disclaimer for club matching ([4e65a3f](https://git.liamhardman.com/liam/fm24-golang/commit/4e65a3f0e7a3af75b72a0043f520c4d024e895e6))

## [1.12.3](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.2...v1.12.3) (2025-06-13)

### 🐛 Bug Fixes

* added ability to configure storage retention ([4348438](https://git.liamhardman.com/liam/fm24-golang/commit/43484381397a4dd58a1dad59bce22a397099a623))

## [1.12.2](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.1...v1.12.2) (2025-06-13)

### 🐛 Bug Fixes

* fixed plot sizing ([4c50f18](https://git.liamhardman.com/liam/fm24-golang/commit/4c50f18781db8030a48b256e7666d6aa842613dd))

## [1.12.1](https://git.liamhardman.com/liam/fm24-golang/compare/v1.12.0...v1.12.1) (2025-06-13)

### 🐛 Bug Fixes

* auto defining of minimum minutes so at least 100 players match, styling fixes etc ([0501847](https://git.liamhardman.com/liam/fm24-golang/commit/0501847c1614da5a94ce904bb06a2b54cab3a5c6))
* styling for performance page ([2974b58](https://git.liamhardman.com/liam/fm24-golang/commit/2974b5816874e5b11082a4a0fb5b3bce50b408fe))

## [1.12.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.11.0...v1.12.0) (2025-06-13)

### 🚀 Features

* tabulated performance page ([cab76f5](https://git.liamhardman.com/liam/fm24-golang/commit/cab76f55a40a39bb6e2d9449e52bcc609248bedd))

## [1.11.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.10.1...v1.11.0) (2025-06-12)

### 🚀 Features

* very initial scatter plot implementation ([83c897c](https://git.liamhardman.com/liam/fm24-golang/commit/83c897caa967e5b168580b7a9b8646580efdef8c))

## [1.10.1](https://git.liamhardman.com/liam/fm24-golang/compare/v1.10.0...v1.10.1) (2025-06-12)

### 🐛 Bug Fixes

* new stats in performance page ([e0ab190](https://git.liamhardman.com/liam/fm24-golang/commit/e0ab190b341ac972b6569de76c0d374e1c3088d0))

## [1.10.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.9.0...v1.10.0) (2025-06-12)

### 🚀 Features

* new stats ([7d0bffa](https://git.liamhardman.com/liam/fm24-golang/commit/7d0bffa126454a7d8ead4bb4952600a1da94bb71))

### 🐛 Bug Fixes

* fix for minutes processing ([d3bf2aa](https://git.liamhardman.com/liam/fm24-golang/commit/d3bf2aa1a11aba658c73e780edaec23674bd95d6))

## [1.9.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.8.5...v1.9.0) (2025-06-12)

### 🚀 Features

* performance page very WIP ([cd24024](https://git.liamhardman.com/liam/fm24-golang/commit/cd24024152bf9b31abf50cf5dc674cbe018d32ee))

## [1.8.5](https://git.liamhardman.com/liam/fm24-golang/compare/v1.8.4...v1.8.5) (2025-06-12)

### 🐛 Bug Fixes

* nation page table fix ([b82fd9f](https://git.liamhardman.com/liam/fm24-golang/commit/b82fd9f1bae533b946ac281b3dd17c707254ab17))

## [1.8.4](https://git.liamhardman.com/liam/fm24-golang/compare/v1.8.3...v1.8.4) (2025-06-12)

### 🐛 Bug Fixes

* fixed typeerror stopping ratings for nations from showign on nations page ([52f8192](https://git.liamhardman.com/liam/fm24-golang/commit/52f8192a79a68b63e463a1b45c128fd16bd51285))

## [1.8.3](https://git.liamhardman.com/liam/fm24-golang/compare/v1.8.2...v1.8.3) (2025-06-10)

### 🐛 Bug Fixes

* attempted slowdown fix ([7338c92](https://git.liamhardman.com/liam/fm24-golang/commit/7338c926cf9f14de72353b76d3a6a2f9416d2343))

## [1.8.2](https://git.liamhardman.com/liam/fm24-golang/compare/v1.8.1...v1.8.2) (2025-06-10)

### 🐛 Bug Fixes

* multi replica fixes ([efb6d56](https://git.liamhardman.com/liam/fm24-golang/commit/efb6d5659aab790552dfd47522d5990778056e14))

## [1.8.1](https://git.liamhardman.com/liam/fm24-golang/compare/v1.8.0...v1.8.1) (2025-06-10)

### 🐛 Bug Fixes

* hopeful team logo fix ([1573638](https://git.liamhardman.com/liam/fm24-golang/commit/1573638f17a75ff8e0e742ed1f2669f1f2d1438f))

## [1.8.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.7.1...v1.8.0) (2025-06-10)

### 🚀 Features

* ability to use sortitoutsi CDN ([f7e602f](https://git.liamhardman.com/liam/fm24-golang/commit/f7e602f64885add8d7465c805e8ed0227aafbb77))

## [1.7.1](https://git.liamhardman.com/liam/fm24-golang/compare/v1.7.0...v1.7.1) (2025-06-10)

### 🐛 Bug Fixes

* change to render images at 256x256 instead of 512x512 ([23222e5](https://git.liamhardman.com/liam/fm24-golang/commit/23222e58a0c0b9d55924948f5c830837522ea54a))

## [1.7.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.6.0...v1.7.0) (2025-06-09)

### 🚀 Features

* multiple replica support ([e3411e1](https://git.liamhardman.com/liam/fm24-golang/commit/e3411e14a470ca88083e22c0f01778c8b46a77b9))

### 🐛 Bug Fixes

* attempted resolution of issues in build due to api url changes ([92eed18](https://git.liamhardman.com/liam/fm24-golang/commit/92eed185689fec048379dff8545949ca1b469554))

## [1.6.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.5.0...v1.6.0) (2025-06-09)

### 🚀 Features

* fixes and settings for player masking support ([328588e](https://git.liamhardman.com/liam/fm24-golang/commit/328588e1f74604d362b152005d6be51974e06ab3))

## [1.5.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.4.1...v1.5.0) (2025-06-09)

### 🚀 Features

* initial implementation of attribute masking support ([2a341bd](https://git.liamhardman.com/liam/fm24-golang/commit/2a341bdbc4f754c3b434696dc2cc7826a578edd1))

## [1.4.1](https://git.liamhardman.com/liam/fm24-golang/compare/v1.4.0...v1.4.1) (2025-06-09)

### 🐛 Bug Fixes

* add semantic-release configuration and debug workflow to resolve ERELEASEBRANCHES error ([7097358](https://git.liamhardman.com/liam/fm24-golang/commit/7097358536c01c991202d7f40b1e72b67159dc3f))
* changed url for releases ([87dd8b9](https://git.liamhardman.com/liam/fm24-golang/commit/87dd8b999165dfb17148d628d75c03be9ac73e8a))
* GitHub readiness update ([0b7e285](https://git.liamhardman.com/liam/fm24-golang/commit/0b7e285e6d5af6b684ff98bbe043a3c9beb94501))
* hopeful fix of release pipeline for the 89432894289th time ([b809aca](https://git.liamhardman.com/liam/fm24-golang/commit/b809acad76c3c0784011613ec027cd872bbef033))
* hopeful fix of releases ([9ca6844](https://git.liamhardman.com/liam/fm24-golang/commit/9ca6844b1f85084f4c9ccfcd84b7a389b50389e4))
* hopefully fixing up release config ([d2666e4](https://git.liamhardman.com/liam/fm24-golang/commit/d2666e40e55fcb90d084827b98adc58efc0553e1))
* small readme update ([d611dac](https://git.liamhardman.com/liam/fm24-golang/commit/d611dacd6c9ece4eeaec794b48976f6b408326f6))

## [1.4.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.3.0...v1.4.0) (2025-06-09)

### 🚀 Features

* MUCH better export controls ([d3adfd1](https://git.liamhardman.com/liam/fm24-golang/commit/d3adfd18401f92f319965c54fb735b5fb7ffaa27))

### 🐛 Bug Fixes

* hopefully fixing release pipeline ([367b47a](https://git.liamhardman.com/liam/fm24-golang/commit/367b47a21f18d9a76a970f964214b0a7c96a630b))

## [1.3.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.2.0...v1.3.0) (2025-06-09)

### 🚀 Features

* Data export ([68ecf96](https://git.liamhardman.com/liam/fm24-golang/commit/68ecf962a20c36153d8b291e97d0ae85d71c107e))
* MUCH better export controls ([d3adfd1](https://git.liamhardman.com/liam/fm24-golang/commit/d3adfd1fc4b138cc9583e4fe8828fb35556ee665))
* nation player filters ([c4cdd55](https://git.liamhardman.com/liam/fm24-golang/commit/c4cdd55ebdce05d543ae652c30ef558aec3bc550))
* revamp of the dataset page ([43196d5](https://git.liamhardman.com/liam/fm24-golang/commit/43196d56b7ff7a5122f1167fee36acd1e78e0d74))
* setting overall rating to be mean of all roles rather than best ([b035b23](https://git.liamhardman.com/liam/fm24-golang/commit/b035b23034b831dcebd180c246d424b0ddb3d7bf))
* testing if logic change works for scaled vs non-scaled ratings, added settings menu ([5e65048](https://git.liamhardman.com/liam/fm24-golang/commit/5e65048e54b138cc9583e4fe8828fb35556ee665))

### 🐛 Bug Fixes

* ability to toggle logos on and off ([7076559](https://git.liamhardman.com/liam/fm24-golang/commit/70765597d11ab7ec9f52086e1430133f47b50ea2))
* added more tests ([8e35104](https://git.liamhardman.com/liam/fm24-golang/commit/8e351041d4049fa2530006ed97f815367e1dbfe2))
* attempted fuzzy fix ([bf0d2f6](https://git.liamhardman.com/liam/fm24-golang/commit/bf0d2f61e0d24eae432e5069761186a45c495e31))
* better flag quality ([ec6021b](https://git.liamhardman.com/liam/fm24-golang/commit/ec6021b4f88a16f293603e36d61d592545f4a4eb))
* better landing page ([72c0590](https://git.liamhardman.com/liam/fm24-golang/commit/72c0590fe24b604881951ff02f6a2fd1dd6bb53b))
* data refresh when settings changed ([cb4f843](https://git.liamhardman.com/liam/fm24-golang/commit/cb4f843348d54f0850249bb0122c456bd5495c27))
* doc page ([e2ca185](https://git.liamhardman.com/liam/fm24-golang/commit/e2ca18554edbb765963b2a4275c6bd4b59045904))
* docs page stuff ([d93314a](https://git.liamhardman.com/liam/fm24-golang/commit/d93314a4fcf2cce610c45bdc27c5609bbef8084c))
* docspage ([05e00c6](https://git.liamhardman.com/liam/fm24-golang/commit/05e00c6597b21ec6db04916c78cfb528312be9ea))
* hopeful fix for team logo matching ([7eb6ec0](https://git.liamhardman.com/liam/fm24-golang/commit/7eb6ec01644748b1eb9e6cb9e7e0d5aa03c9b729))
* hopefully adding working team logos ([9f9635c](https://git.liamhardman.com/liam/fm24-golang/commit/9f9635cfb551b9a2c7036625ca06aa26ab84371c))
* improved landing page ([56bf115](https://git.liamhardman.com/liam/fm24-golang/commit/56bf115cc36d3d8b32a5b3367e11839e68a3e782))
* local deploy doc ([bb397a2](https://git.liamhardman.com/liam/fm24-golang/commit/bb397a262336f5dc722647f0f9cb12392db3ec82))
* logging for scaling algorithm switch ([6bd19e7](https://git.liamhardman.com/liam/fm24-golang/commit/6bd19e7c3495779da553c90bd58ae3321516124f))
* making the scaled rating setting actually work ([a1310fd](https://git.liamhardman.com/liam/fm24-golang/commit/a1310fde84ba686ab2848f897d93a591522e9b94))
* minor roadmap ([495114c](https://git.liamhardman.com/liam/fm24-golang/commit/495114c86bbc9111ad78ad5839d3caf9ca1cc35a))
* no longer showing logos directly on the playerdatatable ([9f82f93](https://git.liamhardman.com/liam/fm24-golang/commit/9f82f939b9d652396ab83051cac5ab0f1e52cb4a))
* optimized fuzzy finding ([48f5817](https://git.liamhardman.com/liam/fm24-golang/commit/48f5817ec58a82f4313b04a3212c3800a22d9942))
* readme update ([d5b0854](https://git.liamhardman.com/liam/fm24-golang/commit/d5b085494dea7102a519bf5080f2303f0ef2fae9))
* reorganization of settings ([ab9e994](https://git.liamhardman.com/liam/fm24-golang/commit/ab9e994660c7601bc8d4236bc85c6ec3b86101aa))
* roadmap ([1160a77](https://git.liamhardman.com/liam/fm24-golang/commit/1160a774d682dcc9193c35c65b4eacf01f6821b9))
* small text improvements ([b09f489](https://git.liamhardman.com/liam/fm24-golang/commit/b09f48921aa8467ff0a1aea1e4e3b70cb1ec3e28))
* team and face switch in settings actually work ([72a6630](https://git.liamhardman.com/liam/fm24-golang/commit/72a6630dca43a5c822ce4b4fe7b03947a7993337))
* team logo in team view page,and path fix ([5372f9d](https://git.liamhardman.com/liam/fm24-golang/commit/5372f9df024665d7e1c88187bccac9a3bc695057))
* Testing fix team mapping ([2d7e8f7](https://git.liamhardman.com/liam/fm24-golang/commit/2d7e8f7149067d4b931b19575d97e92ef5423e8a))
* unit tests ([1c979f4](https://git.liamhardman.com/liam/fm24-golang/commit/1c979f4cf11ef6ed9216a7b10f246f550b42838c))
* upload component improvement ([fc91ded](https://git.liamhardman.com/liam/fm24-golang/commit/fc91dedc3f462680c75ea60b993f73b8701330d2))

## [1.2.0](https://git.liamhardman.com/liam/fm24-golang/compare/v1.1.4...v1.2.0) (2025-05-25)

### 🚀 Features

* rebrand to FM-Dash ([beb61e3](https://git.liamhardman.com/liam/fm24-golang/commit/beb61e3f69f800ce204105b236e500483d871b81))

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

## [Unreleased]

### Fixed
- **Multi-replica deployment consistency**: Fixed intermittent "dataset not found" errors in multi-replica deployments
  - Added session affinity (ClientIP) to Kubernetes backend service to ensure request stickiness
  - ~~Modified data retrieval logic to prioritize persistent storage over in-memory cache for better cross-replica consistency~~ **Reverted for performance**
  - ~~Changed upload process from async to sync storage to ensure immediate data availability~~ **Reverted for performance**
  - Improved hybrid storage retrieval with async memory warming
  - Added retry mechanism with exponential backoff in frontend service for handling race conditions
  - ~~Enhanced storage verification step in upload process~~ **Removed for performance**

### Performance
- **Restored original performance**: Reverted storage optimizations that caused slowdown in single-replica deployments
  - Uploads use async storage again for fast response times
  - Data retrieval prioritizes fast in-memory cache over persistent storage
  - Removed blocking verification step from upload process
  - Session affinity in Kubernetes handles most multi-replica consistency issues
