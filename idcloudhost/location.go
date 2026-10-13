package idcloudhost

import (
	"context"
	"fmt"
	"net/http"
)

// Location is a data centre resources can be created in.
type Location struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	CountryCode string `json:"country_code"`
	// IsDefault marks the location calls use when Client.Location is "".
	IsDefault   bool `json:"is_default"`
	IsPreferred bool `json:"is_preferred"`
	OrderNr     int  `json:"order_nr"`
}

// ListLocations returns every location.
func (c *Client) ListLocations(ctx context.Context) ([]Location, error) {
	var locations []Location
	if err := c.do(ctx, http.MethodGet, "/config/locations", nil, nil, &locations); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to list locations: %w", err)
	}
	return locations, nil
}

// HostPool is a server class VMs can be placed on
// (CreateVMRequest.DesignatedPoolUUID).
type HostPool struct {
	UUID                string `json:"uuid"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	IsDefaultDesignated bool   `json:"is_default_designated"`
}

// ListHostPools returns the server classes in the client's location.
func (c *Client) ListHostPools(ctx context.Context) ([]HostPool, error) {
	var pools []HostPool
	if err := c.do(ctx, http.MethodGet, c.located("/user-resource/host_pool/list"), nil, nil, &pools); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to list host pools: %w", err)
	}
	return pools, nil
}
