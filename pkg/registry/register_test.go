package registry

import (
	"reflect"
	"testing"
)

func TestSetKeyGetValue(t *testing.T) {
	r := NewRegistry()
	defer tearDown(t, r)
	if err := r.setKey("/flowtest/foo", "bar"); err != nil {
		t.Fatal(err)
	}
	val, err := r.getValue("/flowtest/foo")
	if err != nil {
		t.Fatal(err)
	}
	if val != "bar" {
		t.Fatalf("exptected value (bar) got %s", val)
	}
}

func TestGetdirKeys(t *testing.T) {
	r := NewRegistry()
	defer tearDown(t, r)
	if err := r.createDir("/flowtest/foo"); err != nil {
		t.Fatal(err)
	}
	keys, err := r.getDirKeys("/flowtest")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(keys)
	if len(keys) > 1 {
		t.Fatal("lenght of keys cannot be greater then 1")
	}
	if keys[0] != "/flowtest/foo" {
		t.Fatalf("expected (/flowtest/foo) got %s", keys[0])
	}
}

func TestKeyValList(t *testing.T) {
	r := NewRegistry()
	defer tearDown(t, r)
	kvList := keyValList{}
	kvList.add("/flowtest/foo", "bar")
	kvList.add("/flowtest/moo", "mar")
	for _, kv := range kvList.list {
		if err := r.setKey(kv.key, kv.value); err != nil {
			t.Fatal(err)
		}
	}
	newKvList, err := r.getKeyValues("/flowtest")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(kvList, newKvList) {
		t.Fatal("expected equal lists")
	}
}

func TestExtractEndpointFromKey(t *testing.T) {
	key := "/flow/endpoints/api/1.1:3000"
	host, port := extractEndpointFromKey(key)
	if host != "1.1" {
		t.Fatal("expected 1.1 got %s", host)
	}
	if port != 3000 {
		t.Fatal("expected 3000 got %d", port)
	}
}

func tearDown(t *testing.T, r *Registry) {
	if err := r.deleteKey("/flowtest"); err != nil {
		t.Fatal(err)
	}
	r.client.Close()
}
