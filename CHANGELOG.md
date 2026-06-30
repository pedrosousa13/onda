# Changelog

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
