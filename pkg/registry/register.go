package registry

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/go-etcd/etcd"

	"github.com/twanies/flow/api"
)

const (
	namespace  = "/flow"
	serviceKey = "services"
)

type ServiceNameKey struct {
	Name string
	Key  string
}

type Register interface {
	GetService(key string) (*api.Service, error)
	ListServices() (api.ServiceList, error)
	CreateService(svc *api.Service) (*api.Service, error)
	DeleteService(serviceName string) error
}

type Registry struct {
	client *etcd.Client
}

func NewRegistry() *Registry {
	machines := []string{"http://127.0.0.1:2379"}
	client := etcd.NewClient(machines)
	return &Registry{client}
}

func (r *Registry) CreateService(svc *api.Service) (*api.Service, error) {
	key := MakeServiceKey(svc.Name)

	var err error
	err = r.setKey(join(key, "protocol"), svc.Protocol)
	err = r.setKey(join(key, "name"), svc.Name)
	err = r.setKey(join(key, "frontend", "scheme"), svc.Frontend.Scheme)
	err = r.setKey(join(key, "frontend", "route"), svc.Frontend.Route)
	err = r.setKey(join(key, "frontend", "targetpath"), svc.Frontend.TargetPath)
	if err != nil {
		return nil, err
	}
	if err = r.createNodes(key, svc.Nodes); err != nil {
		return nil, err
	}
	return svc, nil
}

// Delete a service by its name recursively
func (r *Registry) DeleteService(serviceName string) error {
	key := MakeServiceKey(serviceName)
	return r.deleteKey(key, true)
}

// Stores a single node in the registry
func (r *Registry) createNode(key string, node *api.Node) error {
	if err := r.setKey(join(key, "nodes", node.String(), "host"), node.Host); err != nil {
		return err
	}
	port := fmt.Sprintf("%d", node.Port)
	if err := r.setKey(join(key, "nodes", node.String(), "port"), port); err != nil {
		return err
	}
	return nil
}

// Stores multiple nodes in the registry
func (r *Registry) createNodes(key string, nodes []api.Node) error {
	for _, node := range nodes {
		if err := r.createNode(key, &node); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) GetService(key string) (*api.Service, error) {
	kvals, err := r.getKeyVals(key)
	if err != nil {
		return nil, err
	}
	frontend, err := r.getKeyVals(key, "frontend")
	if err != nil {
		return nil, err
	}
	nodes, err := r.getNodes(key)
	if err != nil {
		return nil, err
	}
	svc := &api.Service{
		Name:     kvals.get(key, "name"),
		Protocol: kvals.get(key, "protocol"),
		Frontend: api.FrontendMeta{
			Scheme:     frontend.get(key, "frontend", "scheme"),
			Route:      frontend.get(key, "frontend", "route"),
			TargetPath: frontend.get(key, "frontend", "targetpath"),
		},
		Nodes: nodes,
	}
	return svc, nil
}

// list all stored services
func (r *Registry) ListServices() (api.ServiceList, error) {
	keys, err := r.getDirKeys(namespace, serviceKey)
	if err != nil {
		return api.ServiceList{}, err
	}

	var services api.ServiceList
	for _, key := range keys {
		service, err := r.GetService(key)
		if err != nil {
			return api.ServiceList{}, nil
		}
		services = append(services, *service)
	}
	return services, nil
}

func (r *Registry) getNode(key string) (*api.Node, error) {
	keyVals, err := r.getKeyVals(key)
	if err != nil {
		return nil, err
	}
	port, _ := strconv.Atoi(keyVals.get(key, "port"))
	node := &api.Node{
		Host: keyVals.get(key, "host"),
		Port: port,
	}
	return node, nil
}

func (r *Registry) getNodes(serviceKey string) (api.NodeList, error) {
	var nodes api.NodeList
	keys, err := r.getDirKeys(serviceKey, "nodes")
	if err != nil {
		return api.NodeList{}, nil
	}
	for _, key := range keys {
		node, err := r.getNode(key)
		if err != nil {
			return api.NodeList{}, err
		}
		nodes = append(nodes, *node)
	}
	return nodes, err
}

// etcd helpers
func (r *Registry) getValue(keys ...string) (string, error) {
	resp, err := r.client.Get(strings.Join(keys, "/"), false, false)
	if err != nil {
		return "", err
	}
	if isDir(resp.Node) {
		return "", errors.New("value not found")
	}
	return resp.Node.Value, nil
}

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

type keyValMap map[string]string

func (kv keyValMap) get(keys ...string) string {
	val, ok := kv[strings.Join(keys, "/")]
	if !ok {
		return ""
	}
	return val
}

func (r *Registry) getKeyVals(keys ...string) (keyValMap, error) {
	resp, err := r.client.Get(strings.Join(keys, "/"), false, true)
	if err != nil {
		return nil, err
	}
	kv := make(keyValMap, len(resp.Node.Nodes))
	for _, node := range resp.Node.Nodes {
		if !isDir(node) {
			kv[node.Key] = node.Value
		}
	}
	return kv, nil
}

func (r *Registry) setKey(key, val string) error {
	_, err := r.client.Set(key, val, 0)
	return err
}

func (r *Registry) deleteKey(key string, rf bool) error {
	_, err := r.client.Delete(key, rf)
	if err != nil {
		return err
	}
	return nil
}

func (r *Registry) createDir(keys ...string) error {
	_, err := r.client.CreateDir(strings.Join(keys, "/"), 0)
	if err != nil {
		return err
	}
	return nil
}

// helpers
func isDir(n *etcd.Node) bool {
	return n.Dir && n != nil
}

func join(keys ...string) string {
	return strings.Join(keys, "/")
}

func MakeServiceKey(name string) string {
	return join(namespace, serviceKey, name)
}
