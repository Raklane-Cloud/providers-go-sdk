package idcloudhost

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Raklane-Cloud/providers-go-sdk/internal/httpx"
)

// Firewall rule values.
const (
	DirectionInbound  = "inbound"
	DirectionOutbound = "outbound"

	EndpointAny        = "any"
	EndpointIPPrefixes = "ip_prefixes"
)

// Firewall is a set of rules assigned to VMs.
type Firewall struct {
	UUID string `json:"uuid"`
	// Name is what the list call reports; DisplayName what create
	// reports. Use FirewallName to read either.
	Name              string               `json:"name"`
	DisplayName       string               `json:"display_name"`
	Description       string               `json:"description"`
	BillingAccountID  int                  `json:"billing_account_id"`
	Rules             []FirewallRule       `json:"rules"`
	ResourcesAssigned []FirewallAssignment `json:"resources_assigned"`
	CreatedAt         string               `json:"created_at"`
}

// FirewallName is the firewall's name, whichever field carries it.
func (f Firewall) FirewallName() string {
	if f.Name != "" {
		return f.Name
	}
	return f.DisplayName
}

// FirewallRule allows one kind of traffic.
type FirewallRule struct {
	UUID string `json:"uuid,omitempty"`
	// Protocol is "tcp", "udp" or "icmp".
	Protocol  string `json:"protocol"`
	Direction string `json:"direction"`
	// PortStart nil means every port; PortEnd nil means PortStart only.
	PortStart *int `json:"port_start"`
	PortEnd   *int `json:"port_end"`
	// EndpointSpecType is EndpointAny or EndpointIPPrefixes; the latter
	// lists addresses or CIDR ranges in EndpointSpec.
	EndpointSpecType string   `json:"endpoint_spec_type"`
	EndpointSpec     []string `json:"endpoint_spec,omitempty"`
}

// FirewallAssignment is a resource a firewall applies to.
type FirewallAssignment struct {
	ResourceType string `json:"resource_type"`
	ResourceUUID string `json:"resource_uuid"`
}

// Port is a helper for FirewallRule's port fields.
func Port(n int) *int { return &n }

// CreateFirewallRequest describes a new firewall.
type CreateFirewallRequest struct {
	DisplayName string
	Description string
	// BillingAccountID defaults to Client.BillingAccountID.
	BillingAccountID int
	Rules            []FirewallRule
}

// ListFirewalls returns the firewalls in the client's location.
func (c *Client) ListFirewalls(ctx context.Context) ([]Firewall, error) {
	var firewalls []Firewall
	if err := c.do(ctx, http.MethodGet, c.located("/network/firewalls"), nil, nil, &firewalls); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to list firewalls: %w", err)
	}
	return firewalls, nil
}

// CreateFirewall creates a firewall.
func (c *Client) CreateFirewall(ctx context.Context, req CreateFirewallRequest) (*Firewall, error) {
	body := map[string]any{
		"display_name":       req.DisplayName,
		"billing_account_id": c.billingAccount(req.BillingAccountID),
		"rules":              nonNilRules(req.Rules),
	}
	if req.Description != "" {
		body["description"] = req.Description
	}
	var firewall Firewall
	if err := c.do(ctx, http.MethodPost, c.located("/network/firewalls"), nil, httpx.JSON(body), &firewall); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to create firewall %s: %w", req.DisplayName, err)
	}
	return &firewall, nil
}

// UpdateFirewallRequest replaces a firewall's name, description and rules.
type UpdateFirewallRequest struct {
	Name        string
	Description string
	Rules       []FirewallRule
}

// UpdateFirewall replaces a firewall's configuration.
func (c *Client) UpdateFirewall(ctx context.Context, uuid string, req UpdateFirewallRequest) (*Firewall, error) {
	body := map[string]any{"rules": nonNilRules(req.Rules)}
	if req.Name != "" {
		body["name"] = req.Name
	}
	if req.Description != "" {
		body["description"] = req.Description
	}
	var firewall Firewall
	if err := c.do(ctx, http.MethodPut, c.located("/network/firewalls/"+url.PathEscape(uuid)), nil, httpx.JSON(body), &firewall); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to update firewall %s: %w", uuid, err)
	}
	return &firewall, nil
}

// DeleteFirewall deletes a firewall.
func (c *Client) DeleteFirewall(ctx context.Context, uuid string) error {
	if err := c.do(ctx, http.MethodDelete, c.located("/network/firewalls/"+url.PathEscape(uuid)), nil, nil, nil); err != nil {
		return fmt.Errorf("idcloudhost: failed to delete firewall %s: %w", uuid, err)
	}
	return nil
}

// AssignFirewall applies a firewall to a VM.
func (c *Client) AssignFirewall(ctx context.Context, firewallUUID, vmUUID string) error {
	path := c.located("/network/firewalls/" + url.PathEscape(firewallUUID) + "/vms")
	if err := c.do(ctx, http.MethodPost, path, url.Values{"vm_uuid": {vmUUID}}, nil, nil); err != nil {
		return fmt.Errorf("idcloudhost: failed to assign firewall %s to vm %s: %w", firewallUUID, vmUUID, err)
	}
	return nil
}

// UnassignFirewall removes a firewall from a VM.
func (c *Client) UnassignFirewall(ctx context.Context, firewallUUID, vmUUID string) error {
	path := c.located("/network/firewalls/" + url.PathEscape(firewallUUID) + "/vms")
	if err := c.do(ctx, http.MethodDelete, path, url.Values{"vm_uuid": {vmUUID}}, nil, nil); err != nil {
		return fmt.Errorf("idcloudhost: failed to unassign firewall %s from vm %s: %w", firewallUUID, vmUUID, err)
	}
	return nil
}

// nonNilRules makes an empty rule list encode as [] rather than null.
func nonNilRules(rules []FirewallRule) []FirewallRule {
	if rules == nil {
		return []FirewallRule{}
	}
	return rules
}
