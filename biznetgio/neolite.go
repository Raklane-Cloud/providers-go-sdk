package biznetgio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Raklane-Cloud/providers-go-sdk/internal/httpx"
)

// Billing cycles a NEO Lite server can be ordered on. There is no hourly
// cycle.
const (
	CycleMonthly    = "m"
	CycleQuarterly  = "q"
	CycleSemiAnnual = "s"
	CycleAnnual     = "a"
	CycleBiennial   = "b"
	CycleTriennial  = "t"
)

// Account statuses: an order is Pending until its invoice is paid and the
// server is provisioned, then Active.
const (
	StatusActive     = "Active"
	StatusPending    = "Pending"
	StatusSuspended  = "Suspended"
	StatusTerminated = "Terminated"
)

// Server power actions for SetNeoLiteState.
const (
	StateStart    = "start"
	StateStop     = "stop"
	StateShutdown = "shutdown"
	StateReset    = "reset"
	StateSuspend  = "suspend"
	StateResume   = "resume"
)

// Account is a NEO Lite server as the billing side knows it: one ordered
// product, its status and its server's settings.
type Account struct {
	AccountID       ID             `json:"account_id"`
	Domain          string         `json:"domain"`
	Status          string         `json:"status"`
	BillingCycle    string         `json:"billingcycle"`
	DateCreated     string         `json:"date_created"`
	NextDue         string         `json:"next_due"`
	RecurringAmount int64          `json:"recurring_amount"`
	ProductID       ID             `json:"product_id"`
	ProductName     string         `json:"product_name"`
	Description     string         `json:"description"`
	CategoryID      ID             `json:"category_id"`
	CategoryName    string         `json:"category_name"`
	Details         AccountDetails `json:"extra_details"`
	LastInvoice     Invoice        `json:"last_invoice"`

	// Raw is the account exactly as the API sent it, for fields this
	// package doesn't model.
	Raw json.RawMessage `json:"-"`
}

func (a *Account) UnmarshalJSON(b []byte) error {
	type plain Account
	var p plain
	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}
	*a = Account(p)
	a.Raw = append(json.RawMessage(nil), b...)
	return nil
}

// AccountDetails are a NEO Lite server's settings.
type AccountDetails struct {
	Region      string `json:"region"`
	RegionLabel string `json:"region_label"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// User is the login user the server was created with.
	User      string `json:"ciuser"`
	KeypairID ID     `json:"neosshkey_id"`
	SSHKeys   string `json:"sshkeys"`
	OSName    string `json:"osname"`
	DiskSize  string `json:"disk_size"`
}

// Invoice is an account's latest invoice.
type Invoice struct {
	ID        ID     `json:"id"`
	Status    string `json:"status"`
	Date      string `json:"date"`
	DueDate   string `json:"duedate"`
	PayBefore string `json:"paybefore"`
	DatePaid  string `json:"datepaid"`
	Type      string `json:"invoice_type"`
}

// Order is what placing an order returns.
type Order struct {
	OrderID   string `json:"order_id"`
	AccountID ID     `json:"account_id"`
}

// VM is a NEO Lite server's machine.
type VM struct {
	VMID   ID     `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"`
	// Uptime is in seconds; MaxMem and Mem in bytes or MB, as the
	// hypervisor reports them.
	Uptime  int64 `json:"uptime"`
	MaxDisk int64 `json:"maxdisk"`
	MaxMem  int64 `json:"maxmem"`
	Mem     int64 `json:"mem"`
	CPUs    int64 `json:"cpus"`
}

// CreateNeoLiteRequest orders a NEO Lite server. ProductID, Cycle, OS,
// KeypairID, User and ConsolePassword are required.
type CreateNeoLiteRequest struct {
	ProductID ID
	// Cycle is one of the Cycle constants.
	Cycle string
	// OS is an operating system's name from ListNeoLiteOS.
	OS        string
	KeypairID ID
	// User is the login user; ConsolePassword its password on the
	// provider's web console.
	User            string
	ConsolePassword string
	VMName          string
	Description     string
	Promocode       string
	// PayWithCreditCard pays the order's invoice with the card saved on
	// the account, so the server is provisioned right away.
	PayWithCreditCard bool
}

// CreateNeoLite places an order for a NEO Lite server. The server exists
// once the account is Active (see GetNeoLite).
func (c *Client) CreateNeoLite(ctx context.Context, req CreateNeoLiteRequest) (*Order, error) {
	body := map[string]any{
		"product_id":           int64(req.ProductID),
		"cycle":                req.Cycle,
		"select_os":            req.OS,
		"keypair_id":           int64(req.KeypairID),
		"ssh_and_console_user": req.User,
		"console_password":     req.ConsolePassword,
	}
	if req.VMName != "" {
		body["vm_name"] = req.VMName
	}
	if req.Description != "" {
		body["description"] = req.Description
	}
	if req.Promocode != "" {
		body["promocode"] = req.Promocode
	}
	if req.PayWithCreditCard {
		body["pay_invoice_with_cc"] = true
	}
	var order Order
	if err := c.do(ctx, http.MethodPost, "/neolites", nil, httpx.JSON(body), &order); err != nil {
		return nil, fmt.Errorf("biznetgio: failed to order neo lite %s: %w", req.VMName, err)
	}
	return &order, nil
}

// GetNeoLite returns one NEO Lite account.
func (c *Client) GetNeoLite(ctx context.Context, accountID ID) (*Account, error) {
	var a Account
	if err := c.do(ctx, http.MethodGet, "/neolites/accounts/"+accountID.String(), nil, nil, &a); err != nil {
		return nil, fmt.Errorf("biznetgio: failed to get neo lite %s: %w", accountID, err)
	}
	return &a, nil
}

// ListNeoLites returns the NEO Lite accounts, only those with status when
// it isn't "".
func (c *Client) ListNeoLites(ctx context.Context, status string) ([]Account, error) {
	var query url.Values
	if status != "" {
		query = url.Values{"status": {status}}
	}
	var accounts []Account
	if err := c.do(ctx, http.MethodGet, "/neolites/accounts", query, nil, &accounts); err != nil {
		return nil, fmt.Errorf("biznetgio: failed to list neo lites: %w", err)
	}
	return accounts, nil
}

// GetNeoLiteVM returns the machine of an active NEO Lite account.
func (c *Client) GetNeoLiteVM(ctx context.Context, accountID ID) (*VM, error) {
	var vm VM
	if err := c.do(ctx, http.MethodGet, "/neolites/accounts/"+accountID.String()+"/vm-details", nil, nil, &vm); err != nil {
		return nil, fmt.Errorf("biznetgio: failed to get neo lite %s's vm: %w", accountID, err)
	}
	return &vm, nil
}

// DeleteNeoLite cancels a NEO Lite account and deletes its server.
func (c *Client) DeleteNeoLite(ctx context.Context, accountID ID) error {
	if err := c.do(ctx, http.MethodDelete, "/neolites/"+accountID.String(), nil, nil, nil); err != nil {
		return fmt.Errorf("biznetgio: failed to delete neo lite %s: %w", accountID, err)
	}
	return nil
}

// SetNeoLiteState starts, stops or resets a NEO Lite server - one of the
// State constants.
func (c *Client) SetNeoLiteState(ctx context.Context, accountID ID, state string) error {
	path := "/neolites/accounts/" + accountID.String() + "/vm-state/" + url.PathEscape(state)
	if err := c.do(ctx, http.MethodPut, path, nil, nil, nil); err != nil {
		return fmt.Errorf("biznetgio: failed to %s neo lite %s: %w", state, accountID, err)
	}
	return nil
}
