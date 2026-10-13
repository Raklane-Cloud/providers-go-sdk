package biznetgio

import (
	"context"
	"testing"
)

func TestListNeoLiteProducts(t *testing.T) {
	c, _ := fakeAPI(t, map[string]func() (int, string){
		"GET /neolites/products": ok(`{"success":true,"data":[{"product_id":1559,"name":"NEO Lite XS 2.2",
			"options":{"cores":2,"memory":2},
			"billing":[{"label":"Monthly","cycle":"m","price":150000},{"label":"Annually","cycle":"a","price":1500000}]}]}`),
	})
	products, err := c.ListNeoLiteProducts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 1 || products[0].Options.Cores != 2 {
		t.Fatalf("products = %+v", products)
	}
	if price, ok := products[0].Price(CycleMonthly); !ok || price != 150000 {
		t.Errorf("monthly = %d, %v", price, ok)
	}
	if _, ok := products[0].Price(CycleBiennial); ok {
		t.Error("biennial shouldn't be on sale")
	}
}

func TestListNeoLiteOS(t *testing.T) {
	c, _ := fakeAPI(t, map[string]func() (int, string){
		"GET /neolites/products/1559/oss": ok(`{"success":true,"data":[{"name":"ubuntu-24.04","vmid":"9001"}]}`),
	})
	oss, err := c.ListNeoLiteOS(context.Background(), 1559)
	if err != nil {
		t.Fatal(err)
	}
	if len(oss) != 1 || oss[0].Name != "ubuntu-24.04" || oss[0].VMID != 9001 {
		t.Fatalf("oss = %+v", oss)
	}
}
