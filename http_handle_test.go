package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)


func TestHttpHandleRequests(t *testing.T) {
	handler := http.HandlerFunc(apihandler)

	t.Run("GET /api", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected, _ := json.Marshal(domainsToAddresses)
		if rr.Body.String() != string(expected)+"\n" {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), string(expected))
		}
	})

	t.Run("POST /api", func(t *testing.T) {
		newDomain := map[string]string{"example.com": "192.168.1.1"}
		jsonDomain, _ := json.Marshal(newDomain)
		req, _ := http.NewRequest("POST", "/api", bytes.NewBuffer(jsonDomain))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
		}

		mu.RLock()
		defer mu.RUnlock()
		if domainsToAddresses["example.com"] != "192.168.1.1" {
			t.Errorf("handler did not add domain: got %v want %v", domainsToAddresses["example.com"], "192.168.1.1")
		}
	})

	t.Run("PATCH /api", func(t *testing.T) {
		newDomain := map[string]string{"example.com": "192.168.1.1"}
		jsonDomain, _ := json.Marshal(newDomain)
		req, _ := http.NewRequest("POST", "/api", bytes.NewBuffer(jsonDomain))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		updatedDomain := map[string]string{"example.com": "192.168.1.2"}
		jsonDomain_patch, _ := json.Marshal(updatedDomain)
		req_patch, _ := http.NewRequest("PATCH", "/api", bytes.NewBuffer(jsonDomain_patch))
		rr_patch := httptest.NewRecorder()
		handler.ServeHTTP(rr_patch, req_patch)

		if status := rr_patch.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		mu.RLock()
		defer mu.RUnlock()
		if domainsToAddresses["example.com"] != "192.168.1.2" {
			t.Errorf("handler did not update domain: got %v want %v", domainsToAddresses["example.com"], "192.168.1.2")
		}
	})

	t.Run("DELETE /api", func(t *testing.T) {
		// First, add a domain to delete
		newDomain := map[string]string{"example.com": "192.168.1.1"}
		jsonDomain, _ := json.Marshal(newDomain)
		req, _ := http.NewRequest("POST", "/api", bytes.NewBuffer(jsonDomain))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Now, delete the domain
		req, _ = http.NewRequest("DELETE", "/api?domain=example.com", nil)
		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		mu.RLock()
		defer mu.RUnlock()
		if _, exists := domainsToAddresses["example.com"]; exists {
			t.Errorf("handler did not delete domain: example.com still exists")
		}
	})

	t.Run("DELETE /api by address", func(t *testing.T) {
		// First, add a domain to delete
		newDomain := map[string]string{"example.net": "192.168.1.2"}
		jsonDomain, _ := json.Marshal(newDomain)
		req, _ := http.NewRequest("POST", "/api", bytes.NewBuffer(jsonDomain))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Now, delete the domain by address
		req, _ = http.NewRequest("DELETE", "/api?address=192.168.1.2", nil)
		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		mu.RLock()
		defer mu.RUnlock()
		if _, exists := domainsToAddresses["example.net"]; exists {
			t.Errorf("handler did not delete domain: example.net still exists")
		}
	})

}
