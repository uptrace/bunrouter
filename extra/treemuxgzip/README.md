# Gzip compression middleware for treemux

To install:

```bash
go get github.com/vmihailenco/treemux/extra/treemuxgzip
```

To use:

```go
import "github.com/vmihailenco/treemux/extra/treemuxgzip"

router := treemux.New()
// Compress everything with default compression level.
router.Use(treemuxgzip.Middleware)
```

With options:

```go
import (
    "github.com/klauspost/compress/gzip"
    "github.com/vmihailenco/treemux/extra/treemuxgzip"
)

router := treemux.New()
router.Use(treemuxgzip.New(
    treemuxgzip.WithCompressionLevel(gzip.BestSpeed),
    treemuxgzip.WithContentTypes("application/json"),
).Middleware)
```
