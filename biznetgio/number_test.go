package biznetgio

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestIntReadsNumbersAndStrings(t *testing.T) {
	for in, want := range map[string]Int{
		`1`: 1, `"1"`: 1, `" 42 "`: 42, `"59000.00"`: 59000, `59000.5`: 59001,
		`1073741824`: 1073741824, `""`: 0, `null`: 0,
	} {
		var n Int
		if err := json.Unmarshal([]byte(in), &n); err != nil {
			t.Errorf("%s: %v", in, err)
		} else if n != want {
			t.Errorf("%s = %d, want %d", in, n, want)
		}
	}
	for _, in := range []string{`"abc"`, `true`, `"1,5"`} {
		var n Int
		if err := json.Unmarshal([]byte(in), &n); err == nil {
			t.Errorf("%s: want an error, got %d", in, n)
		}
	}
}

func fixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// Real responses: products send every number as a string, operating
// systems as numbers.
func TestRealProductsResponse(t *testing.T) {
	c, _ := fakeAPI(t, map[string]func() (int, string){"GET /neolites/products": ok(fixture(t, "products.json"))})
	products, err := c.ListNeoLiteProducts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 2 {
		t.Fatalf("products = %+v", products)
	}
	p := products[1]
	if p.ProductID != 740 || p.Options.Cores != 2 || p.Options.Memory != 4096 {
		t.Errorf("product = %+v", p)
	}
	if price, ok := p.Price(CycleMonthly); !ok || price != 139000 {
		t.Errorf("monthly = %d, %v", price, ok)
	}
	disk := products[0].Billing[0].Components[0]
	if disk.Field != "disk_size" || disk.Prices[0].Price != 1650 || disk.Prices[0].QtyMax != 1000 {
		t.Errorf("disk component = %+v", disk)
	}
}

func TestRealOSResponse(t *testing.T) {
	c, _ := fakeAPI(t, map[string]func() (int, string){"GET /neolites/products/731/oss": ok(fixture(t, "oss.json"))})
	oss, err := c.ListNeoLiteOS(context.Background(), 731)
	if err != nil {
		t.Fatal(err)
	}
	if len(oss) != 2 || oss[0].Name != "Ubuntu-24.04" || oss[0].VMID != 1014 || oss[0].MaxMem != 1073741824 || oss[0].MaxCPU != 1 {
		t.Fatalf("oss = %+v", oss)
	}
}
