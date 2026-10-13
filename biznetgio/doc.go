// Package biznetgio is a client for the Biznet GIO portal API
// (https://api.portal.biznetgio.com): NEO Lite virtual servers, their
// products and operating systems, and SSH keypairs.
//
// A NEO Lite server is ordered, not just created: CreateNeoLite places an
// order for a product on a billing cycle (monthly or longer - there is no
// hourly billing) and the server is provisioned once its invoice is paid.
// Set PayWithCreditCard to pay it with the card saved on the account, or
// it waits for a manual payment. GetNeoLite reports the order's progress.
//
// The API takes no cloud-init or user-data: configure a new server over
// SSH, with a keypair registered through ImportKeypair.
package biznetgio
