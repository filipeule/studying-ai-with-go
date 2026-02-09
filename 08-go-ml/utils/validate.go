package utils

import (
	"fmt"
	"math"
	"slices"
	"sort"

	"github.com/go-gota/gota/dataframe"
)

func ValidateData(
	df dataframe.DataFrame,
	features []string,
	target string,
	lowerBoundMultiplier,
	upperBoundMultiplier float64,
) (dataframe.DataFrame, error) {
	// get all columns and required columns
	allColumns := df.Names()
	requiredColumns := slices.Clone(features)
	requiredColumns = append(requiredColumns, target)

	for _, col := range requiredColumns {
		if !slices.Contains(allColumns, col) {
			return df, fmt.Errorf("column %s not found in the dataset", col)
		}
	}

	// validate all columns for missing and negative values
	for _, colName := range requiredColumns {
		col := df.Col(colName)
		for i := range col.Len() {
			value := col.Elem(i).Float()
			if math.IsNaN(value) {
				return df, fmt.Errorf("missing value found in column %s at row %d", colName, i+1)
			}

			if value < 0 {
				return df, fmt.Errorf("negative value found in column %s at row %d", colName, i+1)
			}
		}
	}

	// track valid rows (non-outliers)
	validRows := make([]bool, df.Nrow())
	for i := range validRows {
		validRows[i] = true
	}

	// find outliers in all columns at once
	outlierCount := 0
	for _, colName := range requiredColumns {
		values := df.Col(colName).Float()
		if len(values) < 4 {
			continue
		}

		// calculate quartiles and IQR
		sortedVals := make([]float64, len(values))
		copy(sortedVals, values)
		sort.Float64s(sortedVals)

		n := len(sortedVals)
		q1, q3 := sortedVals[n/4], sortedVals[(3*n)/4]
		iqr := q3 - q1

		// define my bounds
		lowerBound := q1 - lowerBoundMultiplier*iqr
		upperBound := q3 + upperBoundMultiplier*iqr

		// identify outliners
		for i, v := range values {
			if v < lowerBound || v > upperBound {
				validRows[i] = false
				outlierCount++
				// only log a few to avoid console flooding
				if outlierCount <= 3 {
					fmt.Printf("  - removing outlier in '%s' at row '%d': %.2f (outside range %.2f - %.2f)\n", colName, i+1, v, lowerBound, upperBound)
				}
			}
		}
	}

	// build a list of row indices to keep
	rowsToKeep := make([]int, 0, df.Nrow())
	for i, isValid := range validRows {
		if isValid {
			rowsToKeep = append(rowsToKeep, i)
		}
	}

	// print information about dropped rows, if any
	if outlierCount > 0 {
		fmt.Printf("removed %d outlier records (%.1f%% of data)\n", outlierCount, 100*float64(outlierCount)/float64(df.Nrow()))
	}

	// return filtered df if there are rows to drop
	if len(rowsToKeep) < df.Nrow() {
		return df.Subset(rowsToKeep), nil
	}

	return df, nil
}

func CheckDatasetSize(numSamples, numFeatures int) error {
	minRecommendedSamples := numFeatures * 20

	if numSamples < numFeatures+2 {
		return fmt.Errorf("dataset has too few samples (%d) for the number of features (%d) - model will be overfitted", numSamples, numFeatures)
	}
	
	if numSamples < minRecommendedSamples {
		return fmt.Errorf("dataset size (%d samples) is smaller than the recommended (%d samples) for %d features - results may be unreliable", numSamples, minRecommendedSamples, numFeatures)
	}

	return nil
}
