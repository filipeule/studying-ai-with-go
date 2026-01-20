package main

import "fmt"

type RiskProperty struct {
	maxHighRiskApprovalRate float64
}

func (p *RiskProperty) Name() string {
	return "Risk Property"
}

func (p *RiskProperty) Check(model *LoanApprovalAI, applicants []Applicant) (bool, []Applicant) {
	var riskyApprovals []Applicant
	var highRiskApproved, highRiskTotal int

	for _, applicant := range applicants {
		isHighRisk := applicant.creditScore < 0.5 && applicant.debtToIncome > 0.5

		if isHighRisk {
			highRiskTotal++
			if model.ApproveLoan(applicant) {
				highRiskApproved++
				riskyApprovals = append(riskyApprovals, applicant)
			}
		}
	}

	if highRiskTotal == 0 {
		return true, nil
	}

	// calculate the approval rate for high risk applicants
	highRiskApprovalRate := float64(highRiskApproved) / float64(highRiskTotal)
	fmt.Printf(
		"high-risk approval rate: %.2f%% (maximum allowed: %.2f%%)\n", highRiskApprovalRate*100, p.maxHighRiskApprovalRate*100,
	)

	return highRiskApprovalRate <= p.maxHighRiskApprovalRate, riskyApprovals
}
