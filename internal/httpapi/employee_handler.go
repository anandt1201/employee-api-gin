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
	c.JSON(200, gin.H{"data": responses})
}

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
	c.JSON(200, gin.H{"data": newEmployeeResponse(e)})
}

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
	c.JSON(201, gin.H{"data": newEmployeeResponse(e)})
}

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
	c.JSON(200, gin.H{"data": newEmployeeResponse(e)})
}

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
