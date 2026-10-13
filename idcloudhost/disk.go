package idcloudhost

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Raklane-Cloud/providers-go-sdk/internal/httpx"
)

// Disk source image types for CreateDiskRequest.SourceImageType.
const (
	DiskSourceEmpty    = "EMPTY"
	DiskSourceOSBase   = "OS_BASE"
	DiskSourceDisk     = "DISK"
	DiskSourceSnapshot = "SNAPSHOT"
)

// Disk is a block storage disk, attached to a VM or not.
type Disk struct {
	UUID             string         `json:"uuid"`
	Status           string         `json:"status"`
	DisplayName      string         `json:"display_name"`
	SizeGB           int            `json:"size_gb"`
	BillingAccountID int            `json:"billing_account_id"`
	SourceImageType  string         `json:"source_image_type"`
	SourceImage      string         `json:"source_image"`
	StoragePoolUUID  string         `json:"storage_pool_uuid"`
	Snapshots        []DiskSnapshot `json:"snapshots"`
	CreatedAt        string         `json:"created_at"`
	UpdatedAt        string         `json:"updated_at"`
}

// DiskSnapshot is a point-in-time copy of a disk.
type DiskSnapshot struct {
	UUID        string `json:"uuid"`
	DiskUUID    string `json:"disk_uuid"`
	DisplayName string `json:"display_name"`
	SizeGB      int    `json:"size_gb"`
	CreatedAt   string `json:"created_at"`
}

// CreateDiskRequest describes a new disk.
type CreateDiskRequest struct {
	// SizeGB is required unless the source image implies it.
	SizeGB      int
	DisplayName string
	// BillingAccountID defaults to Client.BillingAccountID.
	BillingAccountID int
	// SourceImageType is one of the DiskSource constants; empty is
	// DiskSourceEmpty.
	SourceImageType string
	// SourceImage is required unless SourceImageType is DiskSourceEmpty:
	// an OS image such as "ubuntu_24.04", or a disk or snapshot UUID.
	SourceImage string
}

// CreateDisk creates a disk, not attached to any VM.
func (c *Client) CreateDisk(ctx context.Context, req CreateDiskRequest) (*Disk, error) {
	form := url.Values{}
	if req.SizeGB > 0 {
		form.Set("size_gb", strconv.Itoa(req.SizeGB))
	}
	if req.DisplayName != "" {
		form.Set("display_name", req.DisplayName)
	}
	if id := c.billingAccount(req.BillingAccountID); id != 0 {
		form.Set("billing_account_id", strconv.Itoa(id))
	}
	sourceType := req.SourceImageType
	if sourceType == "" {
		sourceType = DiskSourceEmpty
	}
	form.Set("source_image_type", sourceType)
	if req.SourceImage != "" {
		form.Set("source_image", req.SourceImage)
	}
	var disk Disk
	if err := c.do(ctx, http.MethodPost, c.located("/storage/disks"), nil, httpx.Form(form), &disk); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to create disk %s: %w", req.DisplayName, err)
	}
	return &disk, nil
}

// GetDisk returns one disk.
func (c *Client) GetDisk(ctx context.Context, uuid string) (*Disk, error) {
	var disk Disk
	if err := c.do(ctx, http.MethodGet, c.located("/storage/disks/"+url.PathEscape(uuid)), nil, nil, &disk); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to get disk %s: %w", uuid, err)
	}
	return &disk, nil
}

// ListDisks returns every disk in the client's location.
func (c *Client) ListDisks(ctx context.Context) ([]Disk, error) {
	var disks []Disk
	if err := c.do(ctx, http.MethodGet, c.located("/storage/disks"), nil, nil, &disks); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to list disks: %w", err)
	}
	return disks, nil
}

// DeleteDisk deletes a disk and all its snapshots.
func (c *Client) DeleteDisk(ctx context.Context, uuid string) error {
	if err := c.do(ctx, http.MethodDelete, c.located("/storage/disks/"+url.PathEscape(uuid)), nil, nil, nil); err != nil {
		return fmt.Errorf("idcloudhost: failed to delete disk %s: %w", uuid, err)
	}
	return nil
}

// AttachDisk attaches a disk to a VM. On Linux it shows up as
// /dev/disk/by-id/<DiskDeviceID(diskUUID)>.
func (c *Client) AttachDisk(ctx context.Context, vmUUID, diskUUID string) (*VMDisk, error) {
	form := url.Values{"uuid": {vmUUID}, "storage_uuid": {diskUUID}}
	var disk VMDisk
	if err := c.do(ctx, http.MethodPost, c.located("/user-resource/vm/storage/attach"), nil, httpx.Form(form), &disk); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to attach disk %s to vm %s: %w", diskUUID, vmUUID, err)
	}
	return &disk, nil
}

// DetachDisk detaches a disk from a VM, keeping the disk and its data.
func (c *Client) DetachDisk(ctx context.Context, vmUUID, diskUUID string) error {
	form := url.Values{"uuid": {vmUUID}, "storage_uuid": {diskUUID}}
	if err := c.do(ctx, http.MethodPost, c.located("/user-resource/vm/storage/detach"), nil, httpx.Form(form), nil); err != nil {
		return fmt.Errorf("idcloudhost: failed to detach disk %s from vm %s: %w", diskUUID, vmUUID, err)
	}
	return nil
}

// ResizeVMDisk grows a disk attached to a VM to sizeGB. Disks can't
// shrink.
func (c *Client) ResizeVMDisk(ctx context.Context, vmUUID, diskUUID string, sizeGB int) (*VMDisk, error) {
	form := url.Values{"uuid": {vmUUID}, "disk_uuid": {diskUUID}, "size_gb": {strconv.Itoa(sizeGB)}}
	var disk VMDisk
	if err := c.do(ctx, http.MethodPatch, c.located("/user-resource/vm/storage"), nil, httpx.Form(form), &disk); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to resize disk %s on vm %s: %w", diskUUID, vmUUID, err)
	}
	return &disk, nil
}

// DiskDeviceID is the name an attached disk gets under /dev/disk/by-id
// on a Linux VM: "virtio-" plus the first 20 characters of its UUID.
func DiskDeviceID(diskUUID string) string {
	id := diskUUID
	if len(id) > 20 {
		id = id[:20]
	}
	return "virtio-" + id
}
