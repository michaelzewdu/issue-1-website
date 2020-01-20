package web

import (
	"net/http"
)

// getFront returns a handler for GET / requests.
func getFront(s *Setup) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s.templates.ExecuteTemplate(w, "front.layout", nil)
	}
}
