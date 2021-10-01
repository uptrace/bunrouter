# Gzip compression middleware for bunrouter

To install:

```bash
go get github.com/uptrace/bunrouter/extra/bunroutergzip
```

To use:

```go
import "github.com/uptrace/bunrouter/extra/bunroutergzip"

router := bunrouter.New(
    // Compress everything with default compression level.
    bunrouter.WithMiddleware(bunroutergzip.NewMiddleware()),
)
```

With options:

```go
import (
    "github.com/klauspost/compress/gzip"
    "github.com/uptrace/bunrouter/extra/bunroutergzip"
)

gzipMiddleware := bunroutergzip.NewMiddleware(
    bunroutergzip.WithCompressionLevel(gzip.BestSpeed),
    bunroutergzip.WithContentTypes("application/json"),
)

router := bunrouter.New(
    bunrouter.WithMiddleware(gzipMiddleware),
)
```
