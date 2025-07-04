package handlers

import (
	"net/http"
	"strconv"
	"gm58-hr-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PositionHandler struct {
	db *gorm.DB
}

func NewPositionHandler(db *gorm.DB) *PositionHandler {
	return &PositionHandler{
		db: db,
	}
}

func (ph *PositionHandler) GetPositions(c *gin.Context) {
	var positions []models.Position
	
	query := ph.db.Preload("Department").Preload("Currency")
	
	// Filter by department if provided
	if department := c.Query("department"); department != "" {
		query = query.Joins("JOIN departments ON positions.department_id = departments.id").
			Where("departments.name = ?", department)
	}
	
	// Filter by active status
	if isActive := c.Query("is_active"); isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("positions.is_active = ?", active)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Position{}).Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&positions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch positions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"positions": positions,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

func (ph *PositionHandler) CreatePosition(c *gin.Context) {
	var position models.Position
	if err := c.ShouldBindJSON(&position); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate department exists
	var department models.Department
	if err := ph.db.First(&department, position.DepartmentID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department"})
		return
	}

	// Validate currency exists
	var currency models.Currency
	if err := ph.db.First(&currency, position.CurrencyID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid currency"})
		return
	}

	// Validate salary range
	if position.MinSalary > position.MaxSalary {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Minimum salary cannot be greater than maximum salary"})
		return
	}

	if err := ph.db.Create(&position).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create position"})
		return
	}

	// Reload with relationships
	ph.db.Preload("Department").Preload("Currency").First(&position, position.ID)

	c.JSON(http.StatusCreated, position)
}

func (ph *PositionHandler) GetPosition(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid position ID"})
		return
	}

	var position models.Position
	if err := ph.db.Preload("Department").Preload("Currency").
		First(&position, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch position"})
		return
	}

	c.JSON(http.StatusOK, position)
}

func (ph *PositionHandler) UpdatePosition(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid position ID"})
		return
	}

	var position models.Position
	if err := ph.db.First(&position, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch position"})
		return
	}

	var updateData models.Position
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate department if changed
	if updateData.DepartmentID != 0 && updateData.DepartmentID != position.DepartmentID {
		var department models.Department
		if err := ph.db.First(&department, updateData.DepartmentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department"})
			return
		}
	}

	// Validate currency if changed
	if updateData.CurrencyID != 0 && updateData.CurrencyID != position.CurrencyID {
		var currency models.Currency
		if err := ph.db.First(&currency, updateData.CurrencyID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid currency"})
			return
		}
	}

	// Validate salary range
	minSalary := updateData.MinSalary
	maxSalary := updateData.MaxSalary
	if minSalary == 0 {
		minSalary = position.MinSalary
	}
	if maxSalary == 0 {
		maxSalary = position.MaxSalary
	}
	if minSalary > maxSalary {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Minimum salary cannot be greater than maximum salary"})
		return
	}

	if err := ph.db.Model(&position).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update position"})
		return
	}

	// Reload with relationships
	ph.db.Preload("Department").Preload("Currency").First(&position, position.ID)

	c.JSON(http.StatusOK, position)
}

func (ph *PositionHandler) DeletePosition(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid position ID"})
		return
	}

	var position models.Position
	if err := ph.db.First(&position, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch position"})
		return
	}

	// Check if position is being used by any employees
	var employeeCount int64
	ph.db.Model(&models.Employee{}).Where("position_id = ?", position.ID).Count(&employeeCount)
	if employeeCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete position as it is assigned to employees"})
		return
	}

	if err := ph.db.Delete(&position).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete position"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Position deleted successfully"})
}

func (ph *PositionHandler) GetPositionsByDepartment(c *gin.Context) {
	departmentID, err := strconv.ParseUint(c.Param("departmentId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department ID"})
		return
	}

	var positions []models.Position
	if err := ph.db.Preload("Currency").Where("department_id = ? AND is_active = ?", departmentID, true).Find(&positions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch positions"})
		return
	}

	c.JSON(http.StatusOK, positions)
}
