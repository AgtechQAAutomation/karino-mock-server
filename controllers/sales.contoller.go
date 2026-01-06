package controllers

import (
	// "math"
	// "strconv"

	// "strings"
	// "fmt"
	// "regexp"
	// "context"
	// "log"
	"time"

	// "database/sql"
	
	"github.com/gofiber/fiber/v2"
	"github.com/shyamsundaar/karino-mock-server/models/sales"
	"github.com/shyamsundaar/karino-mock-server/initializers"
	

	// "github.com/shyamsundaar/karino-mock-server/models/farmers"
	// "karino-mock-server/query"
	// "github.com/shyamsundaar/karino-mock-server/query"
	// "gorm.io/gorm"
)

// CreateCustomerSalesDetailHandler handles POST /spic_to_erp/customers/:coopId/salesorders
// @Summary      Create a new sales order detail
// @Description  Create a new record in the sales orders table
// @Tags         salesorders
// @Accept       json
// @Produce      json
// @Param        coopId  path      string                            true  "Cooperative ID"
// @Param        detail  body      sales.Order          true  "Create Detail Payload"
// @Success      201     {object}  sales.CreateSalesOrderResponse
// @Router       /spic_to_erp/customers/{coopId}/salesorders [post]
func CreateCustomerSalesDetailHandler(c *fiber.Ctx) error {
	// 1. Get CoopID from URL Parameter
	coopId := c.Params("coopId")

	// 2. Parse request body
	var payload sales.Order
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	// 3. Map request → DB model
	newDetail := sales.Order{
		OrderID:                payload.OrderID,
		OrderNumber:            payload.OrderNumber,
		ContractID:             payload.ContractID,
		FarmerID:               payload.FarmerID,
		FarmerName:             payload.FarmerName,

		// Coop comes from URL
		ClubID:                 coopId,
		ClubName:               payload.ClubName,

		FarmerResourceCategory: payload.FarmerResourceCategory,
		ContractCrop:           payload.ContractCrop,
		ContractCropVareity:    payload.ContractCropVareity,
		ContractArea:           payload.ContractArea,

		SponsorID:              payload.SponsorID,
		SponsorName:            payload.SponsorName,
		BuyerID:                payload.BuyerID,
		BuyerName:              payload.BuyerName,

		PackageSetCaptionPT:    payload.PackageSetCaptionPT,

		RegionID:               payload.RegionID,
		RegionPartID:           payload.RegionPartID,
		SettlementID:           payload.SettlementID,
		SettlementPartID:       payload.SettlementPartID,

		CustomZone1ID:          payload.CustomZone1ID,
		CustomZone2ID:          payload.CustomZone2ID,

		PickupDate:             payload.PickupDate,
		CreatedBy:              payload.CreatedBy,

		OrderItems:             payload.OrderItems,
	}

	// 4. Save to DB
	if err := initializers.DB.Create(&newDetail).Error; err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 5. Response DTO
	response := sales.CreateSalesOrderResponse{
	Success: true,
	Data: sales.CreateSalesOrderResponseData{
		TempERPSalesOrderId: "TEMP-SO-001",
		ErpSalesOrderId:     "ERP-SO-10001",
		ErpSalesOrderCode:   "SO2026-001",
		SpicSalesOrderId:    "SPIC-SO-7788",
		CreatedAt:           time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:           time.Now().UTC().Format(time.RFC3339),
		Message:             "Sales order created successfully",
	},
}
	return c.Status(fiber.StatusCreated).JSON(response)
}

