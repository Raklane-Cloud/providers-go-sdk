package idcloudhost

import (
	"context"
	"testing"
)

func TestFloatingIPs(t *testing.T) {
	c, seen := fakeAPI(t, map[string]func() (int, string){
		"GET /jkt01/network/ip_addresses":            ok(`[{"address":"1.1.1.1","assigned_to":"vm-1"},{"address":"2.2.2.2","assigned_to":null}]`),
		"DELETE /jkt01/network/ip_addresses/1.1.1.1": ok(``),
		"POST /jkt01/network/ip_addresses":           ok(`{"address":"3.3.3.3"}`),
	})
	ctx := context.Background()
	ips, err := c.ListFloatingIPs(ctx, ListFloatingIPsOptions{VMUUID: "vm-1"})
	if err != nil || len(ips) != 2 || ips[0].AssignedTo != "vm-1" || ips[1].AssignedTo != "" {
		t.Fatalf("ListFloatingIPs = %+v, %v", ips, err)
	}
	if (*seen)[0].Query.Get("vm_uuid") != "vm-1" {
		t.Errorf("query = %v", (*seen)[0].Query)
	}
	if err := c.DeleteFloatingIP(ctx, "1.1.1.1"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.CreateFloatingIP(ctx, "edge", 0); err != nil {
		t.Fatal(err)
	}
	if body := (*seen)[2].JSON; body["billing_account_id"] != float64(6) || body["name"] != "edge" {
		t.Errorf("create body = %v", body)
	}
}
