package controllers

import (
	"math"
	"regexp"
	"strconv"
	"fmt"
	"time"
	// "log"

	"github.com/gofiber/fiber/v2"

	// "github.com/shyamsundaar/karino-mock-server/models/delivery"
	// "gorm.io/gorm"
	// "github.com/gin-gonic/gin"
	"github.com/shyamsundaar/karino-mock-server/initializers"
	"github.com/shyamsundaar/karino-mock-server/models/delivery"
	"context"
	"github.com/shyamsundaar/karino-mock-server/query"
	"github.com/shyamsundaar/karino-mock-server/models/sales"
)

func GenerateNextDeliveryDocumentCode(
	ctx context.Context,
	q *query.Query,
) (string, error) {

	dd := q.CreateDeliveryDocuments.WithContext(ctx)

	last, err := dd.
		Where(q.CreateDeliveryDocuments.DeliveryDocumentCode.Neq("")).
		Order(q.CreateDeliveryDocuments.Id.Desc()).
		First()

	next := 1
	if err == nil && last.DeliveryDocumentCode != "" {
		re := regexp.MustCompile(`\d+$`)
		if m := re.FindString(last.DeliveryDocumentCode); m != "" {
			n, _ := strconv.Atoi(m)
			next = n + 1
		}
	}

	// ✅ incrementing number
	return fmt.Sprintf("GT2 2025/%d", next), nil
}


func GenerateAndSetNextERPItemIdGen(
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

// CreateCustomerDeliveryDocumentDetailsHandler handles POST /spic_to_erp/customers/:coopId/salesorders/deliverydocuments
// @Summary      Create deliverydocuments details for a sales order
// @Description  Create deliverydocuments details for a sales order
// @Tags         deliverydocuments
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        detail  body      delivery.CreateDeliveryDocumentSchema    true  "Create delivery document Payload"
// @Success      200    {object}  delivery.CreateDeliveryDocumentSuccessResponse
// @Router       /spic_to_erp/customers/{coopId}/salesorders/deliverydocuments [post]
func CreateCustomerDeliveryDocumentDetailsHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")

	var payload *delivery.CreateDeliveryDocumentSchema
	var salesOrder sales.SalesOrder
	var deliverydocument delivery.CreateDeliveryDocuments

	//var salesOrderItems sales.SalesOrderItem
	var salesOrderItemsList []sales.SalesOrderItem
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	var noof_order_items int

	salesErr := initializers.DB.Where("order_id = ? AND erp_sales_order_code = ?", payload.OrderID, payload.ErpSalesOrderCode).First(&salesOrder).Error
	if salesErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Message": "OrderId or SalesOrder not found "})
	}

	deliverydocumenterr := initializers.DB.Where("order_id = ?", payload.OrderID).First(&deliverydocument).Error

	if deliverydocumenterr == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Message": "Delivery Documents already Created for the OrderId"})
	}

	deliverydocumentserr := initializers.DB.Model(&salesOrder).Where("order_id = ?", payload.OrderID).Pluck("noof_order_items", &noof_order_items).Error

	if deliverydocumentserr != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Message": "Error fetching delivery documents"})
	}

	if payload.NoofDeliveryDocuments > noof_order_items {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Message": "Number of delivery documents cannot be greater than number of order items"})
	}

	orderItemserr := initializers.DB.Where("order_id = ?", payload.OrderID).Find(&salesOrderItemsList).Error
	if orderItemserr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Message": "No items found"})
	}

	// Split salesOrderItemsList into N chunks
	n := payload.NoofDeliveryDocuments
	total := len(salesOrderItemsList)

	if n <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Message": "NoofDeliveryDocuments must be greater than 0",
		})
	}

	// Calculate size of each chunk
	chunkSize := total / n
	remainder := total % n

	var chunks [][]sales.SalesOrderItem

	start := 0
	for range n {
		end := start + chunkSize
		if remainder > 0 {
			end++
			remainder--
		}

		if end > total {
			end = total
		}

		chunks = append(chunks, salesOrderItemsList[start:end])
		start = end
	}
	now := time.Now().UTC()

	ctx := context.Background()
	q := query.Use(initializers.DB)

		
	for docIndex, document := range chunks {
	deliveryDocCode, err := GenerateNextDeliveryDocumentCode(ctx, q)
	if err != nil {
		return err
	}

	for _, item := range document {

		deliveryItem := delivery.CreateDeliveryDocuments{
			CoopID:               coopId,
			ErpSalesOrderCode:    payload.ErpSalesOrderCode,
			OrderID:              payload.OrderID,

			DeliveryDocumentID:   strconv.Itoa(docIndex + 1),
			DeliveryDocumentCode: deliveryDocCode, 

			OrderItemID:          item.OrderItemID,
			CreatedAt:            &now,
			UpdatedAt:            &now,
		}

		if err := initializers.DB.Create(&deliveryItem).Error; err != nil {
			return err
		}
	}
}


	// for docIndex, document := range chunks { // Each delivery doc
	// 	for _, item := range document { // Each item inside doc
	// 		// Save in DB
	// 		deliveryItem := delivery.CreateDeliveryDocuments{
	// 			CoopID:               coopId,
	// 			ErpSalesOrderCode:    payload.ErpSalesOrderCode,
	// 			OrderID:              payload.OrderID,
	// 			DeliveryDocumentID:   strconv.Itoa(docIndex + 1),
	// 			DeliveryDocumentCode: "",
	// 			OrderItemID:          item.OrderItemID,
	// 			CreatedAt:            &now,
	// 			UpdatedAt:            &now,
	// 		}
	// 		initializers.DB.Create(&deliveryItem)
	// 	}
	// }
	
	// Response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"deliveryDocuments": chunks,
	})

	// return c.Status(fiber.StatusCreated).JSON(salesOrderItemsList)
}

// GetCustomerDeliveryDocumentDetailHandler handles GET /spic_to_erp/customers/:coopId/salesorders/deliverydocuments
// @Summary      List salesorder updated within date ranges
// @Description  Get a paginated list of farmer details for a specific cooperative
// @Tags         deliverydocuments
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        updatedFrom   query     string  false  " "
// @Param        updatedTo     query     string  false  " "
// @Param        page          query     int     false  "Page number"    default(1)
// @Param        limit         query     int     false  "Items per page" default(10)
// @Success      200    {object}  sales.ListSalesOrderResponse
// @Router       /spic_to_erp/customers/{coopId}/salesorders/deliverydocuments [get]
func GetCustomerDeliveryDocumentDetailHandler(c *fiber.Ctx) error {
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
}

// GetDeliveryDetailParticularHandler handles GET /spic_to_erp/customers/:coopId/salesorders/:orderId/deliverydocuments
// @Summary      List deliverydocuments details for a sales order
// @Description  Get a paginated list of farmer details for a specific cooperative
// @Tags         deliverydocuments
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        orderId path      string  true   " "
// @Success      200    {object}  delivery.DeliveryNotesResponse
// @Router       /spic_to_erp/customers/{coopId}/salesorders/{orderId}/deliverydocuments [get]
func GetDeliveryDetailParticularHandler(c *fiber.Ctx) error {
	
	var salesorder []sales.SalesOrder

	data := make([]delivery.DeliverySalesOrder, 0)
	
	for _, f := range salesorder {
		data = append(data, delivery.DeliverySalesOrder{
			TempERPSalesOrderId: f.TempID,
			ERPSalesOrderId:     f.ErpSalesOrderId,
			ERPSalesOrderCode:   f.ErpSalesOrderCode,
			SPICSalesOrderId:    f.OrderID,
			ERPItemID:           f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			OrderItemID:           f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.Status(fiber.StatusOK).JSON(delivery.DeliveryNotesResponse{
		
	})
}
