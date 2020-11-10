# OpenTelemetry integration for treemux

To install:

```bash
go get github.com/vmihailenco/treemux/extra/treemuxotel
```

To use:

```go
import "github.com/vmihailenco/treemux/extra/treemuxotel"

router := treemux.New()
router.Use(treemuxotel.Middleware)
```

With options:

```go
import "github.com/vmihailenco/treemux/extra/treemuxotel"

router := treemux.New()
router.Use(treemuxotel.New(treemuxotel.WithClientIP(false)).Middleware)
```
