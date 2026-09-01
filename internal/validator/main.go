package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	url := flag.String("url", "", "Base URL of the deployed site to validate")
	flag.Parse()

	if *url == "" {
		log.Fatal("--url is required")
	}

	warnings, err := run(*url)
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, "WARNING:", w)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	log.Printf("%s is valid\n", *url)
}
