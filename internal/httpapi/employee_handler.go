package httpapi

import (
	"log/slog"
	"strconv"

	"example.com/employee-api/internal/employee"

	"github.com/gin-gonic/gin"
)

type EmployeeHandler struct {
	service *employee.Service
	log     *slog.Logger
}

func NewEmployeeHandler(service *employee.Service, log *slog.Logger) *EmployeeHandler {
	return &EmployeeHandler{service: service, log: log}
}

// list returns all employees.
// @Summary List employees
// @Description Returns employees ordered by ID. Maximum 10 requests per client IP in any rolling 5-second window; blocked requests do not extend the window.
// @Tags employees
// @Produce json
// @Success 200 {object} EmployeeListResponse
// @Failure 429 {object} ErrorResponse
// @Header 429 {integer} Retry-After "Seconds until the next request is allowed, rounded up"
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/employees [get]
func (h *EmployeeHandler) list(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		h.respondError(c, err)
		return
	}
	responses := make([]EmployeeResponse, len(items))
	for i, item := range items {
		responses[i] = newEmployeeResponse(item)
	}
	c.JSON(200, EmployeeListResponse{Data: responses})
}

// get returns an employee by ID.
// @Summary Get employee
// @Tags employees
// @Produce json
// @Param id path int true "Positive employee ID" minimum(1)
// @Success 200 {object} EmployeeDataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/employees/{id} [get]
func (h *EmployeeHandler) get(c *gin.Context) {
	n, ok := parseID(c)
	if !ok {
		return
	}
	e, err := h.service.Get(c.Request.Context(), n)
	if err != nil {
		h.respondError(c, err)
		return
	}
	c.JSON(200, EmployeeDataResponse{Data: newEmployeeResponse(e)})
}

// create creates an employee.
// @Summary Create employee
// @Description Requires name, email, and department. Email is normalized and unique. Request body limit is 1 MiB.
// @Tags employees
// @Accept json
// @Produce json
// @Param employee body employee.EmployeeRequest true "Employee to create"
// @Success 201 {object} EmployeeDataResponse
// @Header 201 {string} Location "Relative URL of the created employee"
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 413 {object} ErrorResponse
// @Failure 415 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/employees [post]
func (h *EmployeeHandler) create(c *gin.Context) {
	in, ok := decodeJSON[employee.EmployeeRequest](c)
	if !ok {
		return
	}
	e, err := h.service.Create(c.Request.Context(), in)
	if err != nil {
		h.respondError(c, err)
		return
	}
	c.Header("Location", "/api/v1/employees/"+strconv.FormatInt(e.ID, 10))
	c.JSON(201, EmployeeDataResponse{Data: newEmployeeResponse(e)})
}

// update replaces an employee.
// @Summary Replace employee
// @Description Full replacement: name, email, and department are required. Request body limit is 1 MiB.
// @Tags employees
// @Accept json
// @Produce json
// @Param id path int true "Positive employee ID" minimum(1)
// @Param employee body employee.EmployeeRequest true "Replacement employee"
// @Success 200 {object} EmployeeDataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 413 {object} ErrorResponse
// @Failure 415 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/employees/{id} [put]
func (h *EmployeeHandler) update(c *gin.Context) {
	n, ok := parseID(c)
	if !ok {
		return
	}
	in, ok := decodeJSON[employee.EmployeeRequest](c)
	if !ok {
		return
	}
	e, err := h.service.Update(c.Request.Context(), n, in)
	if err != nil {
		h.respondError(c, err)
		return
	}
	c.JSON(200, EmployeeDataResponse{Data: newEmployeeResponse(e)})
}

// delete removes an employee.
// @Summary Delete employee
// @Tags employees
// @Produce json
// @Param id path int true "Positive employee ID" minimum(1)
// @Success 204 "Employee deleted; no response body"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/employees/{id} [delete]
func (h *EmployeeHandler) delete(c *gin.Context) {
	n, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), n); err != nil {
		h.respondError(c, err)
		return
	}
	c.Status(204)
}
