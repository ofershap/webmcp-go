package webmcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterValidation(t *testing.T) {
	m := NewMux()
	if err := m.Register(Tool{Name: "Bad Name!"}); err == nil {
		t.Fatal("want invalid-name error")
	}
	if err := m.Register(Tool{Name: "ok-name"}); err == nil {
		t.Fatal("want no-handler error")
	}
	tool := Tool{
		Name:    "filter-templates",
		Handler: func(input json.RawMessage) (any, error) { return nil, nil },
	}
	if err := m.Register(tool); err != nil {
		t.Fatal(err)
	}
	if err := m.Register(tool); err == nil {
		t.Fatal("want duplicate error")
	}
}

func TestRegisterJSShape(t *testing.T) {
	m := NewMux()
	err := m.Register(Tool{
		Name:        "order-prints",
		Description: "Orders prints.",
		InputSchema: ObjectSchema(
			Integer("copies", "number of copies").Req(),
			Enum("pageSize", "paper size", "Letter", "A4"),
		),
		Handler: func(input json.RawMessage) (any, error) { return map[string]int{"queued": 1}, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(m.ServeHTTP("/.well-known/webmcp"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/.well-known/webmcp/register.js")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	buf := make([]byte, 8192)
	n, _ := resp.Body.Read(buf)
	js := string(buf[:n])
	for _, want := range []string{"modelContext", "registerTool", "order-prints", "/.well-known/webmcp/invoke/", `"copies"`, `"pageSize"`} {
		if !strings.Contains(js, want) {
			t.Errorf("register.js missing %q", want)
		}
	}
}

func TestInvokeRoundTrip(t *testing.T) {
	m := NewMux()
	var got struct {
		Copies int `json:"copies"`
	}
	if err := m.Register(Tool{
		Name: "order-prints",
		Handler: func(input json.RawMessage) (any, error) {
			if err := json.Unmarshal(input, &got); err != nil {
				return nil, err
			}
			return map[string]string{"status": "queued"}, nil
		},
	}); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(m.ServeHTTP("/webmcp"))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/webmcp/invoke/order-prints", "application/json", strings.NewReader(`{"copies":3}`))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if got.Copies != 3 {
		t.Fatalf("handler got %+v", got)
	}
	var out map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["status"] != "queued" {
		t.Fatalf("result %v", out)
	}
}

func TestInvokeUnknownTool(t *testing.T) {
	m := NewMux()
	srv := httptest.NewServer(m.ServeHTTP("/webmcp"))
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/webmcp/invoke/nope", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", resp.StatusCode)
	}
}

func TestScriptTag(t *testing.T) {
	tag := ScriptTag("/webmcp")
	if !strings.Contains(tag, `/webmcp/register.js`) || !strings.HasPrefix(tag, "<script") {
		t.Fatalf("tag %q", tag)
	}
}
