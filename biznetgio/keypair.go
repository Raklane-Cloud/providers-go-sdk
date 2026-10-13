package biznetgio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Raklane-Cloud/providers-go-sdk/internal/httpx"
)

// Keypair is an SSH public key new servers can be created with.
type Keypair struct {
	KeypairID ID     `json:"keypair_id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

// UnmarshalJSON reads the id from either keypair_id or neosshkey_id - the
// field's name varies between endpoints.
func (k *Keypair) UnmarshalJSON(b []byte) error {
	var raw struct {
		KeypairID   ID     `json:"keypair_id"`
		NeoSSHKeyID ID     `json:"neosshkey_id"`
		Name        string `json:"name"`
		PublicKey   string `json:"public_key"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	k.KeypairID = raw.KeypairID
	if k.KeypairID == 0 {
		k.KeypairID = raw.NeoSSHKeyID
	}
	k.Name, k.PublicKey = raw.Name, raw.PublicKey
	return nil
}

// ListKeypairs returns the account's keypairs.
func (c *Client) ListKeypairs(ctx context.Context) ([]Keypair, error) {
	var keypairs []Keypair
	if err := c.do(ctx, http.MethodGet, "/neolites/keypairs/", nil, nil, &keypairs); err != nil {
		return nil, fmt.Errorf("biznetgio: failed to list keypairs: %w", err)
	}
	return keypairs, nil
}

// ImportKeypair registers an OpenSSH public key under name.
func (c *Client) ImportKeypair(ctx context.Context, name, publicKey string) (*Keypair, error) {
	var k Keypair
	body := httpx.JSON(map[string]string{"name": name, "public_key": publicKey})
	if err := c.do(ctx, http.MethodPost, "/neolites/keypairs/import", nil, body, &k); err != nil {
		return nil, fmt.Errorf("biznetgio: failed to import keypair %s: %w", name, err)
	}
	return &k, nil
}

// DeleteKeypair deletes a keypair.
func (c *Client) DeleteKeypair(ctx context.Context, keypairID ID) error {
	if err := c.do(ctx, http.MethodDelete, "/neolites/keypairs/"+keypairID.String(), nil, nil, nil); err != nil {
		return fmt.Errorf("biznetgio: failed to delete keypair %s: %w", keypairID, err)
	}
	return nil
}
