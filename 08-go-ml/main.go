package main

import (
	"go-regress/model"
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
	dataModel, dataContext, err := GetOrTrainModel(config, logger)
	if err != nil {
		logger.Fatalf("model error: %v\n", err)
	}

	// save model if requested
	if config.SaveModelPath != "" {
		if err := model.SaveModelToJSON(
			dataModel,
			config.SaveModelPath,
			config.ModelDesc,
			dataContext.Data.Nrow(),
		); err != nil {
			logger.Fatalf("error saving model: %v\n", err)
		}
	}
}
