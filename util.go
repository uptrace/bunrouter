package treemux

import (
	"encoding/json"
	"net/http"
)

type H map[string]interface{}

// JSON marshals the value as JSON and writes it to the response writer.
// Don't hesitate to copy-paste this function to your project and customize it as necessary.
func JSON(w http.ResponseWriter, value H) error {
	if value == nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(value); err != nil {
		return err
	}

	return nil
}

func defaultErrorHandler(w http.ResponseWriter, req Request, err error) {
	w.WriteHeader(http.StatusBadRequest)
	_ = JSON(w, H{
		"message": err.Error(),
	})
}
