# Request logging middleware for treemux

To install:

```bash
go get github.com/vmihailenco/treemux/extra/reqlog
```

To use:

```go
import "github.com/vmihailenco/treemux/extra/reqlog"

router := treemux.New()
router.Use(reqlog.Middleware)
```

With options:

```go
import "github.com/vmihailenco/treemux/extra/reqlog"

router := treemux.New()
router.Use(reqlog.New(reqlog.WithVerbose(false)).Middleware)
```
