package controllers

import (
	"math"
	"strconv"

	// "strings"
	// "fmt"
	// "regexp"
	// "context"
	// "log"
	"time"

	// "database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/shyamsundaar/karino-mock-server/initializers"
	"github.com/shyamsundaar/karino-mock-server/models/sales"

	// "github.com/google/uuid"
	// "github.com/shyamsundaar/karino-mock-server/models/farmers"
	// "karino-mock-server/query"
	// "github.com/shyamsundaar/karino-mock-server/query"
	"gorm.io/gorm"
)

// CreateCustomerSalesDetailHandler handles POST /spic_to_erp/customers/:coopId/salesorders
// @Summary      Create a new sales order detail
// @Description  Create a new record in the sales orders table
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        coopId  path      string                            true  "Cooperative ID"
// @Param        detail  body      sales.SalesOrder    true  "Create order Payload"
// @Success      201     {object}  sales.CreateSalesOrderResponse
// @Router       /spic_to_erp/customers/{coopId}/salesorders [post]
func CreateCustomerSalesOrderHandler(c *fiber.Ctx) error {
	// 1. Get CoopID from URL
	coopId := c.Params("coopId")

	// 2. Parse request body
	var payload sales.SalesOrder
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(sales.ErrorSalesOrderResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	// 3. Basic validations (keep minimal like farmer handler)
	if payload.OrderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(sales.ErrorSalesOrderResponse{
			Success: false,
			Message: "order_id is required",
		})
	}

	if payload.FarmerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(sales.ErrorSalesOrderResponse{
			Success: false,
			Message: "farmer_id is required",
		})
	}

	// 4. Map payload → SalesOrder DB model
	newOrder := sales.SalesOrder{
		// TempID:                 TempID,
		CoopID: coopId,

		OrderID:     payload.OrderID,
		OrderNumber: payload.OrderNumber,
		ContractID:  payload.ContractID,

		FarmerID:   payload.FarmerID,
		FarmerName: payload.FarmerName,

		ClubID:   payload.ClubID,
		ClubName: payload.ClubName,

		FarmerResourceCategory: payload.FarmerResourceCategory,
		ContractCrop:           payload.ContractCrop,
		ContractCropVareity:    payload.ContractCropVareity,
		ContractArea:           payload.ContractArea,

		SponsorID:   payload.SponsorID,
		SponsorName: payload.SponsorName,

		BuyerID:   payload.BuyerID,
		BuyerName: payload.BuyerName,

		PackageSetCaptionPT: payload.PackageSetCaptionPT,

		RegionID:         payload.RegionID,
		RegionPartID:     payload.RegionPartID,
		SettlementID:     payload.SettlementID,
		SettlementPartID: payload.SettlementPartID,

		CustomZone1ID: payload.CustomZone1ID,
		CustomZone2ID: payload.CustomZone2ID,

		PickupDate: payload.PickupDate,
		CreatedBy:  payload.CreatedBy,
	}

	// 5. DB transaction (parent + children)
	err := initializers.DB.Transaction(func(tx *gorm.DB) error {

		// Save sales order
		if err := tx.Create(&newOrder).Error; err != nil {
			return err
		}

		// Map & save order items
		if len(payload.OrderItems) > 0 {
			var items []sales.SalesOrderItem

			for _, item := range payload.OrderItems {
				items = append(items, sales.SalesOrderItem{
					OrderID:              newOrder.OrderID,
					OrderItemID:          item.OrderItemID,
					OrderItemNumber:      item.OrderItemNumber,
					StockKeepingUnit:     item.StockKeepingUnit,
					ProductGroup:         item.ProductGroup,
					InputItemID:          item.InputItemID,
					InputItemName:        item.InputItemName,
					InputItemNameCaption: item.InputItemNameCaption,
					Quantity:             item.Quantity,
					QuantityUnitKey:      item.QuantityUnitKey,
					UnitPrice:            item.UnitPrice,
					Price:                item.Price,
					PriceUnitKey:         item.PriceUnitKey,
					NumberOfUnits:        item.NumberOfUnits,
				})
			}

			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(sales.ErrorSalesOrderResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	// 6. Response DTO (exactly as you defined)
	response := sales.CreateSalesOrderResponse{
		Success: true,
		Data: sales.CreateSalesOrderResponseData{
			TempERPSalesOrderId: newOrder.TempID,
			ErpSalesOrderId:     "", // populate later if async
			ErpSalesOrderCode:   "",
			SpicSalesOrderId:    newOrder.OrderID,
			CreatedAt:           newOrder.CreatedAt.Format(time.RFC3339),
			UpdatedAt:           newOrder.UpdatedAt.Format(time.RFC3339),
			Message:             "Sales order created successfully",
		},
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetCustomerSalesDetailHandler handles GET /spic_to_erp/customers/:coopId/salesorders
// @Summary      List salesorder updated within date ranges
// @Description  Get a paginated list of farmer details for a specific cooperative
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        updatedFrom   query     string  false  " "
// @Param        updatedTo     query     string  false  " "
// @Param        page          query     int     false  "Page number"    default(1)
// @Param        limit         query     int     false  "Items per page" default(10)
// @Success      200    {object}  sales.ListSalesOrderResponse
// @Router       /spic_to_erp/customers/{coopId}/salesorders [get]
func GetCustomerSalesDetailHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")
	var salesorder []sales.SalesOrder

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit
	var totalRecords int64

	query := initializers.DB.
		Model(&sales.SalesOrder{}).
		Where("coop_id = ?", coopId)

	query.Count(&totalRecords)

	if err := query.
		Limit(limit).
		Offset(offset).
		Find(&salesorder).Error; err != nil {

		return c.Status(fiber.StatusBadGateway).JSON(sales.ErrorSalesOrderResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))

	var data []sales.SalesOrderListResponse
	for _, f := range salesorder {
		data = append(data, sales.SalesOrderListResponse{
			OrderID:     f.OrderID,
			OrderNumber: f.OrderNumber,
			FarmerID:    f.FarmerID,
			FarmerName:  f.FarmerName,
			ClubID:      f.ClubID,
			ClubName:    f.ClubName,
			CreatedAt:   f.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   f.UpdatedAt.Format(time.RFC3339),
		})
	}

	return c.Status(fiber.StatusOK).JSON(sales.ListSalesOrderResponse{
		Data: data,
		Pagination: sales.PaginationInfo{
			Page:        page,
			Limit:       limit,
			TotalItems:  int(totalRecords),
			TotalPages:  totalPages,
			HasPrevious: page > 1,
			HasNext:     page < totalPages,
		},
	})
	// return c.Status(fiber.StatusCreated).JSON(response)
}

// FindSalesOrderDetails handles GET /spic_to_erp/customers/:coopId/salesorders/:salesordersid
// @Summary      Get salesorder details
// @Description  Get a paginated list of SalesOrder details for a specific cooperative
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        salesordersid path      string  true   " "
// @Success      200    {object}  sales.SalesOrderAmountResponse
// @Router       /spic_to_erp/customers/{coopId}/salesorders/{salesordersid} [get]
func GetCustomerSalesOrderDetailsHandler(c *fiber.Ctx) error {
	// Implementation for retrieving sales order details
	response := sales.SalesOrderAmountResponse{
		Message:             "Sales order amount calculated successfully",
		TempERPSalesOrderId: "TEMP-SO-001",
		ErpSalesOrderId:     "ERP-SO-10001",
		ErpSalesOrderCode:   "SO2026-001",
		SpicSalesOrderId:    "SPIC-SO-7788",
		CreatedAt:           time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:           time.Now().UTC().Format(time.RFC3339),
		OrderValue:          12500.50,
		TaxAmount:           2250.09,
		TotalAmount:         14750.59,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}
