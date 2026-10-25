// Package repository contains storage adapters implementing domain repository contracts.
package repository

import (
	"context"
	"strings"
	"sync"

	"example.com/employee-api/internal/employee"
)

type EmployeeList struct {
	mu        sync.RWMutex
	employees []employee.Employee
	nextID    int64
}

var _ employee.Repository = (*EmployeeList)(nil)

func NewEmployeeList() *EmployeeList {
	return &EmployeeList{employees: make([]employee.Employee, 0), nextID: 1}
}

func (r *EmployeeList) List(ctx context.Context) ([]employee.Employee, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]employee.Employee, len(r.employees))
	copy(result, r.employees)
	return result, nil
}

func (r *EmployeeList) Get(ctx context.Context, id int64) (employee.Employee, error) {
	if err := ctx.Err(); err != nil {
		return employee.Employee{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.employees {
		if e.ID == id {
			return e, nil
		}
	}
	return employee.Employee{}, employee.ErrNotFound
}

func (r *EmployeeList) Create(ctx context.Context, in employee.EmployeeRequest) (employee.Employee, error) {
	if err := ctx.Err(); err != nil {
		return employee.Employee{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.emailExists(in.Email, 0) {
		return employee.Employee{}, employee.ErrConflict
	}
	e := employee.Employee{ID: r.nextID, Name: in.Name, Email: in.Email, Department: in.Department}
	r.nextID++
	r.employees = append(r.employees, e)
	return e, nil
}

func (r *EmployeeList) Update(ctx context.Context, id int64, in employee.EmployeeRequest) (employee.Employee, error) {
	if err := ctx.Err(); err != nil {
		return employee.Employee{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, e := range r.employees {
		if e.ID == id {
			if r.emailExists(in.Email, id) {
				return employee.Employee{}, employee.ErrConflict
			}
			e = employee.Employee{ID: id, Name: in.Name, Email: in.Email, Department: in.Department}
			r.employees[i] = e
			return e, nil
		}
	}
	return employee.Employee{}, employee.ErrNotFound
}

func (r *EmployeeList) Delete(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, e := range r.employees {
		if e.ID == id {
			r.employees = append(r.employees[:i], r.employees[i+1:]...)
			return nil
		}
	}
	return employee.ErrNotFound
}

// Caller holds the lock so checking uniqueness and writing form one operation.
func (r *EmployeeList) emailExists(email string, except int64) bool {
	for _, e := range r.employees {
		if e.ID != except && strings.EqualFold(e.Email, email) {
			return true
		}
	}
	return false
}
