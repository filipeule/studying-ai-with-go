package main

import "fmt"

func main() {
	// set up csv file path
	csvFilePath := "loan_applicants.csv"

	// load applicant data from csv
	applicants, err := LoadApplicantsFromCSV(csvFilePath)
	if err != nil {
		fmt.Printf("error loading applicants: %v\n", err)
		return
	}

	// define some properties we want to check (fairness and risk)
	fairnessProperty := &FairnessProperty{
		maxDisparity: 0.05, // at most 5% difference in approval rates
	}
	riskProperty := &RiskProperty{
		maxHighRiskApprovalRate: 0.1, // at most 10% of high-risk applicants to be approved
	}

	// create some test models
	models := []*LoanApprovalAI{
		{
			incomeWeight:      0.3,
			creditScoreWeight: 0.4,
			loanAmountWeight:  1.0,
			debtRatioWeight:   2.0,
			approvalThresold:  5.0,
			employmentWeight:  0.1,
		},
		{
			incomeWeight:      0.25,
			creditScoreWeight: 0.45,
			loanAmountWeight:  1.2,
			debtRatioWeight:   2.5,
			approvalThresold:  4.5,
			employmentWeight:  0.5,
		},
		{
			incomeWeight:      0.2,
			creditScoreWeight: 0.5,
			loanAmountWeight:  1.5,
			debtRatioWeight:   3.0,
			approvalThresold:  4.0,
			employmentWeight:  0.2,
		},
	}

	// test each model configuration against both properties
	descriptions := []string{
		"Loan Approval AI Model with Initial Parameters",
		"Loan Approval AI Model with Adjusted Parameters",
		"Loan Approval AI Model with Final Parameters",
	}

	for i, model := range models {
		// print the current model's parameters
		PrintModelParams(model, descriptions[i])

		// verify the model against both properties
		VerifyModel(model, fairnessProperty, applicants)
		VerifyModel(model, riskProperty, applicants)
	}
}
