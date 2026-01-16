package controllers

import (
	"fmt"
	"math"
	"time"

	"github.com/AgtechQAAutomation/karino-mock-server/initializers"
	"github.com/gofiber/fiber/v2"

	// "github.com/AgtechQAAutomation/karino-mock-server/models/delivery"
	"github.com/AgtechQAAutomation/karino-mock-server/models/delivery"
	"github.com/AgtechQAAutomation/karino-mock-server/models/deliveryproof"

	// "context"
	"strconv"

	"github.com/google/uuid"
	// "github.com/AgtechQAAutomation/karino-mock-server/query"
)

func GenerateAndSetNextERPproofIDGen() string {
	return uuid.New().String()
}

// CreateDeliveryDocumentsProofHandler handles POST /spic_to_erp/customers/:coopId/deliverydocuments/:deliveryNoteId/proof
// @Summary      Create deliverydocuments proof for a sales order
// @Description  Create deliverydocuments proof for a sales order
// @Tags         deliverydocuments proof
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        deliveryNoteId path      string  true   " "
// @Param        detail  body      deliveryproof.CreateDeliveryDocumentProofSchema    true  "Create delivery document Proof Payload"
// @Success      200    {object}  deliveryproof.CreateDocumentdeliveryProofSuccessResponse
// @Router       /spic_to_erp/customers/{coopId}/deliverydocuments/{deliveryNoteId}/proof [post]
func CreateDeliveryDocumentsProofHandler(c *fiber.Ctx) error {
	var payload deliveryproof.CreateDeliveryDocumentProofSchema
	coopId := c.Params("coopId")
	deliveryNoteId := c.Params("deliveryNoteId")

	// 1. Parse JSON
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// 2. Coop check
	if !isCoopAllowed(coopId) {
		return SendDocumentdeliveryProofErrorResponse(c, "The indicated cooperative does not exist.")
	}

	// 3. Basic Delivery Note validations
	if payload.Waybill.DeliveryNoteID == "" {
		return SendDocumentdeliveryProofErrorResponse(c, "Please indicate the ID of the delivery guide.")
	}

	if deliveryNoteId != payload.Waybill.DeliveryNoteID {
		return SendDocumentdeliveryProofErrorResponse(c, "The guide ID, URL and information sent are not the same.")
	}

	// 4. Check existing delivery note is not expired
	var existingdeliveryNoteId delivery.CreateDeliveryDocuments
	err := initializers.DB.
		Where("delivery_document_id = ? AND coop_id = ? AND status = ?",
			deliveryNoteId, coopId, "NOT EXPIRED").
		First(&existingdeliveryNoteId).Error

	if err != nil {
		return SendDocumentdeliveryProofErrorResponse(c, "The delivery guide has been canceled or is no longer pending.")
	}

	// -----------------------------------------------------------
	// 5. MERGE DUPLICATE ITEMS BY STOCK KEEPING UNIT
	// -----------------------------------------------------------
	merged := make(map[string]deliveryproof.WaybillItem)

	for _, item := range payload.WaybillItems {
		sku := item.StockKeepingUnit

		// Already exists → add counts
		if existing, found := merged[sku]; found {
			existing.NumberOfUnits += item.NumberOfUnits
			existing.Quantity += item.Quantity
			merged[sku] = existing
		} else {
			// Convert payload struct → DB Struct
			merged[sku] = deliveryproof.WaybillItem{
				OrderID:          payload.Waybill.OrderID,
				Name:             item.Name,
				NumberOfUnits:    item.NumberOfUnits,
				Quantity:         item.Quantity,
				QuantityUnitKey:  item.QuantityUnitKey,
				UnitPrice:        item.UnitPrice,
				Price:            item.Price,
				PriceUnitKey:     item.PriceUnitKey,
				Status:           item.Status,
				StockKeepingUnit: item.StockKeepingUnit,
			}
		}
	}

	// Map → slice
	var uniqueItems []deliveryproof.WaybillItem
	for _, v := range merged {
		uniqueItems = append(uniqueItems, v)
	}

	// -----------------------------------------------------------
	// 6. VALIDATE merged items BEFORE INSERTING ANYTHING
	// -----------------------------------------------------------
	var validatedItems []deliveryproof.WaybillItem

	for _, item := range uniqueItems {
		var deliveryNotes []delivery.CreateDeliveryDocuments

		findErr := initializers.DB.
			Where("delivery_document_id = ? AND stock_kepping_unit = ?",
				deliveryNoteId, item.StockKeepingUnit).
			Find(&deliveryNotes).
			Error

		if findErr != nil {
			return SendDocumentdeliveryProofErrorResponse(c,
				"Database error while checking item ("+item.StockKeepingUnit+")")
		}

		if len(deliveryNotes) == 0 {
			return SendDocumentdeliveryProofErrorResponse(c,
				"The indicated item does not exist ("+item.StockKeepingUnit+").")
		}

		// Passed validation
		validatedItems = append(validatedItems, item)
	}

	// -----------------------------------------------------------
	// 7. ALL VALID — INSERT WAYBILL NOW (NO INSERTS BEFORE THIS)
	// -----------------------------------------------------------
	newWaybill := deliveryproof.Waybill{
		ContractID:           payload.Waybill.ContractID,
		CoopID:               coopId,
		OrderID:              payload.Waybill.OrderID,
		RegionID:             payload.Waybill.RegionID,
		RegionPartID:         payload.Waybill.RegionPartID,
		SettlementID:         payload.Waybill.SettlementID,
		SettlementPartID:     payload.Waybill.SettlementPartID,
		CustomZone1ID:        payload.Waybill.CustomZone1ID,
		CustomZone2ID:        payload.Waybill.CustomZone2ID,
		SalesOrderID:         payload.Waybill.SalesOrderID,
		SponsorName:          payload.Waybill.SponsorName,
		CustomerID:           payload.Waybill.CustomerID,
		DeliveryNoteID:       payload.Waybill.DeliveryNoteID,
		DeliveryNoteDocument: payload.Waybill.DeliveryNoteDocument,
		DeliveryPhotos: fmt.Sprintf(`[
            {"url1":"%s","url2":"%s"}
        ]`,
			payload.Waybill.DeliveryPhotoProofURL1,
			payload.Waybill.DeliveryPhotoProofURL2,
		),
	}

	if err := initializers.DB.Create(&newWaybill).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to insert waybill",
			"reason":  err.Error(),
		})
	}

	// -----------------------------------------------------------
	// 8. INSERT MERGED + VALIDATED ITEMS
	// -----------------------------------------------------------
	for i := range validatedItems {
		//validatedItems[i].TempID = newWaybill.TempID
		validatedItems[i].ErpItemID = GenerateAndSetNextERPproofIDGen()
		validatedItems[i].ErpItemID2 = GenerateAndSetNextERPproofIDGen()
	}

	if err := initializers.DB.Create(&validatedItems).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to insert waybill items",
			"reason":  err.Error(),
		})
	}

	// -----------------------------------------------------------
	// 9. RESPONSE
	// -----------------------------------------------------------
	return c.Status(201).JSON(deliveryproof.CreateDocumentdeliveryProofSuccessResponse{
		Success: true,
		Data: deliveryproof.CreateDocumentdeliveryProofResponse{
			TempERPProofId: newWaybill.TempID,
			OrderId:        newWaybill.OrderID,
			Message:        "Delivery proof created successfully",
		},
	})
}

func SendDocumentdeliveryProofErrorResponse(c *fiber.Ctx, message string) error {
	return c.Status(400).JSON(fiber.Map{
		"success": false,
		"data": fiber.Map{
			"TempERPProofId": "",
			"OrderId":        "",
			"Message":        message,
		},
	})
}

// GetDeliveryDocumentsProofHandler handles GET /spic_to_erp/customers/:coopId/deliverydocuments/invoices
// @Summary      Create deliverydocuments proof for a sales order within date range
// @Description  Create deliverydocuments proof for a sales order within date range
// @Tags         deliverydocuments proof
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        updatedFrom   query     string  false  " "
// @Param        updatedTo     query     string  false  " "
// @Param        page          query     int     false  "Page number"    default(1)
// @Param        perPage         query     int     false  "Items per page" default(10)
// @Success      200    {object}  deliveryproof.ListDeliveryDocumentsResponse
// @Router       /spic_to_erp/customers/{coopId}/deliverydocuments/invoices [get]
func GetDeliveryDocumentsProofHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")
	updatedFrom := c.Query("updatedFrom")
	updatedTo := c.Query("updatedTo")
	var deliverydocuments []deliveryproof.DocumentdeliveryProof
	emptydata := make([]deliveryproof.ListDeliveryDocumentsResponse, 0)

	if !isCoopAllowed(coopId) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Message": "The indicated cooperative does not exist.",
		})
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("perPage", "10"))
	if perPage <= 0 {
		perPage = 10
	}

	offset := (page - 1) * perPage
	var totalRecords int64

	query := initializers.DB.
		Model(&deliveryproof.WaybillItem{}).
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
		Limit(perPage).
		Offset(offset).
		Find(&deliverydocuments).Error; err != nil {

		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"Invoice": emptydata,
		})
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	data := make([]deliveryproof.DocumentdeliveryProof, 0)
	for _, f := range deliverydocuments {
		data = append(data, deliveryproof.DocumentdeliveryProof{
			ERPDeliveryDocumentId:   f.ERPDeliveryDocumentId,
			ERPDeliveryDocumentCode: f.ERPDeliveryDocumentCode,
		})
	}

	return c.Status(fiber.StatusOK).JSON(deliveryproof.ListDeliveryDocumentsResponse{
		Data: data,
		Pagination: deliveryproof.PaginationInfo{
			Page:        page,
			Limit:       perPage,
			TotalItems:  int(totalRecords),
			TotalPages:  totalPages,
			HasPrevious: page > 1,
			HasNext:     page < totalPages,
		},
	})
	// return c.Status(fiber.StatusCreated).JSON(response)
}

// GetDeliveryDocumentsProofParticularHandler handles GET /spic_to_erp/customers/:coopId/deliverydocuments/:deliveryNoteId/proof
// @Summary      Create deliverydocuments proof for a sales order
// @Description  Create deliverydocuments proof for a sales order
// @Tags         deliverydocuments proof
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        deliveryNoteId path      string  true   " "
// @Success      200    {object}  deliveryproof.InvoicesResponse
// @Router       /spic_to_erp/customers/{coopId}/deliverydocuments/{deliveryNoteId}/proof [get]
func GetDeliveryDocumentsProofParticularHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")
	orderID := c.Params("orderId")
	emptydata := make([]deliveryproof.InvoicesResponse, 0)

	// if orderID == "" {
	//  return c.Status(400).JSON(fiber.Map{
	//      "success": false,
	//      "message": "order_id is required",
	//  })
	// }
	if !isCoopAllowed(coopId) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Message": "The indicated cooperative does not exist.",
		})
	}
	var order deliveryproof.Waybill
	if err := initializers.DB.
		Where("order_id = ? AND coop_id = ?", orderID, coopId).
		First(&order).Error; err != nil {

		return c.Status(200).JSON(fiber.Map{
			"deliverynotes": emptydata,
		})
	}
	var orderItems []deliveryproof.WaybillItem
	if err := initializers.DB.
		Where("order_id = ?", orderID).
		Find(&orderItems).Error; err != nil {

		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch order items",
		})
	}
	var deliveryDocs []deliveryproof.Waybill
	if err := initializers.DB.
		Where("order_id = ? AND coop_id = ?", orderID, coopId).
		Find(&deliveryDocs).Error; err != nil {

		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch delivery documents",
		})
	}
	deliveryMap := make(map[string][]deliveryproof.Waybill)

	for _, doc := range deliveryDocs {
		deliveryMap[doc.DeliveryNoteID] =
			append(deliveryMap[doc.DeliveryNoteID], doc)
	}
	var response deliveryproof.InvoicesResponse

	for docCode, docs := range deliveryMap {

		note := deliveryproof.Invoice{
			ERPInvoiceId:   docs[0].DeliveryNoteID,
			ERPInvoiceCode: docCode,
			// ERPInvoiceDate: *docs[0].CreatedAt,
		}

		for _, d := range docs {
			for _, item := range orderItems {
				if item.OrderID == d.OrderID {

					note.Items = append(note.Items, deliveryproof.InvoiceItem{
						// ERPItemID:       item.ErpItemID2,
						StockKeepingUnit: item.StockKeepingUnit,
						Quantity:         item.Quantity,
						DeliveryNote: deliveryproof.InvoiceDeliveryNote{
							// TempERPDeliveryNoteId: order.TempID,
							ERPDeliveryDocumentId:   order.DeliveryNoteID,
							ERPDeliveryDocumentCode: order.DeliveryNoteDocument,
							ERPDeliveryDocumentDate: order.OrderID,
							Quantity:                item.Quantity,
							// ERPItemID:           item.ErpItemID,
							// OrderItemID:         item.OrderItemID,
						},
					})
				}
			}
		}

		response.Invoices = append(response.Invoices, note)
	}
	return c.Status(200).JSON(response)
}
