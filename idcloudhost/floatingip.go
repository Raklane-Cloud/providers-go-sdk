package idcloudhost

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Raklane-Cloud/providers-go-sdk/internal/httpx"
)

// FloatingIP is a public IPv4 address, assigned to a VM or reserved.
type FloatingIP struct {
	ID               int    `json:"id"`
	UUID             string `json:"uuid"`
	Address          string `json:"address"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	BillingAccountID int    `json:"billing_account_id"`
	Enabled          bool   `json:"enabled"`
	IsDeleted        bool   `json:"is_deleted"`
	// AssignedTo is the UUID of the VM (or service) using the address;
	// "" while it's only reserved.
	AssignedTo             string `json:"assigned_to"`
	AssignedToResourceType string `json:"assigned_to_resource_type"`
	AssignedToPrivateIP    string `json:"assigned_to_private_ip"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}

// ListFloatingIPsOptions filters ListFloatingIPs. Zero fields don't
// filter.
type ListFloatingIPsOptions struct {
	BillingAccountID int
	VMUUID           string
}

// ListFloatingIPs returns the floating IPs in the client's location.
func (c *Client) ListFloatingIPs(ctx context.Context, opts ListFloatingIPsOptions) ([]FloatingIP, error) {
	query := url.Values{}
	if opts.BillingAccountID != 0 {
		query.Set("billing_account_id", strconv.Itoa(opts.BillingAccountID))
	}
	if opts.VMUUID != "" {
		query.Set("vm_uuid", opts.VMUUID)
	}
	var ips []FloatingIP
	if err := c.do(ctx, http.MethodGet, c.located("/network/ip_addresses"), query, nil, &ips); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to list floating ips: %w", err)
	}
	return ips, nil
}

// CreateFloatingIP reserves a new public IPv4 address.
func (c *Client) CreateFloatingIP(ctx context.Context, name string, billingAccountID int) (*FloatingIP, error) {
	body := map[string]any{"billing_account_id": c.billingAccount(billingAccountID)}
	if name != "" {
		body["name"] = name
	}
	var ip FloatingIP
	if err := c.do(ctx, http.MethodPost, c.located("/network/ip_addresses"), nil, httpx.JSON(body), &ip); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to create floating ip: %w", err)
	}
	return &ip, nil
}

// GetFloatingIP returns one floating IP by address.
func (c *Client) GetFloatingIP(ctx context.Context, address string) (*FloatingIP, error) {
	var ip FloatingIP
	if err := c.do(ctx, http.MethodGet, c.located("/network/ip_addresses/"+url.PathEscape(address)), nil, nil, &ip); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to get floating ip %s: %w", address, err)
	}
	return &ip, nil
}

// AssignFloatingIP points address at a VM.
func (c *Client) AssignFloatingIP(ctx context.Context, address, vmUUID string) (*FloatingIP, error) {
	var ip FloatingIP
	body := httpx.JSON(map[string]string{"vm_uuid": vmUUID})
	if err := c.do(ctx, http.MethodPost, c.located("/network/ip_addresses/"+url.PathEscape(address)+"/assign"), nil, body, &ip); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to assign floating ip %s to vm %s: %w", address, vmUUID, err)
	}
	return &ip, nil
}

// UnassignFloatingIP takes address off its VM; it stays reserved.
func (c *Client) UnassignFloatingIP(ctx context.Context, address string) (*FloatingIP, error) {
	var ip FloatingIP
	if err := c.do(ctx, http.MethodPost, c.located("/network/ip_addresses/"+url.PathEscape(address)+"/unassign"), nil, nil, &ip); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to unassign floating ip %s: %w", address, err)
	}
	return &ip, nil
}

// DeleteFloatingIP releases address, which stops its charges.
func (c *Client) DeleteFloatingIP(ctx context.Context, address string) error {
	if err := c.do(ctx, http.MethodDelete, c.located("/network/ip_addresses/"+url.PathEscape(address)), nil, nil, nil); err != nil {
		return fmt.Errorf("idcloudhost: failed to delete floating ip %s: %w", address, err)
	}
	return nil
}
