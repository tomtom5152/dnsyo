package api

import (
	"net/http"
	"github.com/go-chi/chi"
	. "github.com/tomtom5152/dnsyo/dnsyo"
	"strings"
	"strconv"
	"github.com/go-chi/render"
	"errors"
)

const (
	apiQueryThreads = 200
	maxServers = 500
	defaultServers = 200
)

func (api *APIServer) QueryHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	q := &Query{
		Domain: chi.URLParam(r, "domain"),
	}
	var sl ServerList
	sl = api.Servers

	// check if the user has specified a query type
	if t := r.FormValue("t"); t != "" {
		q.Type = strings.ToUpper(t)
	} else if t := r.FormValue("type"); t != "" {
		q.Type = strings.ToUpper(t)
	} else {
		q.Type = "A"
	}

	// check if we have a country specified, apply the result
	var country string
	if c := r.FormValue("c"); c != "" {
		country = c
	} else if c := r.FormValue("country"); c != "" {
		country = c
	}
	if country != "" {
		sl, err = sl.FilterCountry(country)
		if err != nil {
			render.Render(w,r, ErrInvalidRequest(err))
			return
		}
	}


	// check if we have a number of servers specified, bound and apply the result
	numServers := 0
	if n, _ := strconv.Atoi(r.FormValue("q")); n != 0 {
		numServers = n
	} else if n, _ := strconv.Atoi(r.FormValue("servers")); n != 0 {
		numServers = n
	}

	if numServers == 0 {
		if len(sl) < defaultServers {
			numServers = len(sl)
		} else {
			numServers = defaultServers
		}
	}
	if numServers > maxServers {
		render.Render(w, r, ErrInvalidRequest(errors.New("requested too many servers to query")))
		return
	}

	sl, err = sl.NRandom(numServers)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	q.Results = sl.ExecuteQuery(q, apiQueryThreads)

	render.JSON(w, r, q.Results)
	return
}
