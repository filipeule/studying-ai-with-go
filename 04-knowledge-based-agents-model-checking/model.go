package main

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

// Some means of determining fairness

// Check verifies if the ai model satisfies the fairness property

// Evaluate risk

// loading the csv file

// VerifyModel checks if the model satisfies some property
