# Cross-Origin Resource Sharing for bunrouter

To run this example:

```go
go run main.go
```

To test CORS:

```shell
curl -H "Origin: http://localhost:9999" \
  -H "Access-Control-Request-Method: POST" \
  -X OPTIONS \
  --verbose \
  http://localhost:9999/api/v1/users/123

curl -H "Origin: http://localhost:9999" \
  -H "Access-Control-Request-Method: POST" \
  -X OPTIONS \
  --verbose \
  http://localhost:9999/api/v2/users/123
```

See [documentation](https://bunrouter.uptrace.dev/) for details.
