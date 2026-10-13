package biznetgio

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestCreateNeoLite(t *testing.T) {
	c, seen := fakeAPI(t, map[string]func() (int, string){
		"POST /neolites": ok(`{"success":true,"code":200,"data":{"order_id":"77","account_id":"1234"}}`),
	})
	order, err := c.CreateNeoLite(context.Background(), CreateNeoLiteRequest{
		ProductID: 1559, Cycle: CycleMonthly, OS: "ubuntu-24.04", KeypairID: 9,
		User: "raklane", ConsolePassword: "pw", VMName: "node-1", PayWithCreditCard: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if order.AccountID != 1234 || order.OrderID != "77" {
		t.Fatalf("order = %+v", order)
	}
	got := (*seen)[0]
	if got.Token != "secret" {
		t.Errorf("x-token = %q", got.Token)
	}
	want := map[string]any{
		"product_id": float64(1559), "cycle": "m", "select_os": "ubuntu-24.04", "keypair_id": float64(9),
		"ssh_and_console_user": "raklane", "console_password": "pw", "vm_name": "node-1", "pay_invoice_with_cc": true,
	}
	for k, v := range want {
		if got.JSON[k] != v {
			t.Errorf("body[%s] = %v, want %v", k, got.JSON[k], v)
		}
	}
	if _, set := got.JSON["promocode"]; set {
		t.Error("empty promocode was sent")
	}
}

func TestGetNeoLiteKeepsRaw(t *testing.T) {
	c, _ := fakeAPI(t, map[string]func() (int, string){
		"GET /neolites/accounts/1234": ok(`{"success":true,"code":200,"data":{
			"account_id":1234,"status":"Active","billingcycle":"Monthly","product_id":"1559",
			"extra_details":{"region":"jkt","ciuser":"raklane","neosshkey_id":"9"},
			"ip_address":"203.0.113.7"}}`),
	})
	a, err := c.GetNeoLite(context.Background(), 1234)
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != StatusActive || a.ProductID != 1559 || a.Details.KeypairID != 9 || a.Details.User != "raklane" {
		t.Fatalf("account = %+v", a)
	}
	if !strings.Contains(string(a.Raw), "203.0.113.7") {
		t.Errorf("raw = %s", a.Raw)
	}
}

func TestListNeoLitesFiltersByStatus(t *testing.T) {
	c, seen := fakeAPI(t, map[string]func() (int, string){
		"GET /neolites/accounts": ok(`{"success":true,"data":[{"account_id":1},{"account_id":"2"}]}`),
	})
	accounts, err := c.ListNeoLites(context.Background(), StatusActive)
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 2 || accounts[1].AccountID != 2 {
		t.Fatalf("accounts = %+v", accounts)
	}
	if (*seen)[0].Query.Get("status") != "Active" {
		t.Errorf("query = %v", (*seen)[0].Query)
	}
}

func TestDeleteNeoLiteNotFound(t *testing.T) {
	c, _ := fakeAPI(t, map[string]func() (int, string){
		"DELETE /neolites/1234": func() (int, string) {
			return http.StatusBadRequest, `{"code":400,"detail":"Account not found"}`
		},
	})
	err := c.DeleteNeoLite(context.Background(), 1234)
	if !IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
	if !strings.Contains(err.Error(), "Account not found") {
		t.Errorf("err = %v", err)
	}
}

func TestSetNeoLiteState(t *testing.T) {
	c, _ := fakeAPI(t, map[string]func() (int, string){
		"PUT /neolites/accounts/1234/vm-state/stop": ok(`{"success":true,"code":200,"data":null}`),
	})
	if err := c.SetNeoLiteState(context.Background(), 1234, StateStop); err != nil {
		t.Fatal(err)
	}
}

func TestGetNeoLiteVM(t *testing.T) {
	c, _ := fakeAPI(t, map[string]func() (int, string){
		"GET /neolites/accounts/1234/vm-details": ok(`{"success":true,"data":{"vmid":101,"status":"running","cpus":2}}`),
	})
	vm, err := c.GetNeoLiteVM(context.Background(), 1234)
	if err != nil {
		t.Fatal(err)
	}
	if vm.VMID != 101 || vm.Status != "running" || vm.CPUs != 2 {
		t.Fatalf("vm = %+v", vm)
	}
}

func TestIDRejectsNonNumbers(t *testing.T) {
	var id ID
	if err := id.UnmarshalJSON([]byte(`"abc"`)); err == nil {
		t.Fatal("want an error")
	}
	if err := id.UnmarshalJSON([]byte(`null`)); err != nil || id != 0 {
		t.Fatalf("null: id=%d err=%v", id, err)
	}
}
