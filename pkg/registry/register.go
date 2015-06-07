package registry

import (
	"errors"
	"fmt"
	"net"
	"path"
	"strconv"
	"strings"

	"github.com/coreos/go-etcd/etcd"
	"github.com/twanies/flow/api"
)

const (
	root         string = "/flow"
	servicePath  string = "/services"
	endpointPath string = "/endpoints"

	// keyspace where services register themself, telling the registry there are
	// new, updated or deleted
	serviceWatchPath string = root + "/register" + "/service"
)

type Register interface {
	CreateService(service *api.Service) (*api.Service, error)
	GetService(key string) (*api.Service, error)
	GetServices() ([]api.Service, error)
	CreateEndpoints(endpoints *api.Endpoints) (*api.Endpoints, error)
	GetServiceEndpoints(name string) (*api.Endpoints, error)
	GetEndpoints() ([]api.Endpoints, error)
	WatchServices(services chan []api.Service)
}

type Registry struct {
	client *etcd.Client
}

func NewRegistry() *Registry {
	machines := []string{"http://127.0.0.1:2379"}
	client := etcd.NewClient(machines)
	return &Registry{client}
}

// CreateService stores a new service to the registry
func (r *Registry) CreateService(service *api.Service) (*api.Service, error) {
	kvList := keyValList{}
	kvList.add("protocol", service.Protocol)
	kvList.add("name", service.Name)
	for _, kv := range kvList.list {
		key := path.Join(makeEtcdServiceKey(service.Name), kv.key)
		if err := r.setKey(key, kv.value); err != nil {
			return nil, err
		}
	}
	// let the watchers know the service is successfully created
	if err := r.setKey("/flow/register/service", service.Name); err != nil {
		return nil, err
	}
	return service, nil
}

// GetService retrieves a service from the registry
func (r *Registry) GetService(key string) (*api.Service, error) {
	kvals, err := r.getKeyValues(key)
	if err != nil {
		return nil, err
	}
	service := &api.Service{
		Name:     kvals.get(key, "name"),
		Protocol: kvals.get(key, "protocol"),
	}
	return service, nil
}

func (r *Registry) GetServices() ([]api.Service, error) {
	services := make([]api.Service, 0)
	keys, err := r.getDirKeys(root, servicePath)
	if err != nil {
		return services, err
	}
	for _, key := range keys {
		service, err := r.GetService(key)
		if err != nil {
			return services, err
		}
		services = append(services, *service)
	}
	return services, nil
}

// CreateEndpoints stores endpoints implemented by a service
// endpoints are stored like "/flow/endpoints/{name}/host:port"
func (r *Registry) CreateEndpoints(endpoints *api.Endpoints) (*api.Endpoints, error) {
	for _, endpoint := range endpoints.Subset {
		hostPort := fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
		endpointKeyspace := path.Join(root, endpointPath, endpoints.Name, hostPort)
		if err := r.createDir(endpointKeyspace); err != nil {
			return nil, err
		}
	}
	return endpoints, nil
}

// GetEndpoints retrieves all endpoints stored in the registry
func (r *Registry) GetEndpoints() ([]api.Endpoints, error) {
	var allEndpoints []api.Endpoints
	keys, err := r.getDirKeys(root, endpointPath)
	if err != nil {
		return allEndpoints, err
	}
	for _, key := range keys {
		endpoints, err := r.GetServiceEndpoints(key)
		if err != nil {
			return allEndpoints, err
		}
		allEndpoints = append(allEndpoints, *endpoints)
	}
	return allEndpoints, nil
}

// GetServiceEndpoints retrieves the endpoints from a service by its keyspace
func (r *Registry) GetServiceEndpoints(key string) (*api.Endpoints, error) {
	keys, err := r.getDirKeys(key)
	if err != nil {
		return nil, err
	}
	name := strings.TrimPrefix(key, path.Join(root, endpointPath))
	name = strings.TrimPrefix(name, "/")
	var subset []api.Endpoint
	for _, key := range keys {
		host, port := extractEndpointFromKey(key)
		endpoint := api.Endpoint{host, port}
		subset = append(subset, endpoint)
	}
	endpoints := &api.Endpoints{
		Name:   name,
		Subset: subset,
	}
	return endpoints, nil
}

func (r *Registry) WatchServices(servicesch chan []api.Service) {
	resp := make(chan *etcd.Response)
	go r.client.Watch(serviceWatchPath, 0, true, resp, nil)
	for true {
		<-resp
		services, err := r.GetServices()
		if err != nil {
			panic(err)
		}
		servicesch <- services
	}
}

// extracts the "host:port" string from a full endpoint keyspace
// "/flow/endpoints/{name}/1.1:3000"
func extractEndpointFromKey(key string) (string, int) {
	pathTrimmed := strings.TrimPrefix(key, path.Join(root, endpointPath))
	pathTrimmed = strings.TrimPrefix(pathTrimmed, "/")
	parts := strings.Split(pathTrimmed, "/")
	hostPort := parts[1]
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		panic(err)
	}
	port, _ := strconv.Atoi(portStr)
	return host, port
}

// etcd helper methods
func (r *Registry) setKey(key, val string) error {
	_, err := r.client.Set(key, val, 0)
	if err != nil {
		return err
	}
	return nil
}

func (r *Registry) getValue(keys ...string) (string, error) {
	resp, err := r.client.Get(strings.Join(keys, "/"), false, false)
	if err != nil {
		return "", err
	}
	if isDir(resp.Node) {
		return "", errors.New("key not found")
	}
	return resp.Node.Value, nil
}

// getDirKeys returns all the keys that are etcd directories
func (r *Registry) getDirKeys(keys ...string) ([]string, error) {
	var out []string
	resp, err := r.client.Get(strings.Join(keys, "/"), false, true)
	if err != nil {
		return out, err
	}
	for _, node := range resp.Node.Nodes {
		if isDir(node) {
			out = append(out, node.Key)
		}
	}
	return out, nil
}

func (r *Registry) createDir(key string) error {
	_, err := r.client.CreateDir(key, 0)
	if err != nil {
		return err
	}
	return nil
}

type keyValPair struct {
	key   string
	value string
}

type keyValList struct {
	list []keyValPair
}

func (kvList *keyValList) get(keys ...string) string {
	for _, kv := range kvList.list {
		if kv.key == strings.Join(keys, "/") {
			return kv.value
		}
	}
	return ""
}

func (kvList *keyValList) add(key, value string) {
	kvList.list = append(kvList.list, keyValPair{key, value})
}

func (r *Registry) getKeyValues(keys ...string) (keyValList, error) {
	var kvList keyValList
	resp, err := r.client.Get(strings.Join(keys, "/"), false, true)
	if err != nil {
		return kvList, err
	}
	for _, node := range resp.Node.Nodes {
		if !isDir(node) {
			kvList.add(node.Key, node.Value)
		}
	}
	return kvList, nil
}

func (r *Registry) deleteKey(keys ...string) error {
	_, err := r.client.Delete(strings.Join(keys, "/"), true)
	if err != nil {
		return err
	}
	return nil
}

func isDir(node *etcd.Node) bool {
	return node.Dir && node != nil
}

func makeEtcdServiceKey(name string) string {
	return path.Join(root, servicePath, name)
}
