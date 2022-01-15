# Cross-Origin Resource Sharing for bunrouter

To run this example:

```go
go run main.go
```

To test that CORS are allowed from localhost:

```shell
curl -H "Origin: http://localhost:9999" \
  -H "Access-Control-Request-Method: POST" \
  -X OPTIONS \
  --verbose \
  http://localhost:9999/api/v1/users/123
```

To test that CORS are restricted from other domains:

```shell
curl -H "Origin: https://uptrace.dev" \
  -H "Access-Control-Request-Method: POST" \
  -X OPTIONS \
  --verbose \
  http://localhost:9999/api/v1/users/123
```

See [documentation](https://bunrouter.uptrace.dev/) for details.
