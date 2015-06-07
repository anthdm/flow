package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/twanies/flow/api"
	"github.com/twanies/flow/pkg/registry"
)

const apiVersion string = "v0.0.1"

type Server struct {
	srv      *http.Server
	router   *mux.Router
	l        net.Listener
	registry registry.Register
}

func NewServer(addr string) *Server {
	registry := registry.NewRegistry()
	s := &Server{registry: registry}
	r := createRouter(s)
	s.srv = &http.Server{Addr: addr, Handler: r}
	s.router = r
	return s
}

func (s *Server) Serve() error {
	return s.srv.Serve(s.l)
}

func (s *Server) Close() error {
	return s.l.Close()
}

// TODO: add a stop chan so we can track when the server stops
func (s *Server) ServeAPI() error {
	var err error
	s.l, err = net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}
	go func() {
		log.Printf("api available on http://localhost%s", s.srv.Addr)
		if err := s.Serve(); err != nil {
			log.Println(err)
		}
	}()
	return nil
}

type httpApifunc func(w http.ResponseWriter, r *http.Request, vars map[string]string) error

func (s *Server) getVersion(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	version := api.Version{
		Version:    "0.0.1",
		ApiVersion: "0.0.1",
		GitCommit:  "sifeifh84848",
	}
	return writeJSON(w, http.StatusOK, version)
}

func (s *Server) postCreateService(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return errors.New("missing params")
	}
	svc := api.Service{}
	if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
		return fmt.Errorf("failed to decode the response body: %v", err)
	}
	defer r.Body.Close()
	service, err := s.registry.CreateService(&svc)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, service)
}

func (s *Server) getService(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return errors.New("missin params")
	}
	name := vars["name"]
	service, err := s.registry.GetService("/flow/services/" + name)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, service)
}

func (s *Server) postCreateEndpoints(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return errors.New("missing params")
	}
	endpoints := api.Endpoints{}
	if err := json.NewDecoder(r.Body).Decode(&endpoints); err != nil {
		return fmt.Errorf("failed to decode the response body: %v", err)
	}
	out, err := s.registry.CreateEndpoints(&endpoints)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, out)
}

func (s *Server) getServiceEndpoints(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	name := vars["name"]
	keyspace := "/flow/endpoints/" + name
	endpoints, err := s.registry.GetServiceEndpoints(keyspace)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, endpoints)
}

func (s *Server) getListEndpoints(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	allEndpoints, err := s.registry.GetEndpoints()
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, allEndpoints)
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

func createRouter(s *Server) *mux.Router {
	r := mux.NewRouter()
	m := map[string]map[string]httpApifunc{
		"GET": {
			"/service/{name}":   s.getService,
			"/endpoints/{name}": s.getServiceEndpoints,
			"/endpoints":        s.getListEndpoints,
		},
		"POST": {
			"/service":   s.postCreateService,
			"/endpoints": s.postCreateEndpoints,
		},
		"DELETE": {},
	}
	for method, routes := range m {
		for route, handler := range routes {
			f := makeHttpHandler(handler)
			r.Path("/v{version:[0-9.]+}" + route).Methods(method).Handler(f)
		}
	}
	return r
}

// Looks if the request content type is application/json
func isReqJson(r *http.Request) error {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		if r.Body == nil || r.ContentLength == 0 {
			return nil
		}
	}
	if ct != "application/json" {
		return fmt.Errorf("content type (%s) must be application/json", ct)
	}
	return nil
}

func httpError(w http.ResponseWriter, err error) {
	if err == nil || w == nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}
	statusCode := http.StatusInternalServerError
	for description, status := range map[string]int{
		"no results found": http.StatusNotFound,
		"not authorized":   http.StatusForbidden,
		"wrong parameter":  http.StatusBadRequest,
	} {
		if strings.Contains(err.Error(), description) {
			statusCode = status
		}
	}
	http.Error(w, err.Error(), statusCode)
	return
}

func makeHttpHandler(h httpApifunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if r.Method == "POST" {
			if err := isReqJson(r); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		if err := h(w, r, vars); err != nil {
			httpError(w, err)
			return
		}
	}
}
