package main

import (
	"flag"
	"log"

	"github.com/tjnvr/blog/internal/generator/site"

	"github.com/spf13/afero"
)

func main() {
	skipURLValidation := flag.Bool("skip-url-validation", false, "Skip external URL validation")
	flag.Parse()

	gen, err := site.NewGenerator(afero.NewOsFs(), site.WithSkipURLValidation(*skipURLValidation))
	if err != nil {
		log.Fatalf("Could not create the site generator: %v\n", err)
	}

	if err := gen.Generate(); err != nil {
		log.Fatalf("Site generation error: %v\n", err)
	}

	if err := gen.Validate(); err != nil {
		log.Fatalf("Site validation error: %v\n", err)
	}

	log.Println("Site generated successfully !")
}
