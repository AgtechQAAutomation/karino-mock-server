package controllers

import (
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shyamsundaar/karino-mock-server/initializers"

	// "github.com/shyamsundaar/karino-mock-server/models/delivery"
	"github.com/shyamsundaar/karino-mock-server/models/deliveryproof"

	// "context"
	"strconv"

	"github.com/google/uuid"
	// "github.com/shyamsundaar/karino-mock-server/query"
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
	coopId := c.Params("coopId")

	var payload *deliveryproof.CreateDocumentdeliveryProofSuccessResponse
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	if !isCoopAllowed(coopId) {
		return SendDocumentdeliveryProofErrorResponse(c, "The indicated cooperative does not exist.", payload.Data.OrderId)
	}

	return c.Status(fiber.StatusOK).JSON(
		deliveryproof.CreateDocumentdeliveryProofSuccessResponse{
			Success: true,
			Data: deliveryproof.CreateDocumentdeliveryProofResponse{
				TempERPProofId: GenerateAndSetNextERPproofIDGen(),
				OrderId:        payload.Data.OrderId,
				Message:        "Delivery Document proof created successfully",
			},
		},
	)
}

func SendDocumentdeliveryProofErrorResponse(c *fiber.Ctx, message string, orderId string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"success": false,
		"data": fiber.Map{
			"TempERPProofId": "",
			"OrderId":        orderId,
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
	return c.Status(200).JSON("response")
}
