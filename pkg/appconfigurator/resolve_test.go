package appconfigurator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveEnv(t *testing.T) {
	t.Setenv("CFG_TEST_TOKEN", "tok-123")
	type c struct {
		A     string `resolve:"env"`
		B     string // untagged literal — must not change
		Empty string `resolve:"env"`
	}
	cfg := c{A: "CFG_TEST_TOKEN", B: "literal", Empty: ""}
	if err := Resolve(&cfg); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if cfg.A != "tok-123" {
		t.Errorf("A = %q, want tok-123", cfg.A)
	}
	if cfg.B != "literal" {
		t.Errorf("untagged field mutated: B = %q", cfg.B)
	}
	if cfg.Empty != "" {
		t.Errorf("Empty = %q, want empty (omitted reference left untouched)", cfg.Empty)
	}
}

func TestResolveNested(t *testing.T) {
	t.Setenv("CFG_TEST_TOKEN", "tok-123")
	type inner struct {
		Ref string `resolve:"env"`
	}
	type outer struct {
		Inner inner
		Ptr   *inner
	}
	cfg := outer{Inner: inner{Ref: "CFG_TEST_TOKEN"}, Ptr: &inner{Ref: "CFG_TEST_TOKEN"}}
	if err := Resolve(&cfg); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if cfg.Inner.Ref != "tok-123" || cfg.Ptr.Ref != "tok-123" {
		t.Errorf("nested not resolved: inner=%q ptr=%q", cfg.Inner.Ref, cfg.Ptr.Ref)
	}
}

func TestResolveFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "cert.pem")
	if err := os.WriteFile(p, []byte("CERTDATA"), 0o600); err != nil {
		t.Fatal(err)
	}
	type c struct {
		Cert File `json:"cert" resolve:"file"`
	}
	cfg := c{Cert: File{Path: p}}
	if err := Resolve(&cfg); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if string(cfg.Cert.Data) != "CERTDATA" {
		t.Errorf("Data = %q, want CERTDATA", cfg.Cert.Data)
	}
	if cfg.Cert.Path != p {
		t.Errorf("Path lost: %q", cfg.Cert.Path)
	}
}

func TestResolveFileMissing(t *testing.T) {
	type c struct {
		Cert File `resolve:"file"`
	}
	cfg := c{Cert: File{Path: "/no/such/file.pem"}}
	if err := Resolve(&cfg); err == nil {
		t.Error("expected error for a missing file")
	}
}

func TestResolveUnknownResolver(t *testing.T) {
	type c struct {
		X string `resolve:"vault"`
	}
	if err := Resolve(&c{X: "REF"}); err == nil {
		t.Error("expected error for an unknown resolver kind")
	}
}

func TestFileJSONRoundTrip(t *testing.T) {
	type c struct {
		Cert File `json:"cert" resolve:"file"`
	}
	var cfg c
	if err := json.Unmarshal([]byte(`{"cert":"/etc/x.pem"}`), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if cfg.Cert.Path != "/etc/x.pem" {
		t.Errorf("Path = %q, want /etc/x.pem", cfg.Cert.Path)
	}
	out, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != `{"cert":"/etc/x.pem"}` {
		t.Errorf("marshal = %s", out)
	}
}
