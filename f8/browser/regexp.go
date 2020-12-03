package browser

import (
	"net/http"
	"regexp"
)

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

// RegexpHandler ...
type RegexpHandler struct {
	routes []*route
}

// Handle ...
func (h *RegexpHandler) Handle(pattern string, handler http.Handler) {
	h.routes = append(h.routes, &route{regexp.MustCompile(pattern), handler})
}

// HandleFunc registers the handler function for the given pattern in the DefaultServeMux. The documentation for ServeMux explains how patterns are matched.
func (h *RegexpHandler) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	h.routes = append(h.routes, &route{regexp.MustCompile(pattern), http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	// no pattern matched; send 404 response
	http.NotFound(w, r)
}
