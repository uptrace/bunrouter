package httptreemux

import (
	"net/http"
)

func SimplePanicHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
}

// This was taken from github.com/gocraft/web, which adapted it from the Traffic project.
