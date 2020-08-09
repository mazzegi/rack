package rack

import (
	"net/http"

	"github.com/gorilla/mux"
)

func ExtractVar(name string, r *http.Request) string {
	if v, ok := mux.Vars(r)[name]; ok {
		return v
	}
	return ""
}
