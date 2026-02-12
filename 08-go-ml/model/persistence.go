package model

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type ModelMetadata struct {
	SavedAt     time.Time `json:"save_at"`
	Description string    `json:"description,omitempty"`
	NumSamples  int       `json:"num_samples,omitempty"`
	Version     string    `json:"version,omitempty"`
}

func SaveModelToJSON(model *LinearRegression, filePath string, description string, numSamples int) error {
	// add metadata field directly to the model for saving
	model.SavedAt = time.Now()
	model.Description = description
	model.NumSamples = numSamples
	model.Version = "1.0"

	// marshall the model directly to json
	modelJSON, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling model to json: %w", err)
	}

	// write json to file
	err = os.WriteFile(filePath, modelJSON, 0644)
	if err != nil {
		return fmt.Errorf("error savel model to json file: %w", err)
	}

	fmt.Printf("Model saved to: %s\n", filePath)
	fmt.Printf("  - Target %s\n", model.Target)
	fmt.Printf("  - Features %s\n", model.Features)
	fmt.Printf("  - R-Squared %.4f\n", model.RSquared)

	return nil
}

func LoadModelFromJSON(filePath string) (*LinearRegression, error) {
	// read json file
	modelJSON, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading model file: %w", err)
	}

	// unmarshal
	var loadedModel LinearRegression
	err = json.Unmarshal(modelJSON, &loadedModel)
	if err != nil {
		return nil, fmt.Errorf("invalid model format in file: %w", err)
	}

	// print some model information
	fmt.Printf("Model successfully loaded from %s\n", filePath)

	if loadedModel.Version != "" {
		fmt.Printf("Model version: %s\n", loadedModel.Version)
	}

	if loadedModel.Description != "" {
		fmt.Printf("Model description: %s\n", loadedModel.Description)
	}

	fmt.Printf("Loaded model information:\n")
	fmt.Printf("- Target: %s\n", loadedModel.Target)
	fmt.Printf("- Features: %s\n", loadedModel.Features)
	fmt.Printf("- Intercept: %.4f\n", loadedModel.Intercept)
	fmt.Printf("- Coefficients: %v\n", loadedModel.Coefficients)
	fmt.Printf("- Coefficients: %.4f\n", loadedModel.RSquared)

	return &loadedModel, nil
}
