package main

import (
	"flag"
	"fmt"
	"go-regress/model"
	"log"
)

func GetOrTrainModel(config Config, logger *log.Logger) (*model.LinearRegression, *DataContext, error) {
	var dataModel *model.LinearRegression
	var dataContext DataContext
	var err error

	// application can either load a saved model or train a new one
	if config.LoadModelPath != "" {
		// load an existing model from json file
		dataModel, err = model.LoadModelFromJSON(config.LoadModelPath)
		if err != nil {
			return nil, nil, fmt.Errorf("error loading model: %w", err)
		}

		return dataModel, &dataContext, nil
	}

	// training a new model from a csv file
	if config.CSVFilePath == "" {
		flag.Usage()
		return nil, nil, fmt.Errorf("please provide a path to the csv file using the -file flag")
	}

	// load and prepare training data
	dataContext, err = LoadAndPrepareData(config, logger)
	if err != nil {
		return nil, nil, err
	}

	// train linear regression using the data
	logger.Printf("training model with features: %v\n", config.FeatureVars)
	dataModel, err = model.TrainLinearRegression(dataContext.Data, config.FeatureVars, config.TargetVariable, config.Normalize)
	if err != nil {
		return nil, nil, fmt.Errorf("error training model: %w", err)
	}

	dataModel.PrintModelSummary()

	return dataModel, &dataContext, nil
}
