package idcloudhost

import (
	"context"
	"fmt"
	"net/http"
)

// Image is an operating system VMs can be created with.
type Image struct {
	OSName       string         `json:"os_name"`
	DisplayName  string         `json:"display_name"`
	IsDefault    bool           `json:"is_default"`
	IsAppCatalog bool           `json:"is_app_catalog"`
	Versions     []ImageVersion `json:"versions"`
}

// ImageVersion is one release of an Image.
type ImageVersion struct {
	OSVersion   string `json:"os_version"`
	DisplayName string `json:"display_name"`
	Published   bool   `json:"published"`
}

// ListImages returns the plain OS images (no app catalog images).
func (c *Client) ListImages(ctx context.Context) ([]Image, error) {
	var images []Image
	if err := c.do(ctx, http.MethodGet, "/config/vm_images/plain_os", nil, nil, &images); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to list images: %w", err)
	}
	return images, nil
}
