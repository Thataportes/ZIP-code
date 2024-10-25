package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Struct to capture address data returned by the APIs
type Address struct {
	ZIPCode      string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

// Function to make a request to the first API (BrasilAPI)
func fetchBrasilAPI(ctx context.Context, zipCode string, ch chan<- Address) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", zipCode)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request for BrasilAPI:", err)
		close(ch)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error calling BrasilAPI:", err)
		close(ch)
		return
	}
	defer resp.Body.Close()

	var address Address
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		fmt.Println("Error decoding BrasilAPI response:", err)
		close(ch)
		return
	}

	address.ZIPCode += " (BrasilAPI)"
	ch <- address
}

// Function to make a request to the second API (ViaCEP)
func fetchViaCEP(ctx context.Context, zipCode string, ch chan<- Address) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", zipCode)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request for ViaCEP:", err)
		close(ch)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error calling ViaCEP:", err)
		close(ch)
		return
	}
	defer resp.Body.Close()

	var address Address
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		fmt.Println("Error decoding ViaCEP response:", err)
		close(ch)
		return
	}

	address.ZIPCode += " (ViaCEP)"
	ch <- address
}

func main() {
	zipCode := "01153000"
	// Set a timeout context of 1 second
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create a channel to receive the fastest response
	ch := make(chan Address)

	// Start requests in separate goroutines
	go fetchBrasilAPI(ctx, zipCode, ch)
	go fetchViaCEP(ctx, zipCode, ch)

	// Wait for the fastest response or timeout
	select {
	case address := <-ch:
		fmt.Printf("Fastest result received: %+v\n", address)
	case <-ctx.Done():
		fmt.Println("Timeout error: no response received within the time limit.")
	}
}
