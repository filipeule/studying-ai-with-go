package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func LoadApplicantsFromCSV(filepath string) ([]Applicant, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// read the header row
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading header: %w", err)
	}

	columnIndices := map[string]int{
		"income":         -1,
		"creditScore":    -1,
		"loanAmount":     -1,
		"debtToIncome":   -1,
		"yearsEmployed":  -1,
		"protectedClass": -1,
	}

	// find the column indices
	for i, column := range header {
		col := strings.ToLower(strings.TrimSpace(column))

		switch {
		case strings.Contains(col, "income") && !strings.Contains(col, "debt"):
			columnIndices["income"] = i
		case strings.Contains(col, "credit"):
			columnIndices["creditScore"] = i
		case strings.Contains(col, "loan"):
			columnIndices["loanAmount"] = i
		case strings.Contains(col, "debt"):
			columnIndices["debtToIncome"] = i
		case strings.Contains(col, "employ"):
			columnIndices["yearsEmployed"] = i
		case strings.Contains(col, "protect"):
			columnIndices["protectedClass"] = i
		}
	}

	// verify that all columns are found
	for field, idx := range columnIndices {
		if idx == -1 {
			return nil, fmt.Errorf("required column %s not found in csv", field)
		}
	}

	// read applicant records
	var applicants []Applicant
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading csv: %w", err)
	}

	for i, record := range records {
		// parse values
		income, err := parseFloat(record[columnIndices["income"]])
		if err != nil {
			return nil, fmt.Errorf("invalid income at line %d: %w", i+2, err)
		}
		// convert income to thousands
		income = income / 1000

		creditScore, err := parseFloat(record[columnIndices["creditScore"]])
		if err != nil {
			return nil, fmt.Errorf("invalid credit score at line %d: %w", i+2, err)
		}
		// normalize credit score if in 300 - 850 range
		if creditScore > 1 {
			creditScore = (creditScore - 300) / 550
		}

		loanAmount, err := parseFloat(record[columnIndices["loanAmount"]])
		if err != nil {
			return nil, fmt.Errorf("invalid loan amount at line %d: %w", i+2, err)
		}
		// convert loan amount to thousands
		loanAmount = loanAmount / 1000

		debtRatio, err := parseFloat(record[columnIndices["debtToIncome"]])
		if err != nil {
			return nil, fmt.Errorf("invalid debt ratio at line %d: %w", i+2, err)
		}
		// normalize debt ratio if its given as a percentage
		if debtRatio > 1 {
			debtRatio = debtRatio / 100
		}

		yearsEmployed, err := parseFloat(record[columnIndices["yearsEmployed"]])
		if err != nil {
			return nil, fmt.Errorf("invalid years employed at line %d: %w", i+2, err)
		}

		protectedClass, err := parseBool(record[columnIndices["protectedClass"]])
		if err != nil {
			return nil, fmt.Errorf("invalid protected class at line %d: %w", i+2, err)
		}

		applicants = append(applicants, Applicant{
			income:         income,
			creditScore:    creditScore,
			loanAmount:     loanAmount,
			debtToIncome:   debtRatio,
			yearsEmployed:  yearsEmployed,
			protectedClass: protectedClass,
		})
	}

	fmt.Printf("succesfully loaded %d applicants from csv\n", len(applicants))
	return applicants, nil
}
