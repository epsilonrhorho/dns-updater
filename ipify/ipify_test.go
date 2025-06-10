package ipify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGetIPSuccess(t *testing.T) {
	expectedIP := "1.2.3.4"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ip":"` + expectedIP + `"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.Client())
	c.baseURL = srv.URL

	ip, err := c.GetIP(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != expectedIP {
		t.Fatalf("expected %s, got %s", expectedIP, ip)
	}
}

func TestClientGetIPNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClient(srv.Client())
	c.baseURL = srv.URL

	_, err := c.GetIP(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientGetIPInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`invalid`))
	}))
	defer srv.Close()

	c := NewClient(srv.Client())
	c.baseURL = srv.URL

	_, err := c.GetIP(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("expected json syntax error, got %T", err)
	}
}
