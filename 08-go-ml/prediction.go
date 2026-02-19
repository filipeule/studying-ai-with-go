package main

import (
	"fmt"
	"go-regress/model"
	"strconv"
	"strings"
)

func handlePrediction(config Config, dataModel *model.LinearRegression) {
	// sanity check
	if len(config.DataToPredict) == 0 {
		return
	}

	// parse the key values string into a map
	kvMap := make(map[string]string)
	pairs := strings.SplitSeq(config.DataToPredict, ",")
	for pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 {
			kvMap[kv[0]] = kv[1]
		}
	}

	// make sure the requested features are the same in kind and number as the features in our trained model
	if len(kvMap) != len(dataModel.Features) {
		fmt.Println("cannot do prediction of new data: wrong number of features specified")
		return
	} else {
		// make sure our requested features for predictions match the features field
		for _, feature := range dataModel.Features {
			if kvMap[feature] == "" {
				fmt.Println(
					"cannot do prediction of new data: incorrect features requested. use:",
					strings.Join(dataModel.Features, ","),
				)
				return
			}
		}
	}

	// create an array of floats in the same order as features in the model
	// this ensures the values match up with the correct coefficients
	values := make([]float64, len(dataModel.Features))
	for i, feature := range dataModel.Features {
		if val, ok := kvMap[feature]; ok {
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				values[i] = f
			}
		}
	}

	// prepare the input for prediction
	var newData [][]float64
	newData = append(newData, values)

	// use this to make predictions
	predictions := dataModel.Predict(newData)

	// display the results
	fmt.Println("\nPredictions:")
	displayPredictionTable(newData, predictions, dataModel)
}

func displayPredictionTable(newData [][]float64, predictions []float64, dataModel *model.LinearRegression) {
	const colWidth = 15

	for i, feature := range dataModel.Features {
		if i == 0 {
			fmt.Printf("%-*s", colWidth, feature)
		} else {
			fmt.Printf(" | %-*s", colWidth, feature)
		}
	}

	fmt.Printf(" | %-*s\n", colWidth, "Predicted Value")

	// print a separated line with appropriate length
	totalWidth := (colWidth+3)*(len(dataModel.Features)+1) + 10
	fmt.Println(strings.Repeat("-", totalWidth))

	// print each item with its predicted price
	for i, item := range newData {
		// format each feature value
		fmt.Printf("%-*.2f", colWidth, item[0])

		for j := 0; j < len(item); j++ {
			fmt.Printf(" | %-*.2f", colWidth, item[j])
		}

		// add the prediction
		fmt.Printf(" | $%-*.2f", colWidth-1, predictions[i])
	}

	fmt.Println("\nNote: These predictions are based on the trained linear regression model")
}
