package idcloudhost

import (
	"context"
	"net/http"
	"testing"
)

const vmJSON = `{
	"uuid": "fc880f74-cf03-4a7a-93da-74c506157023", "name": "raklane-node-1", "status": "running",
	"vcpu": 2, "memory": 2048, "private_ipv4": "10.1.8.4", "billing_account": 6,
	"storage": [{"uuid": "d1", "name": "vda", "primary": true, "size": 20},
	            {"uuid": "d2", "name": "vdb", "primary": false, "size": 50}]
}`

func TestCreateVM(t *testing.T) {
	c, seen := fakeAPI(t, map[string]func() (int, string){"POST /jkt01/user-resource/vm": ok(vmJSON)})

	vm, err := c.CreateVM(context.Background(), CreateVMRequest{
		Name: "raklane-node-1", OSName: "ubuntu", OSVersion: "24.04", DiskGB: 40, VCPU: 2, RAMMB: 2048,
		Username: "raklane", Password: "Secret123", PublicKeys: []string{"ssh-ed25519 A", "ssh-ed25519 B"},
		NetworkUUID: "net-1", CloudInit: "#cloud-config\nruncmd: []\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if vm.UUID != "fc880f74-cf03-4a7a-93da-74c506157023" || vm.MemoryMB != 2048 || vm.PrivateIPv4 != "10.1.8.4" || len(vm.Storage) != 2 {
		t.Fatalf("decoded vm %+v", vm)
	}
	got := (*seen)[0]
	if got.APIKey != "secret" {
		t.Errorf("apikey = %q", got.APIKey)
	}
	want := map[string]string{
		"name": "raklane-node-1", "os_name": "ubuntu", "os_version": "24.04", "disks": "40", "vcpu": "2",
		"ram": "2048", "username": "raklane", "password": "Secret123", "billing_account_id": "6",
		"network_uuid": "net-1", "cloud_init": "#cloud-config\nruncmd: []\n", "reserve_public_ip": "true", "backup": "false",
	}
	for k, v := range want {
		if got.Form.Get(k) != v {
			t.Errorf("form %s = %q, want %q", k, got.Form.Get(k), v)
		}
	}
	if keys := got.Form["public_keys"]; len(keys) != 2 {
		t.Errorf("public_keys = %v, want both keys", keys)
	}
}

func TestGetVMNotFound(t *testing.T) {
	c, seen := fakeAPI(t, map[string]func() (int, string){
		"GET /jkt01/user-resource/vm": func() (int, string) {
			return http.StatusBadRequest, `{"errors":{"Error":"No such virtual machine exists. x"}}`
		},
	})
	_, err := c.GetVM(context.Background(), "x")
	if !IsNotFound(err) {
		t.Fatalf("got %v, want not found", err)
	}
	if (*seen)[0].Query.Get("uuid") != "x" {
		t.Errorf("query = %v", (*seen)[0].Query)
	}
}

func TestDeleteVMSendsFormBody(t *testing.T) {
	c, seen := fakeAPI(t, map[string]func() (int, string){"DELETE /jkt01/user-resource/vm": ok(`{"success":true}`)})
	if err := c.DeleteVM(context.Background(), "vm-1"); err != nil {
		t.Fatal(err)
	}
	if (*seen)[0].Form.Get("uuid") != "vm-1" {
		t.Errorf("form = %v", (*seen)[0].Form)
	}
}

func TestNoLocationUsesDefaultPaths(t *testing.T) {
	c, _ := fakeAPI(t, map[string]func() (int, string){"GET /user-resource/vm/list": ok(`[` + vmJSON + `]`)})
	c.Location = ""
	vms, err := c.ListVMs(context.Background())
	if err != nil || len(vms) != 1 {
		t.Fatalf("ListVMs = %v, %v", vms, err)
	}
}
