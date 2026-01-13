package controllers

import (
	"github.com/gofiber/fiber/v2"
	// "github.com/shyamsundaar/karino-mock-server/models/deliveryproof"
)

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
func CreateDeliveryDocumentsProofHandler(c *fiber.Ctx) error{
	return c.Status(200).JSON("response")
}

// CreateDeliveryDocumentsProofHandler handles GET /spic_to_erp/customers/:coopId/deliverydocuments/invoices
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
func GetDeliveryDocumentsProofHandler(c *fiber.Ctx) error{
	return c.Status(200).JSON("response")
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
func GetDeliveryDocumentsProofParticularHandler(c *fiber.Ctx) error{
	return c.Status(200).JSON("response")
}