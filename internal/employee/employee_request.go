package employee

// EmployeeRequest is deliberately separate from Employee: clients cannot choose an ID.
type EmployeeRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Department string `json:"department"`
}
