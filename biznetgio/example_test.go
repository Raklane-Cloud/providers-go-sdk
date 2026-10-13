package biznetgio_test

import (
	"context"
	"log"

	"github.com/Raklane-Cloud/providers-go-sdk/biznetgio"
)

func ExampleClient_CreateNeoLite() {
	ctx := context.Background()
	c := biznetgio.New("api-token")

	key, err := c.ImportKeypair(ctx, "deploy", "ssh-ed25519 AAAA... deploy")
	if err != nil {
		log.Fatal(err)
	}
	order, err := c.CreateNeoLite(ctx, biznetgio.CreateNeoLiteRequest{
		ProductID: 1559, Cycle: biznetgio.CycleMonthly, OS: "ubuntu-24.04",
		KeypairID: key.KeypairID, User: "deploy", ConsolePassword: "Generated-Secret-1",
		VMName: "web-1", PayWithCreditCard: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	// The server is ready once the account is Active; then SSH in to set it up.
	account, err := c.GetNeoLite(ctx, order.AccountID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(account.Status)
}
