package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// apihandler handles HTTP requests to the "/api" endpoint.
// It supports GET, POST, PATCH, and DELETE methods to manage the domainsToAddresses map.
//
// Parameters:
//   w (http.ResponseWriter): The response writer to send responses to the client.
//   r (*http.Request): The incoming HTTP request.
func apihandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Handle GET request: Return the current domainsToAddresses map as JSON.
		mu.RLock() // Read lock
		defer mu.RUnlock()
		json.NewEncoder(w).Encode(domainsToAddresses)
	} else if r.Method == http.MethodPost {
		// Handle POST request: Add new domain-address pairs to the map.
		mu.Lock() // Write lock
		defer mu.Unlock()
		var newDomain map[string]string
		if err := json.NewDecoder(r.Body).Decode(&newDomain); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for domain, address := range newDomain {
			if domain[len(domain)-1] == '.' {
				domain = domain[:len(domain)-1]
			}
			domainsToAddresses[domain] = address
		}
		w.WriteHeader(http.StatusCreated)
	} else if r.Method == http.MethodPatch {
		// Handle PATCH request: Update existing domain-address pairs in the map.
		mu.Lock() // Write lock
		defer mu.Unlock()
		var updatedDomain map[string]string
		if err := json.NewDecoder(r.Body).Decode(&updatedDomain); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for domain, address := range updatedDomain {
			if domain[len(domain)-1] == '.' {
				domain = domain[:len(domain)-1]
			}
			if _, exists := domainsToAddresses[domain]; exists {
				domainsToAddresses[domain] = address
			} else {
				http.Error(w, "Domain not found", http.StatusNotFound)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	} else if r.Method == http.MethodDelete {
		// Handle DELETE request: Remove domain-address pairs from the map based on query parameters.
		mu.Lock() // Write lock
		defer mu.Unlock()
		queryParams := r.URL.Query()
		del_domain := queryParams.Get("domain")
		if del_domain != "" {
			if del_domain[len(del_domain)-1] == '.' {
				del_domain = del_domain[:len(del_domain)-1]
			}
		}
		del_address := queryParams.Get("address")

		for domain, address := range domainsToAddresses {
			if domain == del_domain || address == del_address {
				delete(domainsToAddresses, domain)
			}
		}
	} else {
		// Handle unsupported HTTP methods.
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}


// http_handleRequests starts an HTTP server on the specified port and handles requests to the "/api" endpoint.
// It registers the apihandler function to handle requests to the "/api" endpoint and runs the server in a separate goroutine.
// If the server encounters a fatal error, it logs the error and terminates the program.
//
// Parameters:
//   port (int): The port number on which the HTTP server will listen for incoming requests.
func http_handleRequests(port int) {
	http.HandleFunc("/api", apihandler)
	go func() {
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
	}()
}
