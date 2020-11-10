# OpenTelemetry integration for treemux

To install:

```bash
go get github.com/vmihailenco/treemux/extra/treemuxotel
```

To use:

```go
import "github.com/vmihailenco/treemux/extra/treemuxotel"

router := treemux.New(
    treemux.WithMiddleware(treemuxotel.NewMiddleware()),
)
```

With options:

```go
import "github.com/vmihailenco/treemux/extra/treemuxotel"

otelMiddleware := treemuxotel.NewMiddleware(
    treemuxotel.WithClientIP(false),
)

router := treemux.New(
    treemux.WithMiddleware(otelMiddleware),
)
```
