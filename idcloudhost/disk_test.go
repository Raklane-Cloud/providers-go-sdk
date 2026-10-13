package idcloudhost

import (
	"context"
	"net/http"
	"testing"
)

func TestDisks(t *testing.T) {
	c, seen := fakeAPI(t, map[string]func() (int, string){
		"POST /jkt01/storage/disks":                   ok(`{"uuid":"3d91aa31-16ec-44ee-b8b3-22a0bda6559e","size_gb":50,"status":"Active"}`),
		"POST /jkt01/user-resource/vm/storage/attach": ok(`{"uuid":"3d91aa31-16ec-44ee-b8b3-22a0bda6559e","name":"vdb","size":50}`),
		"PATCH /jkt01/user-resource/vm/storage":       ok(`{"uuid":"3d91aa31-16ec-44ee-b8b3-22a0bda6559e","size":60}`),
		"POST /jkt01/user-resource/vm/storage/detach": ok(`{"success":true}`),
		"DELETE /jkt01/storage/disks/3d91aa31-16ec-44ee-b8b3-22a0bda6559e": func() (int, string) {
			return http.StatusNoContent, ""
		},
	})
	ctx := context.Background()
	disk, err := c.CreateDisk(ctx, CreateDiskRequest{SizeGB: 50, DisplayName: "db data"})
	if err != nil {
		t.Fatal(err)
	}
	if f := (*seen)[0].Form; f.Get("size_gb") != "50" || f.Get("source_image_type") != DiskSourceEmpty || f.Get("billing_account_id") != "6" {
		t.Errorf("create form = %v", f)
	}
	if _, err := c.AttachDisk(ctx, "vm-1", disk.UUID); err != nil {
		t.Fatal(err)
	}
	if f := (*seen)[1].Form; f.Get("uuid") != "vm-1" || f.Get("storage_uuid") != disk.UUID {
		t.Errorf("attach form = %v", f)
	}
	resized, err := c.ResizeVMDisk(ctx, "vm-1", disk.UUID, 60)
	if err != nil || resized.SizeGB != 60 {
		t.Fatalf("resize = %+v, %v", resized, err)
	}
	if err := c.DetachDisk(ctx, "vm-1", disk.UUID); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteDisk(ctx, disk.UUID); err != nil {
		t.Fatal(err)
	}
	if got := DiskDeviceID(disk.UUID); got != "virtio-3d91aa31-16ec-44ee-b" {
		t.Errorf("DiskDeviceID = %q", got)
	}
}
