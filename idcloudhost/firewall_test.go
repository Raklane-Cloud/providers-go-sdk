package idcloudhost

import (
	"context"
	"testing"
)

func TestFirewalls(t *testing.T) {
	c, seen := fakeAPI(t, map[string]func() (int, string){
		"POST /jkt01/network/firewalls":          ok(`{"uuid":"fw-1","display_name":"workers","rules":[]}`),
		"POST /jkt01/network/firewalls/fw-1/vms": ok(`[{"resource_type":"vm","resource_uuid":"vm-1"}]`),
	})
	ctx := context.Background()
	fw, err := c.CreateFirewall(ctx, CreateFirewallRequest{
		DisplayName: "workers",
		Rules:       []FirewallRule{{Protocol: "tcp", Direction: DirectionInbound, PortStart: Port(443), EndpointSpecType: EndpointAny}},
	})
	if err != nil || fw.FirewallName() != "workers" {
		t.Fatalf("CreateFirewall = %+v, %v", fw, err)
	}
	rules, _ := (*seen)[0].JSON["rules"].([]any)
	if len(rules) != 1 || rules[0].(map[string]any)["port_start"] != float64(443) {
		t.Errorf("rules = %v", (*seen)[0].JSON["rules"])
	}
	if err := c.AssignFirewall(ctx, "fw-1", "vm-1"); err != nil {
		t.Fatal(err)
	}
	if (*seen)[1].Query.Get("vm_uuid") != "vm-1" {
		t.Errorf("assign query = %v", (*seen)[1].Query)
	}
}
