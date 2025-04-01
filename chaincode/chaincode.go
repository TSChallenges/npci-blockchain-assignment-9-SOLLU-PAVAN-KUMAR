package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type LoanContract struct {
}

type Loan struct {
	LoanID           string    `json:"loanId"`
	BorrowerID       string    `json:"borrowerId"`
	LenderID         string    `json:"lenderId"`
	Amount           float64   `json:"amount"`
	InterestRate     float64   `json:"interestRate"`
	Duration         int       `json:"duration"`
	Status           string    `json:"status"`
	DisbursementDate string    `json:"disbursementDate"`
	RepaymentDue     float64   `json:"repaymentDue"`
	RemainingBalance float64   `json:"remainingBalance"`
	Collateral       string    `json:"collateral"`
	Defaulted        bool      `json:"defaulted"`
	AuditHistory     []string  `json:"auditHistory"`
}

func (t *LoanContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (t *LoanContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "RequestLoan":
		return t.RequestLoan(stub, args)
	case "ApproveLoan":
		return t.ApproveLoan(stub, args)
	case "DisburseLoan":
		return t.DisburseLoan(stub, args)
	case "RepayLoan":
		return t.RepayLoan(stub, args)
	case "CheckLoanStatus":
		return t.CheckLoanStatus(stub, args)
	case "MarkAsDefaulted":
		return t.MarkAsDefaulted(stub, args)
	case "AddCollateral":
		return t.AddCollateral(stub, args)
	case "GetLoanHistory":
		return t.GetLoanHistory(stub, args)
	default:
		return shim.Error("Invalid function name")
	}
}

// RequestLoan allows a borrower to request a loan.
func (t *LoanContract) RequestLoan(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expected 4")
	}

	loanID := args[0]
	borrowerID := args[1]
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return shim.Error("Invalid loan amount")
	}
	duration, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Invalid loan duration")
	}

	loan := Loan{
		LoanID:       loanID,
		BorrowerID:   borrowerID,
		Amount:       amount,
		Duration:     duration,
		Status:       "Pending",
		RemainingBalance: amount,
	}

	loanBytes, err := json.Marshal(loan)
	if err != nil {
		return shim.Error("Error marshalling loan object")
	}

	err = stub.PutState(loanID, loanBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to request loan: %s", err.Error()))
	}

	return shim.Success(nil)
}

// ApproveLoan allows a lender to approve a loan request.
func (t *LoanContract) ApproveLoan(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expected 2")
	}

	loanID := args[0]
	lenderID := args[1]

	loanBytes, err := stub.GetState(loanID)
	if err != nil || loanBytes == nil {
		return shim.Error("Loan not found")
	}

	var loan Loan
	err = json.Unmarshal(loanBytes, &loan)
	if err != nil {
		return shim.Error("Error unmarshalling loan object")
	}

	if loan.Status != "Pending" {
		return shim.Error("Loan is not in Pending status")
	}

	loan.LenderID = lenderID
	loan.Status = "Approved"
	loan.AuditHistory = append(loan.AuditHistory, "Loan Approved")

	loanBytes, err = json.Marshal(loan)
	if err != nil {
		return shim.Error("Error marshalling loan object")
	}

	err = stub.PutState(loanID, loanBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to approve loan: %s", err.Error()))
	}

	return shim.Success(nil)
}

// DisburseLoan allows the lender to disburse the loan amount to the borrower.
func (t *LoanContract) DisburseLoan(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expected 2")
	}

	loanID := args[0]
	disbursementDate := args[1]

	loanBytes, err := stub.GetState(loanID)
	if err != nil || loanBytes == nil {
		return shim.Error("Loan not found")
	}

	var loan Loan
	err = json.Unmarshal(loanBytes, &loan)
	if err != nil {
		return shim.Error("Error unmarshalling loan object")
	}

	if loan.Status != "Approved" {
		return shim.Error("Loan is not approved")
	}

	loan.Status = "Active"
	loan.DisbursementDate = disbursementDate
	loan.AuditHistory = append(loan.AuditHistory, "Loan Disbursed")

	loanBytes, err = json.Marshal(loan)
	if err != nil {
		return shim.Error("Error marshalling loan object")
	}

	err = stub.PutState(loanID, loanBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to disburse loan: %s", err.Error()))
	}

	return shim.Success(nil)
}

// RepayLoan allows the borrower to make a repayment.
func (t *LoanContract) RepayLoan(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expected 2")
	}

	loanID := args[0]
	repaymentAmount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("Invalid repayment amount")
	}

	loanBytes, err := stub.GetState(loanID)
	if err != nil || loanBytes == nil {
		return shim.Error("Loan not found")
	}

	var loan Loan
	err = json.Unmarshal(loanBytes, &loan)
	if err != nil {
		return shim.Error("Error unmarshalling loan object")
	}

	if loan.Status != "Active" {
		return shim.Error("Loan is not active")
	}

	loan.RemainingBalance -= repaymentAmount
	if loan.RemainingBalance <= 0 {
		loan.Status = "Repaid"
	}

	loan.AuditHistory = append(loan.AuditHistory, fmt.Sprintf("Repayment of %f made", repaymentAmount))

	loanBytes, err = json.Marshal(loan)
	if err != nil {
		return shim.Error("Error marshalling loan object")
	}

	err = stub.PutState(loanID, loanBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to repay loan: %s", err.Error()))
	}

	return shim.Success(nil)
}

// CheckLoanStatus allows checking the loan's status.
func (t *LoanContract) CheckLoanStatus(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expected 1")
	}

	loanID := args[0]

	loanBytes, err := stub.GetState(loanID)
	if err != nil || loanBytes == nil {
		return shim.Error("Loan not found")
	}

	return shim.Success(loanBytes)
}

// MarkAsDefaulted marks the loan as defaulted if repayment is not made.
func (t *LoanContract) MarkAsDefaulted(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expected 1")
	}

	loanID := args[0]

	loanBytes, err := stub.GetState(loanID)
	if err != nil || loanBytes == nil {
		return shim.Error("Loan not found")
	}

	var loan Loan
	err = json.Unmarshal(loanBytes, &loan)
	if err != nil {
		return shim.Error("Error unmarshalling loan object")
	}

	if loan.Status != "Active" {
		return shim.Error("Loan is not active")
	}

	// Check for default condition, for example, if repayment due is overdue
	// Here, we assume that some logic is added to detect overdue loans

	loan.Defaulted = true
	loan.Status = "Defaulted"
	loan.AuditHistory = append(loan.AuditHistory, "Loan Defaulted")

	loanBytes, err = json.Marshal(loan)
	if err != nil {
		return shim.Error("Error marshalling loan object")
	}

	err = stub.PutState(loanID, loanBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to mark loan as defaulted: %s", err.Error()))
	}

	return shim.Success(nil)
}

// AddCollateral allows the borrower to add collateral for a secured loan.
func (t *LoanContract) AddCollateral(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expected 2")
	}

	loanID := args[0]
	collateral := args[1]

	loanBytes, err := stub.GetState(loanID)
	if err != nil || loanBytes == nil {
		return shim.Error("Loan not found")
	}

	var loan Loan
	err = json.Unmarshal(loanBytes, &loan)
	if err != nil {
		return shim.Error("Error unmarshalling loan object")
	}

	if loan.Status != "Pending" && loan.Status != "Approved" {
		return shim.Error("Loan cannot accept collateral in the current state")
	}

	loan.Collateral = collateral
	loan.AuditHistory = append(loan.AuditHistory, fmt.Sprintf("Collateral added: %s", collateral))

	loanBytes, err = json.Marshal(loan)
	if err != nil {
		return shim.Error("Error marshalling loan object")
	}

	err = stub.PutState(loanID, loanBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to add collateral: %s", err.Error()))
	}

	return shim.Success(nil)
}

// GetLoanHistory allows regulators to retrieve the loan's history.
func (t *LoanContract) GetLoanHistory(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expected 1")
	}

	loanID := args[0]

	loanBytes, err := stub.GetState(loanID)
	if err != nil || loanBytes == nil {
		return shim.Error("Loan not found")
	}

	return shim.Success(loanBytes)
}

func main() {
	err := shim.Start(new(LoanContract))
	if err != nil {
		fmt.Printf("Error starting Loan contract: %s", err)
	}
}
