package main

import (
	"log"
	"os"
)

func main() {
	// parse command line arguments
	config := parseCommandLineArgs()

	// setup a logger
	logger := log.New(os.Stdout, "", log.LstdFlags)

	logger.Println("parsed command line flags:", config.FeatureVars)

	// either load or train a model
}