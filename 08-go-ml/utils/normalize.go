package utils

import "math"

func NormalizeFeatures(X [][]float64) ([][]float64, []float64, []float64) {
	// handle empty input case
	if len(X) == 0 || len(X[0]) == 0 {
		return X, []float64{}, []float64{}
	}

	numSamples := len(X)
	numFeatures := len(X[0])

	// initialize slices to store the mean and standard deviation for each features
	means := make([]float64, numFeatures)
	stdDevs := make([]float64, numFeatures)
	normalizedX := make([][]float64, numSamples)

	for i := range normalizedX {
		normalizedX[i] = make([]float64, numFeatures)
	}

	// calculate means and standard deviations in a single pass per feature
	for j := range numFeatures {
		// calculate mean
		var sum float64
		for i := range numSamples {
			sum += X[i][j]
		}
		means[j] = sum / float64(numSamples)

		// calculate standard deviation
		var varianceSum float64
		for i := range numSamples {
			diff := X[i][j] - means[j]
			varianceSum += diff * diff
		}

		// prevent division by zero with small epsilon
		epsilon := 1e-10
		stdDevs[j] = math.Max(math.Sqrt(varianceSum/float64(numSamples)), epsilon)

		// normalize values for this feature
		for i := range numSamples {
			normalizedX[i][j] = (X[i][j] - means[j]) / stdDevs[j]
		}
	}

	return normalizedX, means, stdDevs
}
