package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Loan struct represents a loan request
type Loan struct {
	LoanID           string  `json:"loanId"`
	BorrowerID       string  `json:"borrowerId"`
	LenderID         string  `json:"lenderId"`
	Amount           float64 `json:"amount"`
	InterestRate     float64 `json:"interestRate"`
	Duration         int     `json:"duration"`
	Status           string  `json:"status"` // Pending, Approved, Active, Repaid, Defaulted
	DisbursementDate string  `json:"disbursementDate"`
	RepaymentDue     float64 `json:"repaymentDue"`
	RemainingBalance float64 `json:"remainingBalance"`
	Defaulted        bool    `json:"defaulted"`
}

// SmartContract provides functions for managing loans
type SmartContract struct {
	contractapi.Contract
}

// RequestLoan allows a borrower to request a loan
func (s *SmartContract) RequestLoan(ctx contractapi.TransactionContextInterface, loanID, borrowerID string, amount float64, interestRate float64, duration int) error {
	// Check if the loan ID already exists
	loanAsBytes, err := ctx.GetStub().GetState(loanID)
	if err != nil {
		return fmt.Errorf("failed to read loan state: %v", err)
	}
	if loanAsBytes != nil {
		return fmt.Errorf("the loan with ID %s already exists", loanID)
	}

	// Create a new loan request
	loan := Loan{
		LoanID:       loanID,
		BorrowerID:   borrowerID,
		Amount:       amount,
		InterestRate: interestRate,
		Duration:     duration,
		Status:       "Pending", // Loan is initially in Pending state
		RemainingBalance: amount, // Initially, the entire loan amount is pending repayment
	}

	// Save the loan to the world state
	loanAsBytes, err = json.Marshal(loan)
	if err != nil {
		return fmt.Errorf("failed to marshal loan: %v", err)
	}

	err = ctx.GetStub().PutState(loanID, loanAsBytes)
	if err != nil {
		return fmt.Errorf("failed to create loan: %v", err)
	}

	return nil
}

// ApproveLoan allows a lender to approve a loan request
func (s *SmartContract) ApproveLoan(ctx contractapi.TransactionContextInterface, loanID, lenderID string) error {
	// Retrieve the loan from the world state
	loanAsBytes, err := ctx.GetStub().GetState(loanID)
	if err != nil {
		return fmt.Errorf("failed to read loan state: %v", err)
	}
	if loanAsBytes == nil {
		return fmt.Errorf("loan with ID %s not found", loanID)
	}

	var loan Loan
	err = json.Unmarshal(loanAsBytes, &loan)
	if err != nil {
		return fmt.Errorf("failed to unmarshal loan: %v", err)
	}

	// Check if the loan is already approved
	if loan.Status != "Pending" {
		return fmt.Errorf("loan with ID %s is not in Pending status", loanID)
	}

	// Approve the loan
	loan.LenderID = lenderID
	loan.Status = "Approved"

	// Save the updated loan back to the world state
	loanAsBytes, err = json.Marshal(loan)
	if err != nil {
		return fmt.Errorf("failed to marshal loan: %v", err)
	}

	err = ctx.GetStub().PutState(loanID, loanAsBytes)
	if err != nil {
		return fmt.Errorf("failed to approve loan: %v", err)
	}

	return nil
}

// RepayLoan allows a borrower to repay a loan
func (s *SmartContract) RepayLoan(ctx contractapi.TransactionContextInterface, loanID string, amount float64) error {
	// Retrieve the loan from the world state
	loanAsBytes, err := ctx.GetStub().GetState(loanID)
	if err != nil {
		return fmt.Errorf("failed to read loan state: %v", err)
	}
	if loanAsBytes == nil {
		return fmt.Errorf("loan with ID %s not found", loanID)
	}

	var loan Loan
	err = json.Unmarshal(loanAsBytes, &loan)
	if err != nil {
		return fmt.Errorf("failed to unmarshal loan: %v", err)
	}

	// Check if the loan is in an Active state
	if loan.Status != "Active" {
		return fmt.Errorf("loan with ID %s is not in Active status", loanID)
	}

	// Update the remaining balance
	loan.RemainingBalance -= amount
	if loan.RemainingBalance <= 0 {
		loan.Status = "Repaid"
		loan.RemainingBalance = 0
	}

	// Save the updated loan back to the world state
	loanAsBytes, err = json.Marshal(loan)
	if err != nil {
		return fmt.Errorf("failed to marshal loan: %v", err)
	}

	err = ctx.GetStub().PutState(loanID, loanAsBytes)
	if err != nil {
		return fmt.Errorf("failed to repay loan: %v", err)
	}

	return nil
}

// QueryLoan queries the details of a loan by ID
func (s *SmartContract) QueryLoan(ctx contractapi.TransactionContextInterface, loanID string) (*Loan, error) {
	// Retrieve the loan from the world state
	loanAsBytes, err := ctx.GetStub().GetState(loanID)
	if err != nil {
		return nil, fmt.Errorf("failed to read loan state: %v", err)
	}
	if loanAsBytes == nil {
		return nil, fmt.Errorf("loan with ID %s not found", loanID)
	}

	var loan Loan
	err = json.Unmarshal(loanAsBytes, &loan)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal loan: %v", err)
	}

	return &loan, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating chaincode: %v", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v", err)
	}
}
