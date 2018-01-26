package api

import (
	. "github.com/tomtom5152/dnsyo/dnsyo"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"net/http"
)

type APIServer struct {
	Servers ServerList
	r chi.Router
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}


func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

func NewAPIServer(servers ServerList) (api APIServer) {
	api.Servers = servers
	api.r = chi.NewRouter()

	api.r.Use(middleware.RequestID)
	api.r.Use(middleware.Logger)
	api.r.Use(middleware.Recoverer)
	api.r.Use(middleware.RedirectSlashes)
	//api.r.Use(middleware.URLFormat)
	api.r.Use(render.SetContentType(render.ContentTypeJSON))

	api.r.Route("/v1", func(r chi.Router) {
		r.Get("/query/{domain:.*}", api.QueryHandler)
	})

	return
}

func (api *APIServer) Run(port string) {
	http.ListenAndServe(port, api.r)
}
