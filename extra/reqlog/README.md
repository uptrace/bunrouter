# Request logging middleware for treemux

To install:

```bash
go get github.com/uptrace/treemux/extra/reqlog
```

To use:

```go
import "github.com/uptrace/treemux/extra/reqlog"

router := treemux.New(
    treemux.WithMiddleware(reqlog.NewMiddleware()),
)
```

With options:

```go
import "github.com/uptrace/treemux/extra/reqlog"

reqlogMiddleware := reqlog.NewMiddleware(reqlog.WithVerbose(false))

router := treemux.New(
    treemux.WithMiddleware(reqlogMiddleware),
)
```
