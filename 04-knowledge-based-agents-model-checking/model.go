package main

import "fmt"

type LoanApprovalAI struct {
	// weights for different factors used in the decision making process
	incomeWeight      float64
	creditScoreWeight float64
	loanAmountWeight  float64
	debtRatioWeight   float64
	employmentWeight  float64
	approvalThresold  float64
}

type Applicant struct {
	income         float64 // annual income in thousands, so 50 = $50,000.00
	creditScore    float64 // credit score normalized to 0 to 1, from a typical 300-850
	loanAmount     float64 // loan amount in thousands
	debtToIncome   float64 // debt to income ratio (0-1), already normalized
	yearsEmployed  float64
	protectedClass bool // whether or not the applicant belongs to some protected class
}

// ApproveLoan determines if the applicant should be approved for a loan
func (ai *LoanApprovalAI) ApproveLoan(applicant Applicant) bool {
	loanToIncomeRatio := 0.0
	if applicant.income > 0 {
		loanToIncomeRatio = applicant.loanAmount / applicant.income
	}

	score := applicant.income * ai.incomeWeight +
		applicant.creditScore * ai.creditScoreWeight -
		loanToIncomeRatio * ai.loanAmountWeight -
		applicant.debtToIncome * ai.debtRatioWeight +
		applicant.yearsEmployed * ai.employmentWeight

	return score > ai.approvalThresold
}

func PrintModelParams(model *LoanApprovalAI, description string) {
	fmt.Printf("\n===== %s =====\n", description)
	fmt.Printf("- Income Weight: %.2f\n", model.incomeWeight)
	fmt.Printf("- Credit Score Weight: %.2f\n", model.creditScoreWeight)
	fmt.Printf("- Loan Amount Weight: %.2f\n", model.loanAmountWeight)
	fmt.Printf("- Debt Ratio Weight: %.2f\n", model.debtRatioWeight)
	fmt.Printf("- Employment Weight: %.2f\n", model.employmentWeight)
	fmt.Printf("- Approval Threshold: %.2f\n", model.approvalThresold)
}
