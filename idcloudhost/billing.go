package idcloudhost

import (
	"context"
	"fmt"
	"net/http"
)

// BillingAccount is what created resources are charged to.
type BillingAccount struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	CompanyName string `json:"company_name"`
	IsDefault   bool   `json:"is_default"`
	IsActive    bool   `json:"is_active"`
	IsDeleted   bool   `json:"is_deleted"`
	// CreditAmount is the prepaid balance resources are paid from.
	CreditAmount float64 `json:"credit_amount"`
	AllowDebt    bool    `json:"allow_debt"`
	CanPay       bool    `json:"can_pay"`
}

// ListBillingAccounts returns the account's billing accounts.
func (c *Client) ListBillingAccounts(ctx context.Context) ([]BillingAccount, error) {
	var accounts []BillingAccount
	if err := c.do(ctx, http.MethodGet, "/payment/billing_account/list", nil, nil, &accounts); err != nil {
		return nil, fmt.Errorf("idcloudhost: failed to list billing accounts: %w", err)
	}
	return accounts, nil
}
