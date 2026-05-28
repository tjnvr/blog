package main

import (
	"flag"
	"log"

	"github.com/tjnvr/blog/internal/generator/site"

	"github.com/spf13/afero"
)

func main() {
	var (
		contentDir    = flag.String("content-dir", "./content/markdown", "The directory where are the markdown pages")
		buildDir      = flag.String("build-dir", "./target/build", "The directory where is stored the generated site")
		assetsDir     = flag.String("assets-dir", "./content/assets/", "The directory where are the site assets")
		scriptsDir    = flag.String("scripts-dir", "./scripts/", "The directory where are the site scripts")
		assetsOutDir  = flag.String("assets-out-dir", "./target/build/assets", "The directory where are stored the assets of the generated site")
		scriptsOutDir = flag.String("scripts-out-dir", "./target/build/scripts", "The directory where are stored the scripts of the generated site")

		skipURLValidation = flag.Bool("skip-url-validation", false, "Skip external URL validation")
	)

	flag.Parse()

	gen, err := site.NewGenerator(afero.NewOsFs(), site.WithSkipURLValidation(*skipURLValidation))
	if err != nil {
		log.Fatalf("Could not create the site generator: %v\n", err)
	}

	if err := gen.Generate(*contentDir, *buildDir, *assetsDir, *assetsOutDir, *scriptsDir, *scriptsOutDir); err != nil {
		log.Fatalf("Site generation error: %v\n", err)
	}

	if err := gen.Validate(); err != nil {
		log.Fatalf("Site validation error: %v\n", err)
	}

	log.Println("Site generated successfully !")
}
