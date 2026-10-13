package idcloudhost

import (
	"context"
	"math"
	"testing"
)

func TestPricingVMHourly(t *testing.T) {
	c, _ := fakeAPI(t, map[string]func() (int, string){"GET /pricing/policy": ok(`{"policy":[
		{"resourceType":"CPU","numCpus":1,"price":10},
		{"resourceType":"RAM","megsRam":512,"price":2},
		{"resourceType":"RAM","megsRam":1024,"price":4},
		{"resourceType":"STORAGE","serviceNameInUptime":"backup","price":9},
		{"resourceType":"STORAGE","serviceNameInUptime":"main","price":0.1}
	]}`)})
	pricing, err := c.GetPricing(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// 2 CPUs at 10 + 2048MB at 4/1024MB + 20GB at 0.1.
	got, found := pricing.VMHourly(2, 2048, 20)
	if !found || math.Abs(got-30) > 1e-9 {
		t.Fatalf("VMHourly = %v, %v; want 30", got, found)
	}
	if _, found := (Pricing{}).VMHourly(1, 1024, 20); found {
		t.Fatal("empty pricing should not price a VM")
	}
}
