package main

import (
	"flag"
	"log"
	"os"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/generator/site"
)

func main() {
	skipURLValidation := flag.Bool("skip-url-validation", false, "Skip external URL validation")
	flag.Parse()

	cfg := site.Config{
		AssetsDir:     os.Getenv("ASSETS_DIR"),
		AssetsOutDir:  os.Getenv("ASSETS_OUT_DIR"),
		ContentDir:    os.Getenv("CONTENT_DIR"),
		BuildDir:      os.Getenv("BUILD_DIR"),
		ScriptsDir:    os.Getenv("SCRIPTS_DIR"),
		ScriptsOutDir: os.Getenv("SCRIPTS_OUT_DIR"),
	}

	gen, err := site.NewGenerator(afero.NewOsFs(), cfg, site.WithSkipURLValidation(*skipURLValidation))

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
