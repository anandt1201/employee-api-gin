package repository_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"example.com/employee-api/internal/employee"
	"example.com/employee-api/internal/repository"
)

func TestConcurrentCreateAndListIsolation(t *testing.T) {
	r := repository.NewEmployeeList()
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := r.Create(ctx, employee.EmployeeRequest{Name: "A", Email: fmt.Sprintf("a%d@example.com", i), Department: "IT"})
			if err != nil {
				t.Error(err)
			}
			_, _ = r.List(ctx)
		}(i)
	}
	wg.Wait()
	items, err := r.List(ctx)
	if err != nil || len(items) != 50 {
		t.Fatalf("list: len=%d err=%v", len(items), err)
	}
	for i, e := range items {
		if e.ID != int64(i+1) {
			t.Fatal("IDs must be ordered and unique")
		}
	}
	items[0].Name = "modified"
	stored, _ := r.Get(ctx, items[0].ID)
	if stored.Name == "modified" {
		t.Fatal("list exposed internal storage")
	}
}

func TestUniquenessOnUpdate(t *testing.T) {
	r := repository.NewEmployeeList()
	ctx := context.Background()
	a, _ := r.Create(ctx, employee.EmployeeRequest{Email: "a@example.com"})
	b, _ := r.Create(ctx, employee.EmployeeRequest{Email: "b@example.com"})
	_, err := r.Update(ctx, b.ID, employee.EmployeeRequest{Email: "A@example.com"})
	if !errors.Is(err, employee.ErrConflict) {
		t.Fatalf("want conflict, got %v", err)
	}
	unchanged, _ := r.Get(ctx, b.ID)
	if unchanged.Email != "b@example.com" {
		t.Fatal("failed update mutated employee")
	}
	if err := r.Delete(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	_, err = r.Create(ctx, employee.EmployeeRequest{Email: "a@example.com"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCanceledContext(t *testing.T) {
	r := repository.NewEmployeeList()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := r.Create(ctx, employee.EmployeeRequest{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}
