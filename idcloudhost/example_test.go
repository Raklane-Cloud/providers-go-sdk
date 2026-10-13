package idcloudhost_test

import (
	"context"
	"log"

	"github.com/Raklane-Cloud/providers-go-sdk/idcloudhost"
)

func ExampleClient_CreateVM() {
	ctx := context.Background()
	c := idcloudhost.New("api-key", "jkt01") // "" = the account's default location
	c.BillingAccountID = 123

	vm, err := c.CreateVM(ctx, idcloudhost.CreateVMRequest{
		Name: "web-1", OSName: "ubuntu", OSVersion: "24.04",
		DiskGB: 40, VCPU: 2, RAMMB: 4096,
		Username: "admin", Password: "Generated-Secret-1",
		CloudInit: "#cloud-config\nruncmd:\n  - echo hello\n",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("created", vm.UUID)
}

func ExampleIsNotFound() {
	ctx := context.Background()
	c := idcloudhost.New("api-key", "jkt01")

	// Treat "already gone" as success when cleaning up.
	if err := c.DeleteVM(ctx, "fc880f74-cf03-4a7a-93da-74c506157023"); err != nil && !idcloudhost.IsNotFound(err) {
		log.Fatal(err)
	}
}
