# Gzip compression middleware for treemux

To install:

```bash
go get github.com/uptrace/treemux/extra/treemuxgzip
```

To use:

```go
import "github.com/uptrace/treemux/extra/treemuxgzip"

router := treemux.New(
    // Compress everything with default compression level.
    treemux.WithMiddleware(treemuxgzip.NewMiddleware()),
)
```

With options:

```go
import (
    "github.com/klauspost/compress/gzip"
    "github.com/uptrace/treemux/extra/treemuxgzip"
)

gzipMiddleware := treemuxgzip.NewMiddleware(
    treemuxgzip.WithCompressionLevel(gzip.BestSpeed),
    treemuxgzip.WithContentTypes("application/json"),
)

router := treemux.New(
    treemux.WithMiddleware(gzipMiddleware),
)
```
