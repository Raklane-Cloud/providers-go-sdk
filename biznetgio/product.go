package biznetgio

import (
	"context"
	"fmt"
	"net/http"
)

// Product is a NEO Lite size customers can order.
type Product struct {
	ProductID    ID             `json:"product_id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	CategoryID   ID             `json:"category_id"`
	CategoryName string         `json:"category_name"`
	Options      ProductOptions `json:"options"`
	// Billing is its price on each billing cycle it's sold on.
	Billing []ProductBilling `json:"billing"`
}

// ProductOptions is a product's size.
type ProductOptions struct {
	Type  string `json:"type"`
	Cores Int    `json:"cores"`
	// Memory is in MB.
	Memory Int `json:"memory"`
}

// ProductBilling is a product's price on one billing cycle, in rupiah.
type ProductBilling struct {
	Label string `json:"label"`
	Cycle string `json:"cycle"`
	Price Int    `json:"price"`
	// Components are priced add-ons on this cycle, e.g. extra disk.
	Components []BillingComponent `json:"components"`
}

// BillingComponent is a priced add-on, e.g. "disk_size" per GB.
type BillingComponent struct {
	Label  string           `json:"label"`
	Field  string           `json:"field"`
	Prices []ComponentPrice `json:"prices"`
}

// ComponentPrice is a component's unit price for quantities from QtyMin
// to QtyMax.
type ComponentPrice struct {
	QtyMin Int `json:"qty_min"`
	QtyMax Int `json:"qty_max"`
	Price  Int `json:"price"`
}

// Price is what p costs on cycle, and whether it's sold on it.
func (p Product) Price(cycle string) (int64, bool) {
	for _, b := range p.Billing {
		if b.Cycle == cycle {
			return int64(b.Price), true
		}
	}
	return 0, false
}

// OS is an operating system a product can be ordered with.
type OS struct {
	Name string `json:"name"`
	VMID ID     `json:"vmid"`
	Node string `json:"node"`
	// MaxMem is in bytes.
	MaxMem Int `json:"maxmem"`
	MaxCPU Int `json:"maxcpu"`
}

// ListNeoLiteProducts returns the NEO Lite sizes on sale.
func (c *Client) ListNeoLiteProducts(ctx context.Context) ([]Product, error) {
	var products []Product
	if err := c.do(ctx, http.MethodGet, "/neolites/products", nil, nil, &products); err != nil {
		return nil, fmt.Errorf("biznetgio: failed to list neo lite products: %w", err)
	}
	return products, nil
}

// GetNeoLiteProduct returns one NEO Lite size.
func (c *Client) GetNeoLiteProduct(ctx context.Context, productID ID) (*Product, error) {
	var p Product
	if err := c.do(ctx, http.MethodGet, "/neolites/products/"+productID.String(), nil, nil, &p); err != nil {
		return nil, fmt.Errorf("biznetgio: failed to get neo lite product %s: %w", productID, err)
	}
	return &p, nil
}

// ListNeoLiteOS returns the operating systems productID can be ordered
// with - CreateNeoLiteRequest.OS takes a Name.
func (c *Client) ListNeoLiteOS(ctx context.Context, productID ID) ([]OS, error) {
	var oss []OS
	if err := c.do(ctx, http.MethodGet, "/neolites/products/"+productID.String()+"/oss", nil, nil, &oss); err != nil {
		return nil, fmt.Errorf("biznetgio: failed to list operating systems for product %s: %w", productID, err)
	}
	return oss, nil
}
