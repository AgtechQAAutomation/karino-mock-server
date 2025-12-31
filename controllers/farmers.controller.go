package controllers

import (
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/shyamsundaar/karino-mock-server/initializers"
	"github.com/shyamsundaar/karino-mock-server/models"
	// "github.com/shyamsundaar/karino-mock-server/query"
	// "gorm.io/gorm"
)

// CreateCustomerDetailHandler handles POST /spic_to_erp/customers/:coopId/farmers
// @Summary      Create a new farmer detail
// @Description  Create a new record in the details table
// @Tags         Details
// @Accept       json
// @Produce      json
// @Param        coopId  path      string                            true  "Cooperative ID"
// @Param        detail  body      models.CreateDetailSchema          true  "Create Detail Payload"
// @Success      201     {object}  models.CreateSuccessFarmerResponse
// @Router       /spic_to_erp/customers/{coopId}/farmers [post]
func CreateCustomerDetailHandler(c *fiber.Ctx) error {
	// 1. Get CoopID from URL Parameter
	coopId := c.Params("coopId")
	if coopId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "CoopID is required in URL"})
	}

	var payload *models.CreateDetailSchema

	// 2. Parse the JSON Body
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	// 3. Validate the Body
	// errors := models.ValidateStruct(payload)
	// if errors != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(errors)
	// }

	// 4. Map everything to the DB Model
	newDetail := models.FarmerDetails{
		CoopID:                      coopId, // Set from URL Param
		FarmerID:                    payload.FarmerID,
		FirstName:                   payload.FirstName,
		LastName:                    payload.LastName,
		MobileNumber:                payload.MobileNumber,
		RegionID:                    payload.RegionID,
		RegionPartID:                payload.RegionPartID,
		SettlementID:                payload.SettlementID,
		SettlementPartID:            payload.SettlementPartID,
		CustomGeographyStructure1ID: payload.CustomGeo1ID,
		CustomGeographyStructure2ID: payload.CustomGeo2ID,
		ZipCode:                     payload.ZipCode,
		FarmerKycTypeID:             payload.FarmerKycTypeID,
		FarmerKycType:               payload.FarmerKycType,
		FarmerKycID:                 payload.FarmerKycID,
		ClubID:                      payload.ClubID,
		ClubName:                    payload.ClubName,
		ClubLeaderFarmerID:          payload.ClubLeaderFarmerID,
	}

	// 5. Save to Database (GORM fills in CreatedAt/UpdatedAt here)
	result := initializers.DB.Create(&newDetail)
	if result.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": result.Error.Error()})
	}

	// Inside CreateCustomerDetailHandler...

	response := models.CreateSuccessFarmerResponse{
		Success: true,
		Data: models.FarmerResponse{
			TempERPCustomerID: newDetail.TempID,
			ErpCustomerId:     newDetail.CustomerID,
			ErpVendorId:       newDetail.VendorID,
			FarmerId:          newDetail.FarmerID,
			CreatedAt:         newDetail.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:         newDetail.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Message:           "Farmer detail created successfully",
		},
	}

	return c.Status(fiber.StatusCreated).JSON(response)

}

// FindDetails handles GET /spic_to_erp/customers/:coopId/farmers
// @Summary      List farmer details
// @Description  Get a paginated list of farmer details for a specific cooperative
// @Tags         Details
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        updatedFrom   query     string  false  " "
// @Param        updatedTo     query     string  false  " "
// @Param        page          query     int     false  "Page number"    default(1)
// @Param        limit         query     int     false  "Items per page" default(10)
// @Success      200    {object}  models.ListFarmersResponse
// @Router       /spic_to_erp/customers/{coopId}/farmers [get]
func FindCustomerDetailsHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var farmers []models.FarmerDetails
	var totalRecords int64

	query := initializers.DB.
		Model(&models.FarmerDetails{}).
		Where("coop_id = ?", coopId)

	query.Count(&totalRecords)

	if err := query.
		Limit(limit).
		Offset(offset).
		Find(&farmers).Error; err != nil {

		return c.Status(fiber.StatusBadGateway).JSON(models.ErrorFarmerResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))

	// ✅ Map DB → RESPONSE MODEL
	var data []models.FarmerResponse
	for _, f := range farmers {
		data = append(data, models.FarmerResponse{
			ErpCustomerId:     f.CustomerID,
			TempERPCustomerID: f.TempID,
			ErpVendorId:       f.VendorID,
			// TempErpVendorId:   f.TempVendorID, // if exists
			FarmerId:          f.FarmerID,
			CreatedAt:         f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:         f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.ListFarmersResponse{
		Data: data,
		Pagination: models.PaginationInfo{
			Page:        page,
			Limit:       limit,
			TotalItems:  int(totalRecords),
			TotalPages:  totalPages,
			HasPrevious: page > 1,
			HasNext:     page < totalPages,
		},
	})
}

// CreateVendorDetailHandler handles POST /spic_to_erp/vendors/:coopId/farmers
// @Summary      Create a new farmer detail
// @Description  Create a new record in the details table
// @Tags         Details
// @Accept       json
// @Produce      json
// @Param        coopId  path      string                            true  "Cooperative ID"
// @Param        detail  body      models.CreateDetailSchema          true  "Create Detail Payload"
// @Success      201     {object}  models.CreateSuccessFarmerResponse
// @Router       /spic_to_erp/vendors/{coopId}/farmers [post]
func CreateVendorDetailHandler(c *fiber.Ctx) error {
	// 1. Get CoopID from URL Parameter
	coopId := c.Params("coopId")
	if coopId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "CoopID is required in URL"})
	}

	var payload *models.CreateDetailSchema

	// 2. Parse the JSON Body
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	// 3. Validate the Body
	// errors := models.ValidateStruct(payload)
	// if errors != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(errors)
	// }

	// 4. Map everything to the DB Model
	newDetail := models.FarmerDetails{
		CoopID:                      coopId, // Set from URL Param
		FarmerID:                    payload.FarmerID,
		FirstName:                   payload.FirstName,
		LastName:                    payload.LastName,
		MobileNumber:                payload.MobileNumber,
		RegionID:                    payload.RegionID,
		RegionPartID:                payload.RegionPartID,
		SettlementID:                payload.SettlementID,
		SettlementPartID:            payload.SettlementPartID,
		CustomGeographyStructure1ID: payload.CustomGeo1ID,
		CustomGeographyStructure2ID: payload.CustomGeo2ID,
		ZipCode:                     payload.ZipCode,
		FarmerKycTypeID:             payload.FarmerKycTypeID,
		FarmerKycType:               payload.FarmerKycType,
		FarmerKycID:                 payload.FarmerKycID,
		ClubID:                      payload.ClubID,
		ClubName:                    payload.ClubName,
		ClubLeaderFarmerID:          payload.ClubLeaderFarmerID,
	}

	// 5. Save to Database (GORM fills in CreatedAt/UpdatedAt here)
	result := initializers.DB.Create(&newDetail)
	if result.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": result.Error.Error()})
	}

	// Inside CreateCustomerDetailHandler...

	response := models.CreateSuccessFarmerResponse{
		Success: true,
		Data: models.FarmerResponse{
			TempERPCustomerID: newDetail.TempID,
			ErpCustomerId:     newDetail.CustomerID,
			ErpVendorId:       newDetail.VendorID,
			FarmerId:          newDetail.FarmerID,
			CreatedAt:         newDetail.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:         newDetail.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Message:           "Farmer detail created successfully",
		},
	}

	return c.Status(fiber.StatusCreated).JSON(response)

}

// FindDetails handles GET /spic_to_erp/vendors/:coopId/farmers
// @Summary      List farmer details
// @Description  Get a paginated list of farmer details for a specific cooperative
// @Tags         Details
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        updatedFrom   query     string  false  " "
// @Param        updatedTo     query     string  false  " "
// @Param        page          query     int     false  "Page number"    default(1)
// @Param        limit         query     int     false  "Items per page" default(10)
// @Success      200    {object}  models.ListFarmersResponse
// @Router       /spic_to_erp/vendors/{coopId}/farmers [get]
func FindVendorDetailsHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var farmers []models.FarmerDetails
	var totalRecords int64

	query := initializers.DB.
		Model(&models.FarmerDetails{}).
		Where("coop_id = ?", coopId)

	query.Count(&totalRecords)

	if err := query.
		Limit(limit).
		Offset(offset).
		Find(&farmers).Error; err != nil {

		return c.Status(fiber.StatusBadGateway).JSON(models.ErrorFarmerResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))

	// ✅ Map DB → RESPONSE MODEL
	var data []models.FarmerResponse
	for _, f := range farmers {
		data = append(data, models.FarmerResponse{
			ErpCustomerId:     f.CustomerID,
			TempERPCustomerID: f.TempID,
			ErpVendorId:       f.VendorID,
			// TempErpVendorId:   f.TempVendorID, // if exists
			FarmerId:          f.FarmerID,
			CreatedAt:         f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:         f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.ListFarmersResponse{
		Data: data,
		Pagination: models.PaginationInfo{
			Page:        page,
			Limit:       limit,
			TotalItems:  int(totalRecords),
			TotalPages:  totalPages,
			HasPrevious: page > 1,
			HasNext:     page < totalPages,
		},
	})
}

// FindDetails handles GET /spic_to_erp/customers/:coopId/farmers/:farmerId
// @Summary      List farmer details
// @Description  Get a paginated list of farmer details for a specific cooperative
// @Tags         Details
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        farmerId path      string  true   " "
// @Success      200    {object}  models.FarmerDetailResponse
// @Router       /spic_to_erp/customers/{coopId}/farmers/{farmerId} [get]
func GetCustomerDetailHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")
	farmerId := c.Params("farmerId")

	var farmer models.FarmerDetails

	err := initializers.DB.
		Where("coop_id = ? AND farmer_id = ?", coopId, farmerId).
		First(&farmer).Error

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorFarmerResponse{
			Success: false,
			Message: "Farmer not found",
		})
	}

	response := models.FarmerDetailResponse{
		FarmerID:           farmer.FarmerID,
		Name:               farmer.FirstName + " " + farmer.LastName,
		MobileNumber:       farmer.MobileNumber,
		Cooperative:        farmer.CoopID,
		SettlementID:       farmer.SettlementID,
		SettlementPartID:   farmer.SettlementPartID,
		ZipCode:            farmer.ZipCode,
		FarmerKycTypeID:    farmer.FarmerKycTypeID,
		FarmerKycType:      farmer.FarmerKycType,
		FarmerKycID:        farmer.FarmerKycID,
		ClubID:             farmer.ClubID,
		ClubLeaderFarmerID: farmer.ClubLeaderFarmerID,
		Message:            "Farmer detail fetched successfully",
		EntityID:           farmer.TempID,        // or permanent entity ID
		CustomerCode:       farmer.CustomerID,
		VendorCode:         farmer.VendorID,
		CreatedDate:        farmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedDate:        farmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		// BankDetails: models.BankDetailsInfo{
		// 	IBAN:  farmer.IBAN,   // ensure field exists
		// 	SWIFT: farmer.SWIFT,  // ensure field exists
		// },
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// FindDetails handles GET /spic_to_erp/vendors/:coopId/farmers/:farmerId
// @Summary      List farmer details
// @Description  Get a paginated list of farmer details for a specific cooperative
// @Tags         Details
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        farmerId path      string  true   " "
// @Success      200    {object}  models.FarmerDetailResponse
// @Router       /spic_to_erp/vendors/{coopId}/farmers/{farmerId} [get]
func GetVendorDetailHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")
	farmerId := c.Params("farmerId")

	var farmer models.FarmerDetails

	err := initializers.DB.
		Where("coop_id = ? AND farmer_id = ?", coopId, farmerId).
		First(&farmer).Error

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorFarmerResponse{
			Success: false,
			Message: "Farmer not found",
		})
	}

	response := models.FarmerDetailResponse{
		FarmerID:           farmer.FarmerID,
		Name:               farmer.FirstName + " " + farmer.LastName,
		MobileNumber:       farmer.MobileNumber,
		Cooperative:        farmer.CoopID,
		SettlementID:       farmer.SettlementID,
		SettlementPartID:   farmer.SettlementPartID,
		ZipCode:            farmer.ZipCode,
		FarmerKycTypeID:    farmer.FarmerKycTypeID,
		FarmerKycType:      farmer.FarmerKycType,
		FarmerKycID:        farmer.FarmerKycID,
		ClubID:             farmer.ClubID,
		ClubLeaderFarmerID: farmer.ClubLeaderFarmerID,
		Message:            "Farmer detail fetched successfully",
		EntityID:           farmer.TempID,        // or permanent entity ID
		CustomerCode:       farmer.CustomerID,
		VendorCode:         farmer.VendorID,
		CreatedDate:        farmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedDate:        farmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		// BankDetails: models.BankDetailsInfo{
		// 	IBAN:  farmer.IBAN,   // ensure field exists
		// 	SWIFT: farmer.SWIFT,  // ensure field exists
		// },
	}

	return c.Status(fiber.StatusOK).JSON(response)
}