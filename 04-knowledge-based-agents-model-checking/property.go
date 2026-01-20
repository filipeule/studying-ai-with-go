package main

import (
	"fmt"
)

type Property interface {
	Check(model *LoanApprovalAI, applicants []Applicant) (bool, []Applicant)
	Name() string
}

type FairnessProperty struct {
	maxDisparity float64
}

func (p *FairnessProperty) Name() string {
	return "Fairness Property"
}

func (p *FairnessProperty) Check(model *LoanApprovalAI, applicants []Applicant) (bool, []Applicant) {
	var unfairDecisions []Applicant
	var protectedApproved, protectedTotal, nonProtectedApproved, nonProtectedTotal int

	// loop through all applicants and count approvals for each group
	for _, applicant := range applicants {
		// make a decision
		decision := model.ApproveLoan(applicant)

		// update counters based on protected status class
		if applicant.protectedClass {
			protectedTotal++
			if decision {
				protectedApproved++
			}
		} else {
			nonProtectedTotal++
			if decision {
				nonProtectedApproved++
			}
		}

		// check for potentially problematic individual decisions
		if !decision && applicant.creditScore > 0.7 && applicant.debtToIncome < 0.3 && applicant.income > 60 {
			unfairDecisions = append(unfairDecisions, applicant)
		}
	}

	// calculate approval rates for each group
	protectedRate := float64(protectedApproved) / float64(protectedTotal)
	nonProtectedRate := float64(nonProtectedApproved) / float64(nonProtectedTotal)

	disparity := nonProtectedRate - protectedRate // how much more likey non protected applicants are to be approved

	// print approval rates and disparity
	fmt.Printf(
		"approval rate - protected: %.2f%%, non-protected: %.2f%%\n", protectedRate*100, nonProtectedRate*100,
	)
	fmt.Printf("disparity: %.2f%% (maximum allowd: %.2f%%)\n", disparity*100, p.maxDisparity*100)

	return disparity <= p.maxDisparity && len(unfairDecisions) == 0, unfairDecisions
}
