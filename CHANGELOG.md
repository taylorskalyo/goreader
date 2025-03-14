# Changelog

## [2.0.0](https://github.com/taylorskalyo/goreader/compare/v1.0.1...v2.0.0) (2025-03-14)


### ⚠ BREAKING CHANGES

* remove Left / Right actions
* rewrite using tview

### Features

* add basic support for tables ([dcb3de6](https://github.com/taylorskalyo/goreader/commit/dcb3de6c682ac3deca11757a9bb2682d1fae91af))
* parse chapter titles from navigation documents ([533ae0b](https://github.com/taylorskalyo/goreader/commit/533ae0bafeaf60a60fe7d4ecaa36760b57a8e35c))
* rewrite using tview ([914e124](https://github.com/taylorskalyo/goreader/commit/914e124d8c9e0cb8873c941b90d2608043ffebf9))
* save and persist last read chapter state ([#19](https://github.com/taylorskalyo/goreader/issues/19)) ([8f8e8d0](https://github.com/taylorskalyo/goreader/commit/8f8e8d0187e6fee1bea2a3e6eb55815c61d7a4e4))
* save reading progress within a chapter ([8ca4821](https://github.com/taylorskalyo/goreader/commit/8ca482125ea5badd74740105ad927a202231d1f1))


### Bug Fixes

* always render image before alt text ([961379d](https://github.com/taylorskalyo/goreader/commit/961379dcafcdefd045ae85f1164d7fb679e65440))
* handle relative image src ([9927a46](https://github.com/taylorskalyo/goreader/commit/9927a4685923eb8fc06d451cf4776d5ef6268553))
* handle special keys (e.g. ctrl-c) ([1cb9648](https://github.com/taylorskalyo/goreader/commit/1cb964873b71aaa634f6bf0b95338298a31fdd4f))
* make page stats more consistent ([3ca3f54](https://github.com/taylorskalyo/goreader/commit/3ca3f5458877663b66b0f54f8aa294afac841017))
* re-implement usage and help messages ([13b0924](https://github.com/taylorskalyo/goreader/commit/13b09247062249075cf548faa1d789a2314b7923))
* wait for spawned goroutines in tests ([6b105e4](https://github.com/taylorskalyo/goreader/commit/6b105e4d92afcade063818f2b16bae9854abf9de))


### Miscellaneous Chores

* remove Left / Right actions ([4fa57b8](https://github.com/taylorskalyo/goreader/commit/4fa57b8ab5270a061c37572a464eca7d4d898d4f))

## [1.0.1](https://github.com/taylorskalyo/goreader/compare/v1.0.0...v1.0.1) (2025-02-07)


### Bug Fixes

* replace terbox-go with tcell ([d51af20](https://github.com/taylorskalyo/goreader/commit/d51af202d5ad66749008948a5f14ca10d8712bf1)), closes [#15](https://github.com/taylorskalyo/goreader/issues/15)

## 1.0.0 (2025-02-07)


### Bug Fixes

* check error returned by termbox.Clear ([85f73c0](https://github.com/taylorskalyo/goreader/commit/85f73c0858c9579d66f3fd181597ad0f87c831d0))
