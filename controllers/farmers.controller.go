package controllers

import (
	"strconv"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/wpcodevo/golang-fiber-mysql/initializers"
	"github.com/wpcodevo/golang-fiber-mysql/models"
	"github.com/wpcodevo/golang-fiber-mysql/query"
	"gorm.io/gorm"
)

// CreateDetailHandler handles POST /api/details
// @Summary      Create a new farmer detail
// @Description  Create a new record in the details table
// @Tags         Details
// @Accept       json
// @Produce      json
// @Param        detail  body      models.CreateDetailSchema  true  "Create Detail Payload"
// @Success      201     {object}  models.Detail
// @Failure      400     {array}   models.ErrorResponse
// @Failure      502      {object}  map[string]interface{}
// @Router       /details [post]
func CreateDetailHandler(c *fiber.Ctx) error {
	var payload *models.CreateDetailSchema

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	errors := models.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	newDetail := models.Detail{
		CoopID:             payload.CoopID,
		FarmerID:           payload.FarmerID,
		FirstName:          payload.FirstName,
		LastName:           payload.LastName,
		FarmerKycID:        payload.FarmerKycID,
		ClubLeaderFarmerID: payload.ClubLeaderFarmerID,
	}

	result := initializers.DB.Create(&newDetail)
	if result.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": result.Error.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": fiber.Map{"detail": newDetail}})
}

// FindDetails handles GET /api/details
// @Summary      List farmer details
// @Description  Get a paginated list of farmer details
// @Tags         Details
// @Accept       json
// @Produce      json
// @Param        page   query     int  false  "Page number"   default(1)
// @Param        limit  query     int  false  "Items per page" default(10)
// @Success      200    {object}  models.Detail
// @Router       /details [get]
func FindDetails(c *fiber.Ctx) error {
	// 1. Use the generated Query Agent
	q := query.Use(initializers.DB)
	d := q.Detail

	// 2. Perform Type-Safe pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// The CLI provides methods like Offset and Limit directly
	results, count, err := d.FindByPage((page-1)*limit, limit)

	if err != nil {
		return c.Status(502).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"results": count,
		"data":    results,
	})
}

// UpdateDetail handles PATCH /api/details/:detailId
// @Summary      Update a farmer detail
// @Description  Update fields of an existing detail record
// @Tags         Details
// @Accept       json
// @Produce      json
// @Param        detailId  path      string                     true  "Detail ID"
// @Param        detail    body      models.UpdateDetailSchema  true  "Update Payload"
// @Success      200       {object}  models.Detail
// @Failure      502      {object}  map[string]interface{}
// @Router       /details/{detailId} [patch]
func UpdateDetail(c *fiber.Ctx) error {
	detailId := c.Params("detailId")
	var payload *models.UpdateDetailSchema

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	var detail models.Detail
	result := initializers.DB.First(&detail, "id = ?", detailId)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "No detail with that ID exists"})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	updates := make(map[string]interface{})
	if payload.FirstName != "" {
		updates["first_name"] = payload.FirstName
	}
	if payload.LastName != "" {
		updates["last_name"] = payload.LastName
	}
	if payload.MobileNumber != "" {
		updates["mobile_number"] = payload.MobileNumber
	}
	if payload.ZipCode != "" {
		updates["zip_code"] = payload.ZipCode
	}

	updates["updated_at"] = time.Now()
	initializers.DB.Model(&detail).Updates(updates)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"detail": detail}})
}

// FindDetailById handles GET /api/details/:detailId
// @Summary      Get detail by ID
// @Description  Retrieve a single farmer detail record
// @Tags         Details
// @Produce      json
// @Param        detailId  path      string  true  "Detail ID"
// @Success      200       {object}  models.Detail
// @Failure      502      {object}  map[string]interface{}
// @Router       /details/{detailId} [get]
func FindDetailById(c *fiber.Ctx) error {
	detailId := c.Params("detailId")

	var detail models.Detail
	result := initializers.DB.First(&detail, "id = ?", detailId)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "No detail with that ID exists"})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"detail": detail}})
}

// DeleteDetail handles DELETE /api/details/:detailId
// @Summary      Delete a detail
// @Description  Remove a farmer detail record from the database
// @Tags         Details
// @Param        detailId  path      string  true  "Detail ID"
// @Success      204       "No Content"
// @Failure      502      {object}  map[string]interface{}
// @Router       /details/{detailId} [delete]
func DeleteDetail(c *fiber.Ctx) error {
	detailId := c.Params("detailId")

	result := initializers.DB.Delete(&models.Detail{}, "id = ?", detailId)

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "No detail with that ID exists"})
	} else if result.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": result.Error.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
