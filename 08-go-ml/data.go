package main

import (
	"fmt"
	"go-regress/utils"
	"log"
	"os"

	"github.com/go-gota/gota/dataframe"
)

type DataContext struct {
	Data dataframe.DataFrame
	FeatureData map[string][]float64
	TargetValues []float64
}

func LoadAndPrepareData(config Config, logger *log.Logger) (DataContext, error) {
	var dataContext DataContext

	// read data from csv
	dataFile, err := os.Open(config.CSVFilePath)
	if err != nil {
		return dataContext, fmt.Errorf("could not open file: %w", err)
	}

	dataContext.Data = dataframe.ReadCSV(dataFile)

	// display a summary of the data for the user
	printDataSummary(dataContext.Data, logger, "before outlier removal")

	dataContext.Data, err = utils.ValidateData(
		dataContext.Data,
		config.FeatureVars,
		config.TargetVariable,
		config.OutlierLowerBound,
		config.OutlierUpperBound,
	)
	if err != nil {
		return dataContext, fmt.Errorf("data validation error: %w", err)
	}

	if dataContext.Data.Nrow() > 0 {
		printDataSummary(dataContext.Data, logger, "after outlier removal")
	}

	if err := utils.CheckDatasetSize(dataContext.Data.Nrow(), len(config.FeatureVars)); err != nil {
		logger.Printf("Warning: %v\n", err)
	}

	dataContext.FeatureData = make(map[string][]float64)

	for _, feature := range config.FeatureVars {
		featureCol := dataContext.Data.Col(feature)
		featureValues := make([]float64, featureCol.Len())
		for i := range featureCol.Len() {
			featureValues[i] = featureCol.Elem(i).Float()
		}
		dataContext.FeatureData[feature] = featureValues
	}

	targetCol := dataContext.Data.Col(config.TargetVariable)
	dataContext.TargetValues = make([]float64, targetCol.Len())
	for i := range targetCol.Len() {
		dataContext.TargetValues[i] = targetCol.Elem(i).Float()
	}

	return dataContext, nil
}

func printDataSummary(df dataframe.DataFrame, logger *log.Logger, stage string) {
	logger.Printf("Data Preview (%s):\n", stage)
	logger.Println(df.Describe())
	logger.Printf("Columns in dataset: %v\n", df.Names())
	logger.Printf("Row count: %d\n", df.Nrow())

	// show sample rows
	if df.Nrow() > 0 {
		// get the mininum of 3 or the number of columns available
		numCols := min(df.Ncol(), 3)

		// get the mininum of 5 or the number of rows available
		numRows := min(df.Nrow(), 5)

		columnsToShow := make([]int, numCols)
		for i := range numCols {
			columnsToShow[i] = i
		}

		logger.Println(df.Select(columnsToShow).Subset(numRows))
	}
}