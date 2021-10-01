# OpenTelemetry integration for bunrouter

To install:

```bash
go get github.com/uptrace/bunrouter/extra/bunrouterotel
```

To use:

```go
import "github.com/uptrace/bunrouter/extra/bunrouterotel"

router := bunrouter.New(
    bunrouter.WithMiddleware(bunrouterotel.NewMiddleware()),
)
```

With options:

```go
import "github.com/uptrace/bunrouter/extra/bunrouterotel"

otelMiddleware := bunrouterotel.NewMiddleware(
    bunrouterotel.WithClientIP(false),
)

router := bunrouter.New(
    bunrouter.WithMiddleware(otelMiddleware),
)
```
