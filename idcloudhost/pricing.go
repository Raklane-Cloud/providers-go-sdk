package idcloudhost

import (
	"context"
	"fmt"
	"net/http"
)

// Price resource types.
const (
	ResourceCPU     = "CPU"
	ResourceRAM     = "RAM"
	ResourceStorage = "STORAGE"
)

// Price is one entry of the pricing policy: the hourly price of a unit
// of a resource, in the account's currency.
type Price struct {
	PolicyID     int     `json:"policyId"`
	ResourceType string  `json:"resourceType"`
	Price        float64 `json:"price"`
	// NumCPUs is set on CPU prices, MegsRAM on RAM prices.
	NumCPUs int `json:"numCpus"`
	MegsRAM int `json:"megsRam"`
	// ServiceNameInUptime tells STORAGE prices apart: "main" is disk
	// space, "backup" and "snapshot" are copies of it.
	ServiceNameInUptime     string `json:"serviceNameInUptime"`
	ServiceNameUserFriendly string `json:"serviceNameUserFriendly"`
}

// Pricing is the platform's pricing policy.
type Pricing []Price

// GetPricing returns the pricing policy.
func (c *Client) GetPricing(ctx context.Context) (Pricing, error) {
	var resp struct {
		Policy Pricing `json:"policy"`
	}
	if err := c.do(ctx, http.MethodGet, "/pricing/policy", nil, nil, &resp); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to get pricing: %w", err)
	}
	return resp.Policy, nil
}

// VMHourly estimates what a running VM with vcpu CPUs, ramMB of memory
// and a diskGB boot disk costs per hour. ok is false when the policy has
// no CPU, RAM or disk price to base it on.
func (p Pricing) VMHourly(vcpu, ramMB, diskGB int) (price float64, ok bool) {
	cpu, cpuOK := p.perUnit(ResourceCPU, vcpu, func(e Price) int { return e.NumCPUs })
	ram, ramOK := p.perUnit(ResourceRAM, ramMB, func(e Price) int { return e.MegsRAM })
	var disk float64
	diskOK := false
	for _, e := range p {
		if e.ResourceType == ResourceStorage && e.ServiceNameInUptime == "main" {
			disk, diskOK = e.Price*float64(diskGB), true
			break
		}
	}
	if !cpuOK || !ramOK || !diskOK {
		return 0, false
	}
	return cpu + ram + disk, true
}

// perUnit prices amount of resourceType: an entry for exactly amount if
// the policy has one, otherwise the largest entry's per-unit rate.
func (p Pricing) perUnit(resourceType string, amount int, units func(Price) int) (float64, bool) {
	var best Price
	for _, e := range p {
		if e.ResourceType != resourceType || units(e) <= 0 {
			continue
		}
		if units(e) == amount {
			return e.Price, true
		}
		if units(e) > units(best) {
			best = e
		}
	}
	if units(best) <= 0 {
		return 0, false
	}
	return best.Price / float64(units(best)) * float64(amount), true
}
