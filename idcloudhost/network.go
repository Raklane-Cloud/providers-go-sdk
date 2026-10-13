package idcloudhost

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Network is a private network (VPC). VMs in the same network reach each
// other on their PrivateIPv4 addresses.
type Network struct {
	UUID           string   `json:"uuid"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	VLANID         int      `json:"vlan_id"`
	Subnet         string   `json:"subnet"`
	SubnetIPv6     string   `json:"subnet_ipv6"`
	IsDefault      bool     `json:"is_default"`
	VMUUIDs        []string `json:"vm_uuids"`
	ResourcesCount int      `json:"resources_count"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

// ListNetworks returns the private networks in the client's location.
func (c *Client) ListNetworks(ctx context.Context) ([]Network, error) {
	var networks []Network
	if err := c.do(ctx, http.MethodGet, c.located("/network/networks"), nil, nil, &networks); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to list networks: %w", err)
	}
	return networks, nil
}

// GetNetwork returns one private network.
func (c *Client) GetNetwork(ctx context.Context, uuid string) (*Network, error) {
	var network Network
	if err := c.do(ctx, http.MethodGet, c.located("/network/network/"+url.PathEscape(uuid)), nil, nil, &network); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to get network %s: %w", uuid, err)
	}
	return &network, nil
}

// CreateNetwork creates a private network named name. With name empty it
// returns the default network, creating it if there's none.
func (c *Client) CreateNetwork(ctx context.Context, name string) (*Network, error) {
	var query url.Values
	if name != "" {
		query = url.Values{"name": {name}}
	}
	var network Network
	if err := c.do(ctx, http.MethodPost, c.located("/network/network"), query, nil, &network); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to create network %s: %w", name, err)
	}
	return &network, nil
}

// DeleteNetwork deletes a private network. Only an empty, non-default
// network can be deleted.
func (c *Client) DeleteNetwork(ctx context.Context, uuid string) error {
	if err := c.do(ctx, http.MethodDelete, c.located("/network/network/"+url.PathEscape(uuid)), nil, nil, nil); err != nil {
		return fmt.Errorf("idcloudhost: failed to delete network %s: %w", uuid, err)
	}
	return nil
}
