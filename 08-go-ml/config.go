package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	CSVFilePath       string
	SaveModelPath     string
	LoadModelPath     string
	TargetVariable    string
	ModelDesc         string
	Normalize         bool
	FeatureVars       []string
	DataToPredict     string
	OutlierLowerBound float64
	OutlierUpperBound float64
	Plot              bool
	PlotURI           string
}

func parseCommandLineArgs() Config {
	var config Config
	var features string
	var target string

	flag.StringVar(&config.CSVFilePath, "file", "house_data.csv", "path to csv file containing data")
	flag.StringVar(&config.SaveModelPath, "save", "", "path to save the trained model")
	flag.StringVar(&config.LoadModelPath, "load", "", "path to load a previously trained model")
	flag.StringVar(&config.ModelDesc, "desc", "", "description of the model (used when saving)")
	flag.BoolVar(&config.Normalize, "normalize", true, "normalize features (default: true)")
	flag.BoolVar(&config.Plot, "plot", false, "generate a plot (default: false)")
	flag.StringVar(&features, "features", "", "comma separated list of features that you want to use")
	flag.StringVar(&target, "target", "", "name of the target column")
	flag.Float64Var(&config.OutlierLowerBound, "lower-bound", 1.5, "lower bound multiplier for outlier detection (default: 1.5)")
	flag.Float64Var(&config.OutlierUpperBound, "upper-bound", 1.5, "upper bound multiplier for outlier detection (default: 1.5)")
	flag.StringVar(&config.DataToPredict, "predict", "", "enter key value pairs for prediction, e.g. square_footage=1000,bedrooms=2")
	flag.StringVar(&config.PlotURI, "plot-uri", "http://localhost:8000", "uri for the plot app")

	flag.Parse()

	if (len(features) == 0 || len(target) == 0) && config.LoadModelPath == "" {
		fmt.Println("you must specify at least one feature and one target")
		os.Exit(1)
	}

	if config.OutlierLowerBound < 0 || config.OutlierUpperBound < 0 {
		fmt.Println("outlier bound multipliers must be positive values")
		os.Exit(1)
	}

	var featureList []string
	for f := range strings.SplitSeq(features, ",") {
		featureList = append(featureList, strings.TrimSpace(f))
	}

	config.FeatureVars = featureList

	if len(target) > 0 {
		config.TargetVariable = target
	}

	return config
}
