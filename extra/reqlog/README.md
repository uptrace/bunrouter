# Request logging middleware for bunrouter

To install:

```bash
go get github.com/uptrace/bunrouter/extra/reqlog
```

To use:

```go
import "github.com/uptrace/bunrouter/extra/reqlog"

router := bunrouter.New(
    bunrouter.WithMiddleware(reqlog.NewMiddleware()),
)
```

With options:

```go
import "github.com/uptrace/bunrouter/extra/reqlog"

reqlogMiddleware := reqlog.NewMiddleware(reqlog.WithVerbose(false))

router := bunrouter.New(
    bunrouter.WithMiddleware(reqlogMiddleware),
)
```
