package main

import "fmt"

func main() {
	// set up csv file path
	csvFilePath := "loan_applicants.csv"

	// load applicant data from csv
	_, err := LoadApplicantsFromCSV(csvFilePath)
	if err != nil {
		fmt.Printf("error loading applicants: %v\n", err)
		return
	}
}
