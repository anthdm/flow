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
	if err := isReqJson(r); err != nil {
		return err
	}

	service := api.Service{}
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		return fmt.Errorf("failed to decode the response body")
	}
	out, err := s.registry.CreateService(&service)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, out)
}

func (s *Server) getService(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	name := vars["svc"]
	if name == "" {
		return errors.New("missing parameter (name)")
	}
	serviceKey := registry.MakeServiceKey(name)
	svc, err := s.registry.GetService(serviceKey)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, svc)
}

func (s *Server) deleteService(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	name := vars["svc"]
	if name == "" {
		return errors.New("missing parameter (name)")
	}
	if err := s.registry.DeleteService(name); err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, name+"deleted")
}

func (s *Server) getListServices(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	services, err := s.registry.ListServices()
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, services)
}

func (s *Server) postAddNode(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	return nil
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
			"/version":       s.getVersion,
			"/service/{svc}": s.getService,
			"/service":       s.getListServices,
		},
		"POST": {
			"/service":                s.postCreateService,
			"/service/{svc}/add_node": s.postAddNode,
		},
		"DELETE": {
			"/service/{svc}": s.deleteService,
		},
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
		if err := h(w, r, vars); err != nil {
			httpError(w, err)
			return
		}
	}
}
