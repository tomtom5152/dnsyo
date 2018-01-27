package api

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/tomtom5152/dnsyo/dnsyo"
	"net/http"
)

// Server is responsible for management of the working ServerList and storage of the router
type Server struct {
	Servers dnsyo.ServerList
	r       chi.Router
}

type errResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *errResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// errInvalidRequest produces a chi/render object representing an invalid request
func errInvalidRequest(err error) render.Renderer {
	return &errResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

// errRender produces a chi/render object representing an error whilst rendering
func errRender(err error) render.Renderer {
	return &errResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

// NewAPIServer sets up the routes an produces a new Server instance with the router assigned.
func NewAPIServer(servers dnsyo.ServerList) (api Server) {
	api.Servers = servers
	api.r = chi.NewRouter()

	api.r.Use(middleware.RequestID)
	api.r.Use(middleware.Logger)
	api.r.Use(middleware.Recoverer)
	api.r.Use(middleware.RedirectSlashes)
	//api.r.Use(middleware.URLFormat)
	api.r.Use(render.SetContentType(render.ContentTypeJSON))

	api.r.Route("/v1", func(r chi.Router) {
		r.Get("/query/{domain:.*}", api.queryHandler)
	})

	return
}

// Run starts the Server on the given port. Port can should take the form :<port> or <ip>:<port>.
func (api *Server) Run(port string) {
	http.ListenAndServe(port, api.r)
}
