package biznetgio

import (
	"context"
	"testing"
)

func TestImportKeypair(t *testing.T) {
	c, seen := fakeAPI(t, map[string]func() (int, string){
		"POST /neolites/keypairs/import": ok(`{"success":true,"data":{"neosshkey_id":"42","name":"raklane"}}`),
	})
	k, err := c.ImportKeypair(context.Background(), "raklane", "ssh-ed25519 AAAA")
	if err != nil {
		t.Fatal(err)
	}
	if k.KeypairID != 42 {
		t.Fatalf("keypair = %+v", k)
	}
	if (*seen)[0].JSON["public_key"] != "ssh-ed25519 AAAA" {
		t.Errorf("body = %v", (*seen)[0].JSON)
	}
}

func TestListKeypairs(t *testing.T) {
	c, _ := fakeAPI(t, map[string]func() (int, string){
		"GET /neolites/keypairs/": ok(`{"success":true,"data":[{"keypair_id":1,"name":"a"},{"neosshkey_id":2,"name":"b"}]}`),
	})
	keys, err := c.ListKeypairs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 || keys[0].KeypairID != 1 || keys[1].KeypairID != 2 {
		t.Fatalf("keys = %+v", keys)
	}
}

func TestListKeypairsNullData(t *testing.T) {
	c, _ := fakeAPI(t, map[string]func() (int, string){
		"GET /neolites/keypairs/": ok(`{"success":true,"code":200,"data":null}`),
	})
	keys, err := c.ListKeypairs(context.Background())
	if err != nil || len(keys) != 0 {
		t.Fatalf("keys = %v, err = %v", keys, err)
	}
}
