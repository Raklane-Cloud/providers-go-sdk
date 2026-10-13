package idcloudhost

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Raklane-Cloud/providers-go-sdk/internal/httpx"
)

// VM statuses the API reports.
const (
	VMStatusRunning = "running"
	VMStatusStopped = "stopped"
)

// VM is a virtual machine. Its public IPv4 address isn't part of it: it's
// a floating IP assigned to the VM (see ListFloatingIPs).
type VM struct {
	ID          int    `json:"id"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Hostname    string `json:"hostname"`
	Description string `json:"description"`
	// Status is VMStatusRunning, VMStatusStopped, or a transitional state.
	Status    string `json:"status"`
	VCPU      int    `json:"vcpu"`
	MemoryMB  int    `json:"memory"`
	OSName    string `json:"os_name"`
	OSVersion string `json:"os_version"`
	// PrivateIPv4 is the VM's address in its private network.
	PrivateIPv4        string   `json:"private_ipv4"`
	PublicIPv6         string   `json:"public_ipv6"`
	MAC                string   `json:"mac"`
	Username           string   `json:"username"`
	Backup             bool     `json:"backup"`
	BillingAccountID   int      `json:"billing_account"`
	DesignatedPoolUUID string   `json:"designated_pool_uuid"`
	DesignatedPoolName string   `json:"designated_pool_name"`
	Storage            []VMDisk `json:"storage"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

// VMDisk is a disk attached to a VM.
type VMDisk struct {
	ID      int    `json:"id"`
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	Primary bool   `json:"primary"`
	// SizeGB is the disk's capacity in GiB.
	SizeGB    int    `json:"size"`
	Type      string `json:"type"`
	Pool      string `json:"pool"`
	Shared    bool   `json:"shared"`
	CreatedAt string `json:"created_at"`
}

// CreateVMRequest describes a new VM. Name, OSName, OSVersion, DiskGB,
// VCPU, RAMMB, Username and Password are required by the API (see
// VMParameters for the allowed values).
type CreateVMRequest struct {
	// Name must match ^[0-9a-zA-Z][-0-9a-zA-Z]{2,}[0-9a-zA-Z]$.
	Name        string
	Description string
	OSName      string
	OSVersion   string
	// DiskGB is the boot disk's size.
	DiskGB int
	VCPU   int
	RAMMB  int
	// Username and Password are the login user the platform creates;
	// Password needs a lower and upper case letter and a digit, 8+ long.
	Username string
	Password string
	// PublicKeys are OpenSSH public key lines installed for Username.
	PublicKeys []string
	// BillingAccountID defaults to Client.BillingAccountID.
	BillingAccountID int
	// NetworkUUID is the private network to create the VM in; empty uses
	// the default network.
	NetworkUUID string
	// DesignatedPoolUUID picks the server class (see ListHostPools).
	DesignatedPoolUUID string
	// CloudInit is cloud-init user-data as YAML (e.g. a "#cloud-config"
	// document) or JSON. Its keys override the platform's own.
	CloudInit string
	// NoPublicIP creates the VM without a public IPv4 address.
	NoPublicIP bool
	Backup     bool
	// DiskUUID boots the VM from an existing unattached disk; DiskGB is
	// then ignored.
	DiskUUID string
}

func (r CreateVMRequest) form(billingAccountID int) url.Values {
	v := url.Values{}
	v.Set("name", r.Name)
	set := func(key, value string) {
		if value != "" {
			v.Set(key, value)
		}
	}
	set("description", r.Description)
	set("os_name", r.OSName)
	set("os_version", r.OSVersion)
	if r.DiskUUID != "" {
		v.Set("disk_uuid", r.DiskUUID)
	} else if r.DiskGB > 0 {
		v.Set("disks", strconv.Itoa(r.DiskGB))
	}
	v.Set("vcpu", strconv.Itoa(r.VCPU))
	v.Set("ram", strconv.Itoa(r.RAMMB))
	set("username", r.Username)
	set("password", r.Password)
	for _, key := range r.PublicKeys {
		v.Add("public_keys", key)
	}
	if billingAccountID != 0 {
		v.Set("billing_account_id", strconv.Itoa(billingAccountID))
	}
	set("network_uuid", r.NetworkUUID)
	set("designated_pool_uuid", r.DesignatedPoolUUID)
	set("cloud_init", r.CloudInit)
	v.Set("reserve_public_ip", strconv.FormatBool(!r.NoPublicIP))
	v.Set("backup", strconv.FormatBool(r.Backup))
	return v
}

// CreateVM creates and starts a VM. The call can fail after the VM was
// created (a timeout, a dropped connection): look it up by name with
// ListVMs before retrying, or a second VM is created.
func (c *Client) CreateVM(ctx context.Context, req CreateVMRequest) (*VM, error) {
	var vm VM
	form := req.form(c.billingAccount(req.BillingAccountID))
	if err := c.do(ctx, http.MethodPost, c.located("/user-resource/vm"), nil, httpx.Form(form), &vm); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to create vm %s: %w", req.Name, err)
	}
	return &vm, nil
}

// GetVM returns one VM. A VM that doesn't exist is an error IsNotFound
// recognises.
func (c *Client) GetVM(ctx context.Context, uuid string) (*VM, error) {
	var vm VM
	if err := c.do(ctx, http.MethodGet, c.located("/user-resource/vm"), url.Values{"uuid": {uuid}}, nil, &vm); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to get vm %s: %w", uuid, err)
	}
	return &vm, nil
}

// ListVMs returns every VM in the client's location.
func (c *Client) ListVMs(ctx context.Context) ([]VM, error) {
	var vms []VM
	if err := c.do(ctx, http.MethodGet, c.located("/user-resource/vm/list"), nil, nil, &vms); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to list vms: %w", err)
	}
	return vms, nil
}

// DeleteVM deletes a VM and its boot disk.
func (c *Client) DeleteVM(ctx context.Context, uuid string) error {
	if err := c.do(ctx, http.MethodDelete, c.located("/user-resource/vm"), nil, httpx.Form(url.Values{"uuid": {uuid}}), nil); err != nil {
		return fmt.Errorf("idcloudhost: failed to delete vm %s: %w", uuid, err)
	}
	return nil
}

// StartVM powers a stopped VM on.
func (c *Client) StartVM(ctx context.Context, uuid string) error {
	if err := c.do(ctx, http.MethodPost, c.located("/user-resource/vm/start"), nil, httpx.Form(url.Values{"uuid": {uuid}}), nil); err != nil {
		return fmt.Errorf("idcloudhost: failed to start vm %s: %w", uuid, err)
	}
	return nil
}

// StopVM shuts a VM down through ACPI, or cuts its power when force is
// set.
func (c *Client) StopVM(ctx context.Context, uuid string, force bool) error {
	form := url.Values{"uuid": {uuid}, "force": {strconv.FormatBool(force)}}
	if err := c.do(ctx, http.MethodPost, c.located("/user-resource/vm/stop"), nil, httpx.Form(form), nil); err != nil {
		return fmt.Errorf("idcloudhost: failed to stop vm %s: %w", uuid, err)
	}
	return nil
}

// ModifyVMRequest changes a VM. Zero fields are left as they are. VCPU
// and RAMMB can only change while the VM is stopped.
type ModifyVMRequest struct {
	Name  string
	VCPU  int
	RAMMB int
}

// ModifyVM renames or resizes a VM.
func (c *Client) ModifyVM(ctx context.Context, uuid string, req ModifyVMRequest) (*VM, error) {
	form := url.Values{"uuid": {uuid}}
	if req.Name != "" {
		form.Set("name", req.Name)
	}
	if req.VCPU > 0 {
		form.Set("vcpu", strconv.Itoa(req.VCPU))
	}
	if req.RAMMB > 0 {
		form.Set("ram", strconv.Itoa(req.RAMMB))
	}
	var vm VM
	if err := c.do(ctx, http.MethodPatch, c.located("/user-resource/vm"), nil, httpx.Form(form), &vm); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to modify vm %s: %w", uuid, err)
	}
	return &vm, nil
}
