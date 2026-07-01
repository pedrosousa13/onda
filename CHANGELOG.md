# Changelog

## [1.5.0](https://github.com/pedrosousa13/onda/compare/v1.4.0...v1.5.0) (2026-07-01)


### Features

* **browse:** browse the offline catalog by country / genre / language, with sortable results ([#36](https://github.com/pedrosousa13/onda/issues/36)) ([d1f5556](https://github.com/pedrosousa13/onda/commit/d1f5556db6072c21c8844861b61836e7aeed4c1d))
* opt-in offline station catalog with background download + progress ([#37](https://github.com/pedrosousa13/onda/issues/37)) ([ec52fda](https://github.com/pedrosousa13/onda/commit/ec52fdaa95ae96151bd692c015c6cc17db456b60))

## [1.4.0](https://github.com/pedrosousa13/onda/compare/v1.3.0...v1.4.0) (2026-07-01)


### Features

* loudness normalization to even out station volume ([#28](https://github.com/pedrosousa13/onda/issues/28)) ([90351d4](https://github.com/pedrosousa13/onda/commit/90351d4dc5aa6fb026cc4b0d6dc38056345eb6e9))
* **tui:** recently-played view with opt-in local history ([#25](https://github.com/pedrosousa13/onda/issues/25)) ([052ad3b](https://github.com/pedrosousa13/onda/commit/052ad3b6eaf67e9e1e78a459fc658e47423d90e3))

## [1.3.0](https://github.com/pedrosousa13/onda/compare/v1.2.0...v1.3.0) (2026-06-30)


### Features

* persist playback volume across launches ([#27](https://github.com/pedrosousa13/onda/issues/27)) ([1aefd11](https://github.com/pedrosousa13/onda/commit/1aefd1116687dc5d54c4a78715fff3547c3b1890)), closes [#22](https://github.com/pedrosousa13/onda/issues/22)
* **tui:** live-search off toggle (enter-to-search) ([#26](https://github.com/pedrosousa13/onda/issues/26)) ([7ff51d9](https://github.com/pedrosousa13/onda/commit/7ff51d99b207b77670eeba4ef43805724181cb67))

## [1.2.0](https://github.com/pedrosousa13/onda/compare/v1.1.0...v1.2.0) (2026-06-30)


### Features

* **tui:** debounced live search ([53ecd28](https://github.com/pedrosousa13/onda/commit/53ecd28c8bc2369a415295d2e8e3a34826037a45))
* **tui:** UX polish — padding, centered home, connecting feedback, mouse ([59b0f2e](https://github.com/pedrosousa13/onda/commit/59b0f2e2f323738f3193115f855edbb57151ff4b))


### Bug Fixes

* **tui:** persist chosen bitrate on replay; docs + privacy updates ([876573c](https://github.com/pedrosousa13/onda/commit/876573cdf582cb55bc90cf74dfc88cc2bb12c569))

## [1.1.0](https://github.com/pedrosousa13/onda/compare/v1.0.0...v1.1.0) (2026-06-30)


### Features

* cross-platform auto-update (notify + opt-in self-update) ([#6](https://github.com/pedrosousa13/onda/issues/6)) ([6b3185d](https://github.com/pedrosousa13/onda/commit/6b3185d0db9fd00664a6fc8a76924713a86def3f))

## 1.0.0 (2026-06-30)


### Features

* **app:** reword first-run privacy notice; mark shown before prompt ([ca64db8](https://github.com/pedrosousa13/onda/commit/ca64db84513b964c5b39c791c1c5791b0bd13841))
* **app:** wire store, directory, player, TUI + first-run notice ([cdfba5a](https://github.com/pedrosousa13/onda/commit/cdfba5a8c286517138878cef08d55ffadc74c60c))
* **directory:** aggregating Directory with cache + offline fallback ([095316b](https://github.com/pedrosousa13/onda/commit/095316ba3f60228b60cb386ffb3d40f1289cc9d1))
* **directory:** group multi-bitrate records into stations ([3522765](https://github.com/pedrosousa13/onda/commit/35227656c1ce6e5f9a0873486e3ab492eb8d4ce5))
* **directory:** JSON cache with TTL and stale fallback ([d74acb0](https://github.com/pedrosousa13/onda/commit/d74acb0f6f2b46aa49d8842ace1d14b9dae249be))
* **directory:** Phase 2 — multi-field + fuzzy search ([518de0b](https://github.com/pedrosousa13/onda/commit/518de0ba97ff2b7ab1ab1a790477d0bec6261d36))
* **directory:** Phase 3 — broadcaster-prefix/region dedup + honest limits ([ab92c3d](https://github.com/pedrosousa13/onda/commit/ab92c3d455fd989da330dc5726f122c7e35e2690))
* **directory:** Radio Browser client with mirror fallback ([5bb96d6](https://github.com/pedrosousa13/onda/commit/5bb96d646c7713f78f774fdf33f653604c78aa3e))
* **directory:** Source interface + embedded CC0 offline list ([b08ffd6](https://github.com/pedrosousa13/onda/commit/b08ffd65ea9617b49d50cc7fa13cac5aa1a8b116))
* **domain:** add quality-preference variant selection ([4d97a23](https://github.com/pedrosousa13/onda/commit/4d97a234c582cb604df4d5d853f6f904e5fab59c))
* **domain:** add Station and StreamVariant ([44541f9](https://github.com/pedrosousa13/onda/commit/44541f90121198d1d9c8a2bd46279ba032c5c012))
* **player:** headless mpv lifecycle, controls, and events ([b16de58](https://github.com/pedrosousa13/onda/commit/b16de58e9846b54c697936ff2f6b39497fafda81))
* **player:** mpv JSON IPC framing ([a0c9c8c](https://github.com/pedrosousa13/onda/commit/a0c9c8cb2bcb8190e6a5c3fff140a52c8572cef0))
* Popular (top-voted) default view + search spinner ([ffcf8a2](https://github.com/pedrosousa13/onda/commit/ffcf8a2ff69b42562817a8e8f33972a2bec7c3ea))
* race mirrors, merge duplicate streams, HiFi quality + bitrate chooser ([024586c](https://github.com/pedrosousa13/onda/commit/024586c2f73df1a0083353a5b7fefdf1fa6d2523))
* **store:** config persistence (TOML) with privacy-first defaults ([f6a0cce](https://github.com/pedrosousa13/onda/commit/f6a0cce64d80e7943dba3a86c2b8f63fe01bd41e))
* **store:** favorites and custom stations (JSON) ([7ea05d9](https://github.com/pedrosousa13/onda/commit/7ea05d9874859d3fea4de78c2eaaa17bf4ee6e15))
* **tui:** Phase 1 — chooser clarity, metadata sanitize, esc-home hint ([199ab35](https://github.com/pedrosousa13/onda/commit/199ab352fff466502b8847c1d7e94df9986d4037))
* **tui:** Phase 4 — Home view (now-playing + favorites) ([4e6116c](https://github.com/pedrosousa13/onda/commit/4e6116cdb34719c737688317a7bda9da32d3bd06))
* **tui:** root model, messages, list view, play/stop keys ([c910bce](https://github.com/pedrosousa13/onda/commit/c910bce3789e8686cd76d090380352d0a663e493))
* **tui:** search, favorites, add-station, and settings views ([6ce2138](https://github.com/pedrosousa13/onda/commit/6ce213822e57582d5d9d8e75cca54fb8abaf2cd6))
* **tui:** themed Now-Playing hero redesign ([d5bfd91](https://github.com/pedrosousa13/onda/commit/d5bfd91413b9456ef1f26ca9eadc9f33bc73963a))
* **tui:** wire volume control (+/- keys) to the player ([f761612](https://github.com/pedrosousa13/onda/commit/f76161209b20bf07502080046b9c532eb88c35c8))


### Bug Fixes

* **directory:** bust cache on grouping change + stronger name normalization ([c198e80](https://github.com/pedrosousa13/onda/commit/c198e80ff4c69a41702888c694e7e38705cd49c8))
* **directory:** dedupe variants by quality label; treat codec UNKNOWN as unknown ([edd5397](https://github.com/pedrosousa13/onda/commit/edd5397537068f31f42f929496ac1554b63345b3))
* **player:** reap mpv on dial-failure path; add backoff to Unix dial loop ([a66ecea](https://github.com/pedrosousa13/onda/commit/a66eceaa71c2445b53fffe8552b1de9eaf0c7451))
