package controllers

import (
	"crypto/rand"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"

	// "log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	// "github.com/shyamsundaar/karino-mock-server/models/delivery"
	// "gorm.io/gorm"
	// "github.com/gin-gonic/gin"
	"context"

	"github.com/shyamsundaar/karino-mock-server/initializers"
	"github.com/shyamsundaar/karino-mock-server/models/delivery"
	"github.com/shyamsundaar/karino-mock-server/models/sales"
	"github.com/shyamsundaar/karino-mock-server/query"
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

func GenerateNextDeliveryDocumentID() string {
	return uuid.New().String()
}

func generate9DigitID() string {
	b := make([]byte, 4) // 4 bytes is enough for a 9-digit number
	rand.Read(b)
	// Generate a number between 100,000,000 and 999,999,999
	return fmt.Sprintf("%09d", (uint32(b[0])|uint32(b[1])<<8|uint32(b[2])<<16|uint32(b[3])<<24)%900000000+100000000)
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

	for _, document := range chunks {
		deliveryDocCode, err := GenerateNextDeliveryDocumentCode(ctx, q)
		deliverydocumentId := GenerateNextOrderItemTempID()
		if err != nil {
			return err
		}

		for _, item := range document {
			stockKeepingUnit := generate9DigitID()

			deliveryItem := delivery.CreateDeliveryDocuments{
				CoopID:            coopId,
				ErpSalesOrderCode: payload.ErpSalesOrderCode,
				OrderID:           payload.OrderID,

				DeliveryDocumentID:   deliverydocumentId,
				DeliveryDocumentCode: deliveryDocCode,

				OrderItemID:      item.OrderItemID,
				StockKeppingUnit: stockKeepingUnit,
				CreatedAt:        &now,
				UpdatedAt:        &now,
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
// @Param        perPage         query     int     false  "Items per page" default(10)
// @Success      200    {object}  delivery.ListDeliveryDocumentsResponse
// @Router       /spic_to_erp/customers/{coopId}/salesorders/deliverydocuments [get]
func GetCustomerDeliveryDocumentDetailHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")
	updatedFrom := c.Query("updatedFrom")
	updatedTo := c.Query("updatedTo")
	//var salesorder []sales.SalesOrder
	
	if !isCoopAllowed(coopId) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"Message": "The indicated cooperative does not exist.",
			})
	}

	type SalesWithDelivery struct {
		TempID            string `gorm:"column:temp_id"`
		ErpSalesOrderId   string `gorm:"column:erp_sales_order_id"`
		ErpSalesOrderCode string `gorm:"column:erp_sales_order_code"`
		OrderID           string `gorm:"column:order_id"`
		// ErpDeliveryDocumentCode string `gorm:"column:erp_delivery_document_code"`
	}
	var results []SalesWithDelivery

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("perPage", "10"))
	if perPage < 0 {
		perPage = 10
	}
	offset := (page - 1) * perPage
	var totalRecords int64

	query := initializers.DB.
		Table("sales_orders").
		Select("sales_orders.temp_id, sales_orders.erp_sales_order_id, sales_orders.erp_sales_order_code, sales_orders.order_id").
		Joins("JOIN delivery_documents ON delivery_documents.order_id = sales_orders.order_id").
		Where("sales_orders.coop_id = ?", coopId)

	if updatedFrom != "" && updatedTo != "" {
		fromTime, err := time.Parse(time.RFC3339, updatedFrom)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid updatedFrom format. Use ISO8601 (YYYY-MM-DDTHH:MM:SSZ)",
			})
		}

		toTime, err := time.Parse(time.RFC3339, updatedTo)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid updatedTo format. Use ISO8601 (YYYY-MM-DDTHH:MM:SSZ)",
			})
		}
		query = query.Where("delivery_documents.updated_at >= ? AND delivery_documents.updated_at <= ?", fromTime, toTime)
	}

	query.Select("COUNT(DISTINCT sales_orders.order_id)").Count(&totalRecords)

	if err := query.
		Select("sales_orders.temp_id, sales_orders.erp_sales_order_id, sales_orders.erp_sales_order_code, sales_orders.order_id").
		Group("sales_orders.order_id").
		Limit(perPage).
		Offset(offset).
		Scan(&results).Error; err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(sales.ErrorSalesOrderResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	data := make([]delivery.DeliverydocumentsListResponse, 0)
	for _, f := range results {
		data = append(data, delivery.DeliverydocumentsListResponse{
			TempERPSalesOrderId: f.TempID,
			ErpSalesOrderId:     f.ErpSalesOrderId,
			ErpSalesOrderCode:   f.ErpSalesOrderCode,
			SpicSalesOrderId:    f.OrderID,
		})
	}

	return c.Status(fiber.StatusOK).JSON(delivery.ListDeliveryDocumentsResponse{
		Data: data,
		Pagination: delivery.PaginationInfo{
			Page:        page,
			Limit:       perPage,
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
	coopId := c.Params("coopId")
	orderID := c.Params("orderId")
	emptydata := make([]sales.SalesOrderListResponse, 0)

	// if orderID == "" {
	// 	return c.Status(400).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "order_id is required",
	// 	})
	// }
	if !isCoopAllowed(coopId) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"Message": "The indicated cooperative does not exist.",
			})
	}
	var order sales.SalesOrder
	if err := initializers.DB.
		Where("order_id = ? AND coop_id = ?", orderID, coopId).
		First(&order).Error; err != nil {

		return c.Status(404).JSON(fiber.Map{
			"deliverynotes": emptydata,
		})
	}
	var orderItems []sales.SalesOrderItem
	if err := initializers.DB.
		Where("order_id = ?", orderID).
		Find(&orderItems).Error; err != nil {

		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch order items",
		})
	}
	var deliveryDocs []delivery.CreateDeliveryDocuments
	if err := initializers.DB.
		Where("order_id = ? AND coop_id = ?", orderID, coopId).
		Find(&deliveryDocs).Error; err != nil {

		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch delivery documents",
		})
	}
	deliveryMap := make(map[string][]delivery.CreateDeliveryDocuments)

	for _, doc := range deliveryDocs {
		deliveryMap[doc.DeliveryDocumentCode] =
			append(deliveryMap[doc.DeliveryDocumentCode], doc)
	}
	var response delivery.DeliveryNotesResponse

	for docCode, docs := range deliveryMap {

		note := delivery.DeliveryNote{
			ERPDeliveryDocumentId:   docs[0].DeliveryDocumentID,
			ERPDeliveryDocumentCode: docCode,
			ERPDeliveryDocumentDate: *docs[0].CreatedAt,
		}

		for _, d := range docs {
			for _, item := range orderItems {
				if item.OrderItemID == d.OrderItemID {

					note.Items = append(note.Items, delivery.DeliveryItem{
						ERPItemID2:       item.ErpItemID2,
						StockKeepingUnit: item.StockKeepingUnit,
						Quantity:         item.Quantity,
						SalesOrder: delivery.DeliverySalesOrder{
							TempERPSalesOrderId: order.TempID,
							ERPSalesOrderId:     order.ErpSalesOrderId,
							ERPSalesOrderCode:   order.ErpSalesOrderCode,
							SPICSalesOrderId:    order.OrderID,
							ERPItemID:           item.ErpItemID,
							OrderItemID:         item.OrderItemID,
						},
					})
				}
			}
		}

		response.DeliveryNotes = append(response.DeliveryNotes, note)
	}
	return c.Status(200).JSON(response)
}
