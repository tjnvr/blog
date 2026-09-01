package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	url := flag.String("url", "", "Base URL of the deployed site to validate")
	flag.Parse()

	if *url == "" {
		log.Fatal("--url is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(*url)
	if err != nil {
		log.Fatalf("Could not reach %s: %v\n", *url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		log.Fatalf("%s returned HTTP %d\n", *url, resp.StatusCode)
	}

	log.Printf("%s is reachable (HTTP %d)\n", *url, resp.StatusCode)
}
