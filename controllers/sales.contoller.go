package controllers

import (
	"math"
	"strconv"

	// "strings"
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	// "database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/shyamsundaar/karino-mock-server/initializers"
	models "github.com/shyamsundaar/karino-mock-server/models/farmers"
	"github.com/shyamsundaar/karino-mock-server/models/products"
	"github.com/shyamsundaar/karino-mock-server/models/sales"
	"github.com/google/uuid"
	// "github.com/google/uuid"
	// "github.com/shyamsundaar/karino-mock-server/models/farmers"
	// "karino-mock-server/query"
	"github.com/shyamsundaar/karino-mock-server/query"
	"gorm.io/gorm"
)


func GenerateAndSetNextErpSalesOrderIDGen(
	ctx context.Context,
	q *query.Query,
	salesOrderID uint,
) (string, error) {

	so := q.SalesOrder.WithContext(ctx)

	// 1. Fetch current sales order row
	row, err := so.
		Where(q.SalesOrder.ID.Eq(salesOrderID)).
		First()
	if err != nil {
		return "", err
	}

	// 2. If already generated → return
	if row.ErpSalesOrderId != "" {
		return row.ErpSalesOrderId, nil
	}

	// 3. Fetch last non-empty ERP sales order ID
	last, err := so.
		Where(q.SalesOrder.ErpSalesOrderId.Neq("")).
		Order(q.SalesOrder.ID.Desc()).
		First()

	next := 1
	if err == nil && last.ErpSalesOrderId != "" {
		re := regexp.MustCompile(`\d+$`)
		if m := re.FindString(last.ErpSalesOrderId); m != "" {
			n, _ := strconv.Atoi(m)
			next = n + 1
		}
	}

	// 4. Generate new ERP Sales Order ID
	newErpSalesOrderID := fmt.Sprintf("ERP-SO-%05d", next)

	// 5. Business delay
	time.Sleep(time.Duration(initializers.AppConfig.TimeSeconds) * time.Second)

	// 6. Update ONLY if still empty (race-condition safe)
	_, err = so.
		Where(
			q.SalesOrder.ID.Eq(salesOrderID),
			q.SalesOrder.ErpSalesOrderId.Eq(""),
		).
		UpdateColumnSimple(
			q.SalesOrder.ErpSalesOrderId.Value(newErpSalesOrderID),
			q.SalesOrder.UpdatedAt.Value(time.Now()),
		)

	if err != nil {
		return "", err
	}

	return newErpSalesOrderID, nil
}

func GenerateAndSetNextErpSalesOrderCodeGen(
	ctx context.Context,
	q *query.Query,
	ErpSalesOrderCode uint,
) (string, error) {

	so := q.SalesOrder.WithContext(ctx)

	// 1. Fetch current sales order row
	row, err := so.
		Where(q.SalesOrder.ID.Eq(ErpSalesOrderCode)).
		First()
	if err != nil {
		return "", err
	}

	// 2. If already generated → return
	if row.ErpSalesOrderCode != "" {
		return row.ErpSalesOrderCode, nil
	}

	next := 1

	last, err := so.
		Where(q.SalesOrder.ErpSalesOrderCode.Neq("")).
		Order(q.SalesOrder.ID.Desc()).
		First()

	if err == nil && last.ErpSalesOrderCode != "" {
		re := regexp.MustCompile(`\d+$`)
		if m := re.FindString(last.ErpSalesOrderCode); m != "" {
			n, _ := strconv.Atoi(m)
			next = n + 1
		}
	}

	newErpSalesOrderCode := fmt.Sprintf("ECL 2025/%d", next)

	// 5. Business delay
	time.Sleep(time.Duration(initializers.AppConfig.TimeSeconds) * time.Second)

	// 6. Update ONLY if still empty (race-condition safe)
	_, err = so.
		Where(
			q.SalesOrder.ID.Eq(ErpSalesOrderCode),
			q.SalesOrder.ErpSalesOrderCode.Eq(""),
		).
		UpdateColumnSimple(
			q.SalesOrder.ErpSalesOrderCode.Value(newErpSalesOrderCode),
			q.SalesOrder.UpdatedAt.Value(time.Now()),
		)

	if err != nil {
		return "", err
	}

	return newErpSalesOrderCode, nil
}

func GenerateNextOrderItemTempID() string {
	return uuid.New().String()
}

// CreateCustomerSalesDetailHandler handles POST /spic_to_erp/customers/:coopId/salesorders
// @Summary      Create a new sales order detail
// @Description  Create a new record in the sales orders table
// @Tags         salesoreder
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
	var payload *sales.CreateSalesOrderSchema
	var existingSalesOrder sales.SalesOrder
	var existingFarmer models.FarmerDetails
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(sales.ErrorSalesOrderResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	// 3. Basic validations (keep minimal like farmer handler)
	if !isCoopAllowed(coopId) {
		return SendSalesErrorResponse(c, "The indicated cooperative does not exist.", payload.OrderID)
	}
 
	if payload.OrderID == "" {
		return SendSalesErrorResponse(c, "You must specify the OrderID.", payload.OrderID)
	}

	orderId := initializers.DB.Where("order_id = ? AND coop_id = ?", payload.OrderID, coopId).First(&existingSalesOrder).Error

	if orderId == nil {
		return SendSalesErrorResponse(c, "The OrderId already exist.", payload.OrderID)
	}

	if payload.FarmerID == "" {
		return SendSalesErrorResponse(c, "You must provide the FarmerID.", payload.OrderID)
	}

	if payload.ContractID == "" {
		return SendSalesErrorResponse(c, "You must provide the ContractID.", payload.OrderID)
	}

	farmerId := initializers.DB.
		Where(
			"farmer_id = ? AND coop_id = ?",
			payload.FarmerID,
			coopId,
		).
		First(&existingFarmer).
		Error

	if farmerId != nil {
		return SendSalesErrorResponse(c, "The indicated FarmerId does not exist.", payload.OrderID)
	}

	for _, item := range payload.OrderItems {
		if item.ProductGroup == "" {
			return SendSalesErrorResponse(c, "You must specify the item code or group.", payload.OrderID)
		}

		var product products.Product
		productErr := initializers.DB.Where("product_code = ?", item.ProductGroup).First(&product).Error
		if productErr != nil {
			return SendSalesErrorResponse(c, "The indicated itemcode/group does not exist ().", payload.OrderID)
		}

		if item.Quantity <= 0 {
			return SendSalesErrorResponse(c, "The quantity of the product must be greater than zero.", payload.OrderID)
		}
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

		NoofOrderItems: len(payload.OrderItems),
	}

	// 5. DB transaction (parent + children)
	err := initializers.DB.Transaction(func(tx *gorm.DB) error {

		// Save sales order
		if err := tx.Create(&newOrder).Error; err != nil {
			return err
		}
	// ctx := context.Background()
	// q := query.Use(initializers.DB)
		// Map & save order items
		if len(payload.OrderItems) > 0 {
			var items []sales.SalesOrderItem
			
			for _, item := range payload.OrderItems {
				// erpItemID , err := GenerateNextOrderItemTempID(ctx, q)
				// if err != nil {
				// 	return err
				// }

				items = append(items, sales.SalesOrderItem{
					OrderID:              newOrder.OrderID,
					OrderItemID:          item.OrderItemID,
					OrderItemNumber:      item.OrderItemNumber,
					StockKeepingUnit:     item.StockKeepingUnit,
					ErpItemID:			 GenerateNextOrderItemTempID(),
					ErpItemID2:			 GenerateNextOrderItemTempID(),
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
	ctx := context.Background()
	q := query.Use(initializers.DB)

	go func(orderDBID uint) {
		_, err := GenerateAndSetNextErpSalesOrderIDGen(ctx, q, orderDBID)
		_, err1 := GenerateAndSetNextErpSalesOrderCodeGen(ctx, q, orderDBID)
		if err1 != nil {
			log.Println("❌ ERP SalesOrder Code generation failed:", err1)
		}
		if err != nil {
			log.Println("❌ ERP SalesOrder ID generation failed:", err)
		}
	}(newOrder.ID)

	// 6. Response DTO (exactly as you defined)
	response := sales.CreateSalesOrderResponse{
		Success: true,
		Data: sales.CreateSalesOrderResponseData{
			TempERPSalesOrderId: newOrder.TempID,
			ErpSalesOrderId:     "", // populate later if async
			ErpSalesOrderCode:   "",
			SpicSalesOrderId:    newOrder.OrderID,
			CreatedAt:           newOrder.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:           newOrder.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Message:             "Document saved with success.",
		},
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func SendSalesErrorResponse(c *fiber.Ctx, message string, orderId string) error {
	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"success": false,
		"data": fiber.Map{
			"tempERPSalesOrderId": "0",
			"erpSalesOrderId":     "",
			"erpSalesOrderCode":   "",
			"spicSalesOrderId":    "",
			"createdAt":           now,
			"updatedAt":           now,
			"message":             message,
		},
	})
}

// GetCustomerSalesDetailHandler handles GET /spic_to_erp/customers/:coopId/salesorders
// @Summary      List salesorder updated within date ranges
// @Description  Get a paginated list of farmer details for a specific cooperative
// @Tags         salesoreder
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
	updatedFrom := c.Query("updatedFrom")
	updatedTo := c.Query("updatedTo")
	var salesorder []sales.SalesOrder
	if !isCoopAllowed(coopId) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"Message": "The indicated cooperative does not exist.",
			})
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit
	var totalRecords int64

	query := initializers.DB.
		Model(&sales.SalesOrder{}).
		Where("coop_id = ?", coopId)
	
	if updatedFrom != "" && updatedTo != "" {
		fromTime, err := time.Parse(time.RFC3339, updatedFrom)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"Message": "Invalid updatedFrom format. Use ISO8601 (YYYY-MM-DDTHH:MM:SSZ)",
			})
		}

		toTime, err := time.Parse(time.RFC3339, updatedTo)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"Message": "Invalid updatedTo format. Use ISO8601 (YYYY-MM-DDTHH:MM:SSZ)",
			})
		}

		query = query.Where("updated_at>= ? AND updated_at<= ?", fromTime, toTime)
	}

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

	data := make([]sales.SalesOrderListResponse, 0)
	for _, f := range salesorder {
		data = append(data, sales.SalesOrderListResponse{
			TempERPSalesOrderId: f.TempID,
			ErpSalesOrderId:     f.ErpSalesOrderId,
			ErpSalesOrderCode:   f.ErpSalesOrderCode,
			SpicSalesOrderId:    f.OrderID,
			CreatedAt:           f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:           f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

// FindSalesOrderDetails handles GET /spic_to_erp/customers/:coopId/salesorders/:orderId
// @Summary      Get salesorder details
// @Description  Get a paginated list of SalesOrder details for a specific cooperative
// @Tags         salesoreder
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        orderId path      string  true   " "
// @Success      200    {object}  sales.SalesOrderAmountResponse
// @Router       /spic_to_erp/customers/{coopId}/salesorders/{orderId} [get]
func GetCustomerSalesOrderDetailsHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")
	orderId := c.Params("orderId")

	var salesOrder sales.SalesOrder
	if !isCoopAllowed(coopId) {
		return SendOrderIdErrorResponse(c, "The indicated cooperative does not exist.", orderId)
	}

	err := initializers.DB.Where("coop_id = ? AND order_id = ?", coopId, orderId).First(&salesOrder).Error

	if err != nil {
		return SendOrderIdErrorResponse(c, "There is no order with the indicated OrderID.", orderId)
	}

	// Implementation for retrieving sales order details
	response := sales.SalesOrderAmountResponse{
		Message:             "",
		TempERPSalesOrderId: salesOrder.TempID,
		ErpSalesOrderId:     salesOrder.ErpSalesOrderId,
		ErpSalesOrderCode:   salesOrder.ErpSalesOrderCode,
		SpicSalesOrderId:    salesOrder.OrderID,
		CreatedAt:           time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		OrderValue:          12500.50,
		TaxAmount:           2250.09,
		TotalAmount:         14750.59,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func SendOrderIdErrorResponse(c *fiber.Ctx, msg string, orderId string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"Message":             msg,
		"TempERPSalesOrderId": "",
		"ErpSalesOrderId":     "",
		"ErpSalesOrderCode":   "",
		"SpicSalesOrderId":    orderId,
		"CreatedAt":           "1900-01-01T00:00:00",
		"updatedAt":           "1900-01-01T00:00:00",
		"orderValue":          0.0,
		"taxAmount":           0.0,
		"totalAmount":         0.0,
	})

}
