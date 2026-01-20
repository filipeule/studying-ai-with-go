package main

import "fmt"

func VerifyModel(model *LoanApprovalAI, property Property, applicants []Applicant) {
	propertyName := property.Name()
	fmt.Printf("Verifying %s...\n", propertyName)

	// run the property check on the model and the applicants
	satisfied, counterExamples := property.Check(model, applicants)

	// print out verification result
	if satisfied {
		fmt.Printf("✅ %s is satisfied\n", propertyName)
	} else {
		fmt.Printf("❌ %s is violated. found %d problematic cases\n", propertyName, len(counterExamples))

		// print up to three examples for clarity
		for i := range min(3, len(counterExamples)) {
			a := counterExamples[i]
			fmt.Printf("   Example: %d: Income: $%.1fk, Credit Score: %.2f, Debt Ratio: %.2f, Protected: %v, Decision: %v\n", i+1, a.income, a.creditScore, a.debtToIncome, a.protectedClass, model.ApproveLoan(a))
		}

		// if there are more than 3 examples, just indicate how many more
		if len(counterExamples) > 3 {
			fmt.Printf("   ... and %d more\n", len(counterExamples) - 3)
		}
	}
}