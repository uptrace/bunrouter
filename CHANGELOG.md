## [1.0.20](https://github.com/uptrace/bunrouter/compare/v1.0.19...v1.0.20) (2023-02-12)



## [1.0.19](https://github.com/uptrace/bunrouter/compare/v1.0.18...v1.0.19) (2022-08-30)



## [1.0.18](https://github.com/uptrace/bunrouter/compare/v1.0.17...v1.0.18) (2022-08-09)


### Bug Fixes

* panic when trying to register same routes with different param names ([ace1f17](https://github.com/uptrace/bunrouter/commit/ace1f177c3bae9819612209e6fbb9c0496c8f58a))



## [1.0.17](https://github.com/uptrace/bunrouter/compare/v1.0.16...v1.0.17) (2022-05-25)


### Bug Fixes

* fix another corner case with incorrect param resolving ([7e3d720](https://github.com/uptrace/bunrouter/commit/7e3d7205a85cb90de5b3ddee3e2737e699fe810a))



## [1.0.16](https://github.com/uptrace/bunrouter/compare/v1.0.15...v1.0.16) (2022-05-23)


### Bug Fixes

* fix extracting params from routes with wildcard and named params ([2741aad](https://github.com/uptrace/bunrouter/commit/2741aadef5f5fdf135ffa541cf8b3b8644898ae0))


### Features

* **reqlog:** add trace id ([b2d6582](https://github.com/uptrace/bunrouter/commit/b2d658268dc2d573f5e9e9010a415c6102faa1d1))



## [1.0.15](https://github.com/uptrace/bunrouter/compare/v1.0.14...v1.0.15) (2022-05-08)


### Bug Fixes

* call method not allowed handler once ([37425b7](https://github.com/uptrace/bunrouter/commit/37425b7af9f103a7f26c7f8cc9a04ddac8b805b4))



## [1.0.14](https://github.com/uptrace/bunrouter/compare/v1.0.13...v1.0.14) (2022-04-14)


### Bug Fixes

* use httpsnoop to wrap response writer ([3d86932](https://github.com/uptrace/bunrouter/commit/3d869325af1302c55fdeb8a8ff7e4b7cfd10ece0))



## [1.0.13](https://github.com/uptrace/bunrouter/compare/v1.0.12...v1.0.13) (2022-03-22)


### Bug Fixes

* dedup middlewares applied on the method not allowed handler ([2659f03](https://github.com/uptrace/bunrouter/commit/2659f039be323b8bc4901e35f20a5760e6445640))


### Features

* accept multiple middlewares in Use(fn1, fn2, fn3) ([1486279](https://github.com/uptrace/bunrouter/commit/14862790b51c418384700a5e86aa49833937a772))



## [1.0.12](https://github.com/uptrace/bunrouter/compare/v1.0.11...v1.0.12) (2022-01-19)


### Bug Fixes

* use RawPath when available ([9859bc7](https://github.com/uptrace/bunrouter/commit/9859bc722310e0af24cd8372f585318379ccbbd9))



## [1.0.11](https://github.com/uptrace/bunrouter/compare/v1.0.10...v1.0.11) (2022-01-18)

### Features

- add basicauth middleware
  ([363da1a](https://github.com/uptrace/bunrouter/commit/363da1a989d943c8bbcf7551ad1a06150f6d1f1f))

* Added `Use` function which is an alias for `WithMiddleware`.

* Updated docs to use cleaner version of the API. Instead of:

```go
router.NewGroup("/some/prefix",
	 bunrouter.WithMiddleware(middleware1),
	 bunrouter.WithMiddleware(middleware2),
	 bunrouter.WithGroup(func(group *bunrouter.Group) {}),
)
```

You can use:

```go
router.Use(middleware1).
    Use(middleware2).
    WithGroup("/some/prefix", func(group *bunrouter.Group) {})
```

## [1.0.10](https://github.com/uptrace/bunrouter/compare/v1.0.9...v1.0.10) (2022-01-08)

- harden redir checks

## [1.0.9](https://github.com/uptrace/bunrouter/compare/v1.0.8...v1.0.9) (2021-12-23)

### Bug Fixes

- properly handle wildcard node without a slash
  ([88b4d3e](https://github.com/uptrace/bunrouter/commit/88b4d3ea352c92fc7a87972fc95add8e7f99c328))

### Features

- add Router.ServeHTTPError that returns the error from the handler
  ([9add167](https://github.com/uptrace/bunrouter/commit/9add167b91c37b42846a486a9965f5212d49bafa))

## [1.0.8](https://github.com/uptrace/bunrouter/compare/v1.0.7...v1.0.8) (2021-11-16)

### Bug Fixes

- properly handle empty root node
  ([b37ad45](https://github.com/uptrace/bunrouter/commit/b37ad4595c66454f4a768356298c95976e01d7f2))

## [1.0.7](https://github.com/uptrace/bunrouter/compare/v1.0.6...v1.0.7) (2021-11-12)

### Bug Fixes

- don't panic on path that matches routes common prefix
  ([d89dc38](https://github.com/uptrace/bunrouter/commit/d89dc38defc44bdf4bab13ecb518c2aa42ad9e80))

## [1.0.6](https://github.com/uptrace/bunrouter/compare/v1.0.5...v1.0.6) (2021-11-09)

### Bug Fixes

- propagate error in HTTP compat handlers
  ([5ed4d41](https://github.com/uptrace/bunrouter/commit/5ed4d41e99e8f6614753393f13e3674df29e7fb9))

## [1.0.5](https://github.com/uptrace/bunrouter/compare/v1.0.4...v1.0.5) (2021-11-08)

### Bug Fixes

- fallback to context when params can't be found
  ([ee2eb33](https://github.com/uptrace/bunrouter/commit/ee2eb3339ff421dd80566802304a32265f6e28b1))

### Features

- apply middleware to method not allowed handler
  ([8e295d0](https://github.com/uptrace/bunrouter/commit/8e295d0f01fbdf16061b7a4c53b931e9d709b25b))

## [1.0.4](https://github.com/uptrace/bunrouter/compare/v1.0.3...v1.0.4) (2021-11-07)

### Features

- **reqlog:** support http.Flusher
  ([938d70a](https://github.com/uptrace/bunrouter/commit/938d70aa4743d3c1492af8421a3fff14df986fa0))

## [1.0.3](https://github.com/uptrace/bunrouter/compare/v1.0.2...v1.0.3) (2021-10-21)

### Bug Fixes

- make routes with only colon nodes work
  ([fffd754](https://github.com/uptrace/bunrouter/commit/fffd75448f70a508254b0327c933cfda19eac70f))

## [1.0.2](https://github.com/uptrace/bunrouter/compare/v1.0.1...v1.0.2) (2021-10-19)

### Features

- make redirects work for wildcard routes
  ([04cb9f3](https://github.com/uptrace/bunrouter/commit/04cb9f3fd564d76477dcba7218e29f980503b15d))

## [1.0.1](https://github.com/uptrace/bunrouter/compare/v1.0.0...v1.0.1) (2021-10-15)

### Bug Fixes

- change WithContext to preserve route params
  ([2ca195a](https://github.com/uptrace/bunrouter/commit/2ca195ac8e7d9242d5110b84ede8d50a360f9a47))

# [1.0.0](https://github.com/uptrace/bunrouter/compare/v1.0.0-rc.2...v1.0.0) (2021-10-14)

### Bug Fixes

- make Slice and Map work on empty request
  ([609c7a3](https://github.com/uptrace/bunrouter/commit/609c7a3fcb6f5140c1def406efeee01eb0d80a11))

### Features

- allow configuring reqlog from env variables
  ([486ec10](https://github.com/uptrace/bunrouter/commit/486ec1061ec244559bb072c5b9f78858df8d9fd4))

# [1.0.0-rc.2](https://github.com/uptrace/bunrouter/compare/v1.0.0-rc.1...v1.0.0-rc.2) (2021-10-04)

- Initial release. See the [documentation](https://bunrouter.uptrace.dev/) for details.
