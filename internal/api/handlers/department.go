package handlers

import (
	"net/http"
	"strconv"
	"gm58-hr-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DepartmentHandler struct {
	db *gorm.DB
}

func NewDepartmentHandler(db *gorm.DB) *DepartmentHandler {
	return &DepartmentHandler{
		db: db,
	}
}

func (dh *DepartmentHandler) GetDepartments(c *gin.Context) {
	var departments []models.Department
	
	query := dh.db.Preload("Manager")
	
	// Filter by active status
	if isActive := c.Query("is_active"); isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Department{}).Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&departments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch departments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"departments": departments,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

func (dh *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var department models.Department
	if err := c.ShouldBindJSON(&department); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate manager exists if provided
	if department.ManagerID != nil {
		var manager models.Employee
		if err := dh.db.First(&manager, *department.ManagerID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manager"})
			return
		}
	}

	if err := dh.db.Create(&department).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create department"})
		return
	}

	// Reload with relationships
	dh.db.Preload("Manager").First(&department, department.ID)

	c.JSON(http.StatusCreated, department)
}

func (dh *DepartmentHandler) GetDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department ID"})
		return
	}

	var department models.Department
	if err := dh.db.Preload("Manager").First(&department, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Department not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch department"})
		return
	}

	c.JSON(http.StatusOK, department)
}

func (dh *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department ID"})
		return
	}

	var department models.Department
	if err := dh.db.First(&department, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Department not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch department"})
		return
	}

	var updateData models.Department
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate manager if changed
	if updateData.ManagerID != nil && (department.ManagerID == nil || *updateData.ManagerID != *department.ManagerID) {
		var manager models.Employee
		if err := dh.db.First(&manager, *updateData.ManagerID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manager"})
			return
		}
	}

	if err := dh.db.Model(&department).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update department"})
		return
	}

	// Reload with relationships
	dh.db.Preload("Manager").First(&department, department.ID)

	c.JSON(http.StatusOK, department)
}

func (dh *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department ID"})
		return
	}

	var department models.Department
	if err := dh.db.First(&department, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Department not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch department"})
		return
	}

	// Check if department has any employees
	var employeeCount int64
	dh.db.Model(&models.Employee{}).Where("department_id = ?", department.ID).Count(&employeeCount)
	if employeeCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete department as it has employees"})
		return
	}

	// Check if department has any positions
	var positionCount int64
	dh.db.Model(&models.Position{}).Where("department_id = ?", department.ID).Count(&positionCount)
	if positionCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete department as it has positions"})
		return
	}

	if err := dh.db.Delete(&department).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete department"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Department deleted successfully"})
}
