package httpapi

import "example.com/employee-api/internal/employee"

// EmployeeResponse is the public employee representation returned by the API.
// Mapping explicitly keeps the HTTP contract independent of the storage model.
type EmployeeResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Department string `json:"department"`
}

func newEmployeeResponse(e employee.Employee) EmployeeResponse {
	return EmployeeResponse{ID: e.ID, Name: e.Name, Email: e.Email, Department: e.Department}
}
