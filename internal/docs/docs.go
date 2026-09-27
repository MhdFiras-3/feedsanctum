package docs

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var spec []byte

func HandleSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/yaml")
	w.Write(spec)
}
