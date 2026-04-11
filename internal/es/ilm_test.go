package es

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListILMPolicies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"my-policy": {
				"version": 1,
				"modified_date": "2024-01-01",
				"policy": {
					"phases": {
						"hot": {"min_age": "0ms", "actions": {"rollover": {"max_age": "30d"}}},
						"delete": {"min_age": "90d", "actions": {"delete": {}}}
					}
				}
			},
			"other-policy": {
				"version": 2,
				"policy": {"phases": {"warm": {"min_age": "7d", "actions": {}}}}
			}
		}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	policies, err := c.ListILMPolicies()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policies) != 2 {
		t.Fatalf("expected 2 policies, got %d", len(policies))
	}

	// Find my-policy
	var myPolicy *ILMPolicy
	for i := range policies {
		if policies[i].Name == "my-policy" {
			myPolicy = &policies[i]
			break
		}
	}
	if myPolicy == nil {
		t.Fatal("expected my-policy")
	}
	if myPolicy.Version != 1 {
		t.Fatalf("expected version 1, got %d", myPolicy.Version)
	}
	if _, ok := myPolicy.Phases["hot"]; !ok {
		t.Fatal("expected hot phase")
	}
	if _, ok := myPolicy.Phases["delete"]; !ok {
		t.Fatal("expected delete phase")
	}
	if myPolicy.Phases["delete"].MinAge != "90d" {
		t.Fatalf("expected 90d, got %s", myPolicy.Phases["delete"].MinAge)
	}
}

func TestListILMPolicies_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	policies, err := c.ListILMPolicies()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policies) != 0 {
		t.Fatalf("expected 0 policies, got %d", len(policies))
	}
}

func TestCreateILMPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	err := c.CreateILMPolicy("test-policy", `{"policy":{"phases":{"delete":{"min_age":"30d","actions":{"delete":{}}}}}}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteILMPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	err := c.DeleteILMPolicy("test-policy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
