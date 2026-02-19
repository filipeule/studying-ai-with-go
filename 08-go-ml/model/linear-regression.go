package model

import (
	"fmt"
	"go-regress/utils"
	"math"
	"slices"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

type LinearRegression struct {
	Coefficients    []float64 `json:"coefficients,omitempty"`
	Intercept       float64   `json:"intercept,omitempty"`
	Features        []string  `json:"features,omitempty"`
	Target          string    `json:"target"`
	RSquared        float64   `json:"r_squared,omitempty"`
	FeatureMeans    []float64 `json:"feature_means,omitempty"`
	FeaturesStdDevs []float64 `json:"feature_std_devs,omitempty"`
	IsNormalized    bool      `json:"is_normalized,omitempty"`
	SavedAt         time.Time `json:"-"`
	Description     string    `json:"description,omitempty"`
	NumSamples      int       `json:"num_samples,omitempty"`
	Version         string    `json:"version,omitempty"`
}

func TrainLinearRegression(
	dataFrame dataframe.DataFrame,
	featureNames []string,
	targetName string,
	normalize bool,
) (*LinearRegression, error) {
	// check if all feature columns and target column exists
	columnNames := dataFrame.Names()
	for _, name := range featureNames {
		if !slices.Contains(columnNames, name) {
			return nil, fmt.Errorf("feature column '%s' not found in the dataset", name)
		}
	}

	if !slices.Contains(columnNames, targetName) {
		return nil, fmt.Errorf("target column '%s' not found in the dataset", targetName)
	}

	// get feature columns as float slices
	featureColumns := make([]series.Series, len(featureNames))
	for i, name := range featureNames {
		featureColumns[i] = dataFrame.Col(name)
	}

	// get target columns as float slice
	targetColumn := dataFrame.Col(targetName)

	numSamples := dataFrame.Nrow()
	featureMatrix := make([][]float64, numSamples)
	targetValues := make([]float64, numSamples)

	// fill feature matrix (X) and targer vector (y) with values from dataframe
	for rowIndex := range numSamples {
		featureMatrix[rowIndex] = make([]float64, len(featureNames))

		for colIndex, column := range featureColumns {
			featureMatrix[rowIndex][colIndex] = column.Elem(rowIndex).Float()
		}

		targetValues[rowIndex] = targetColumn.Elem(rowIndex).Float()
	}

	// variables to store normalization params
	var normalizedFeatures [][]float64
	var featureMeans []float64
	var featureStdDev []float64

	if normalize {
		normalizedFeatures, featureMeans, featureStdDev = utils.NormalizeFeatures(featureMatrix)
		featureMatrix = normalizedFeatures
	}

	// create a design matrix
	numFeatures := len(featureNames)

	designMatrix := mat.NewDense(numSamples, numFeatures+1, nil)
	targetVector := mat.NewVecDense(numSamples, nil)

	for rowIndex := range numSamples {
		designMatrix.Set(rowIndex, 0, 1.0)

		for featureIndex := range numFeatures {
			designMatrix.Set(rowIndex, featureIndex+1, featureMatrix[rowIndex][featureIndex])
		}

		targetVector.SetVec(rowIndex, targetValues[rowIndex])
	}

	// step 1 - calculate X^T x (tranpose of X multiplied by x)
	var transposeTimesDesign mat.Dense
	transposeTimesDesign.Mul(designMatrix.T(), designMatrix)

	// step 2 - calculate (X^T X)^(-1)
	var inverseMatrix mat.Dense
	if err := inverseMatrix.Inverse(&transposeTimesDesign); err != nil {
		return nil, fmt.Errorf(
			"failed to compute inverse: %w - matrix may be singular. try add more data or removing highly correlated features",
			err,
		)
	}

	// step 3 - calculate X^T y
	var transposeTimesTarget mat.Dense
	transposeTimesTarget.Mul(designMatrix.T(), targetVector)

	// step 4 - calculate the optimal coefficients
	var coefficientMatrix mat.Dense
	coefficientMatrix.Mul(&inverseMatrix, &transposeTimesTarget)

	// extract coefficients
	interceptAndCoefficients := make([]float64, numFeatures+1)
	for i := range numFeatures + 1 {
		interceptAndCoefficients[i] = coefficientMatrix.At(i, 0)
	}

	// calculate predictions using the trained model
	predictedValues := make([]float64, numSamples)

	for i := range numSamples {
		// start with intercept
		predictedValues[i] = interceptAndCoefficients[0]

		// add contributionn of each feature
		for j := range numFeatures {
			predictedValues[i] += interceptAndCoefficients[j+1] * featureMatrix[i][j]
		}
	}

	// calculate r-squared
	targetMean := stat.Mean(targetValues, nil)

	var totalSumOfSquares, sumOfSquareResiduals float64
	for i := range numSamples {
		totalSumOfSquares += math.Pow(targetValues[i] - targetMean, 2)
		sumOfSquareResiduals += math.Pow(targetValues[i] - predictedValues[i], 2)
	}

	rSquared := 1 - (sumOfSquareResiduals / totalSumOfSquares)

	return &LinearRegression{
		Intercept: interceptAndCoefficients[0],
		Coefficients: interceptAndCoefficients[1:],
		Features: featureNames,
		Target: targetName,
		RSquared: rSquared,
		FeatureMeans: featureMeans,
		FeaturesStdDevs: featureStdDev,
		IsNormalized: normalize,
	}, nil
}

func (l *LinearRegression) PrintModelSummary() {
	fmt.Println("\n==== Model Summary ====")

	fmt.Printf("Regression Equation: %s = %.4f", l.Target, l.Intercept)
	for i, feature := range l.Features {
		if l.Coefficients[i] >= 0 {
			fmt.Printf(" + %.4f x %s", l.Coefficients[i], feature)
		} else {
			fmt.Printf(" - %.4f x %s", -l.Coefficients[i], feature)
		}
	}

	fmt.Println()

	// display model fit statistics
	fmt.Printf("\nModel Performance:\n")
	fmt.Printf("- R-squared: %.4f\n", l.RSquared)
	fmt.Printf(
		"- Interpretation: %.2f%% of variance in %s is explained by this model",
		l.RSquared * 100,
		l.Target,
	)

	fmt.Printf("\nCoefficient Interpretation:\n")
	fmt.Printf("- Intercept (%.4f): The base %s when all features are zero\n", l.Intercept, l.Target)

	for i, feature := range l.Features {
		fmt.Printf(
			"- %s Coefficient (%.4f): for eac additional unit of %s, %s changes by %.4f units\n",
			feature,
			l.Coefficients[i],
			feature,
			l.Target,
			l.Coefficients[i],
		)
	}

	if l.IsNormalized {
		fmt.Printf("\nNote: this model was trained on normalized data. Predictions on new data will automatically be normalized.\n")
	}
}

func (l *LinearRegression) Predict(featuresValues [][]float64) []float64 {
	predictions := make([]float64, len(featuresValues))

	for dataPointIndex, featureRow := range featuresValues {
		// start with intercept
		predictedValue := l.Intercept

		normalizedFeatures := make([]float64, len(featureRow))
		copy(normalizedFeatures, featureRow)

		if l.IsNormalized && len(l.FeatureMeans) == len(featureRow) {
			for i := range normalizedFeatures {
				if l.FeaturesStdDevs[i] > 0 {
					normalizedFeatures[i] = (featureRow[i] - l.FeatureMeans[i]) / l.FeaturesStdDevs[i]
				} else {
					normalizedFeatures[i] = 0
				}
			}
		}

		// add contribuition of each feature
		for featureIndex, coefficient := range l.Coefficients {
			if l.IsNormalized {
				predictedValue += coefficient * normalizedFeatures[featureIndex]
			} else {
				predictedValue += coefficient * featureRow[featureIndex]
			}
		}

		predictions[dataPointIndex] = predictedValue
	}

	return predictions
}