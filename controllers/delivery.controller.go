package controllers

import (
	"strconv"
	"math"
	// "time"
	"github.com/gofiber/fiber/v2"
	"github.com/shyamsundaar/karino-mock-server/models/delivery"
	// "gorm.io/gorm"
	// "github.com/gin-gonic/gin"
	"github.com/shyamsundaar/karino-mock-server/initializers"
	// "github.com/shyamsundaar/karino-mock-server/models/delivery"
)

// GetDeliveryDetailHandler handles GET /spic_to_erp/customers/:coopId/salesorders/deliverydocuments
// @Summary      List deliverydocuments updated within date ranges
// @Description  Get a paginated list of farmer details for a specific cooperative
// @Tags         deliverydocuments
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        updatedFrom   query     string  false  " "
// @Param        updatedTo     query     string  false  " "
// @Param        page          query     int     false  "Page number"    default(1)
// @Param        limit         query     int     false  "Items per page" default(10)
// @Success      200    {object}  delivery.ListDeliveryDocumentsResponse
// @Router       /spic_to_erp/customers/{coopId}/salesorders/deliverydocuments [get]
func GetDeliveryDetailWithinDateHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")
	var deliverydocuments []delivery.DeliverydocumentsListResponse

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit
	var totalRecords int64

	query := initializers.DB.
		Model(&delivery.DeliverydocumentsListResponse{}).
		Where("coop_id = ?", coopId)

	query.Count(&totalRecords)

	if err := query.
		Limit(limit).
		Offset(offset).
		Find(&deliverydocuments).Error; err != nil {

		return c.Status(fiber.StatusBadGateway).JSON(delivery.ErrorDeliverydocumentsResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))

	var data []delivery.DeliverydocumentsListResponse
	for _, f := range deliverydocuments {
		data = append(data, delivery.DeliverydocumentsListResponse{
			// CoopID:              f.CoopID,
			TempERPSalesOrderId: f.TempERPSalesOrderId,
			ErpSalesOrderId:     f.ErpSalesOrderId,
			ErpSalesOrderCode:   f.ErpSalesOrderCode,
			SpicSalesOrderId:    f.SpicSalesOrderId,
			// CreatedAt:           f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			// UpdatedAt:           f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.Status(fiber.StatusOK).JSON(delivery.ListDeliveryDocumentsResponse{
		Data: data,
		Pagination: delivery.PaginationInfo{
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
	return c.Status(fiber.StatusCreated).JSON("hi")
}