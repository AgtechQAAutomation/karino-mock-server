package controllers

import (
	"math"
	"strconv"
	"strings"
	"fmt"
	"regexp"
	"time"
	"database/sql"
	"log"


	"github.com/gofiber/fiber/v2"
	"github.com/shyamsundaar/karino-mock-server/initializers"
	"github.com/shyamsundaar/karino-mock-server/models"
	// "github.com/shyamsundaar/karino-mock-server/query"
	// "gorm.io/gorm"
)

func isCoopAllowed(coopId string) bool {
	// 1. Get the string from your loaded config
	// (Assuming initializers.Config is your global config var)
	rawList := initializers.AppConfig.AllowedCooperatives

	// 2. Split it into a slice
	allowed := strings.Split(rawList, ",")

	// 3. Check if the ID exists in the slice
	for _, id := range allowed {
		if strings.TrimSpace(id) == coopId {
			return true
		}
	}
	return false
}

func GenerateAndSetNextCustomerID(db *sql.DB, detailID uint) (string, error) {
	// 1. Check if customer_id already exists
	var customerID sql.NullString
	err := db.QueryRow(`
		SELECT customer_id
		FROM farmer_details
		WHERE id = ?
	`, detailID).Scan(&customerID)

	if err != nil {
		return "", err
	}

	if customerID.Valid && customerID.String != "" {
		return customerID.String, nil
	}

	// 2. Get last NON-NULL customer_id
	var lastCustomerID sql.NullString
	_ = db.QueryRow(`
		SELECT customer_id
		FROM farmer_details
		WHERE customer_id IS NOT NULL AND customer_id != ''
		ORDER BY id DESC
		LIMIT 1
	`).Scan(&lastCustomerID)

	nextNumber := 1
	if lastCustomerID.Valid {
		re := regexp.MustCompile(`\d+$`)
		match := re.FindString(lastCustomerID.String)
		if match != "" {
			n, _ := strconv.Atoi(match)
			nextNumber = n + 1
		}
	}

	newCustomerID := fmt.Sprintf("CUST-%05d", nextNumber)

	// 3. Business delay
	time.Sleep(5 * time.Second)

	// 4. Update ONLY if still empty
	result, err := db.Exec(`
		UPDATE farmer_details
		SET customer_id = ?
		WHERE id = ? AND (customer_id IS NULL OR customer_id = '')
	`, newCustomerID, detailID)

	if err != nil {
		return "", err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Println("❌ RowsAffected error:", err)
	} else {
		log.Println("✅ Customer ID update rows affected:", rows)
	}

	return newCustomerID, nil
}


func GenerateAndSetNextVendorID(db *sql.DB, detailID uint) (string, error) {
	// 1. Check if vendor_id already exists
	var vendorID sql.NullString
	err := db.QueryRow(`
		SELECT vendor_id
		FROM farmer_details
		WHERE id = ?
	`, detailID).Scan(&vendorID)

	if err != nil {
		return "", err
	}

	if vendorID.Valid && vendorID.String != "" {
		return vendorID.String, nil
	}

	// 2. Get last NON-NULL vendor_id
	var lastVendorID sql.NullString
	_ = db.QueryRow(`
		SELECT vendor_id
		FROM farmer_details
		WHERE vendor_id IS NOT NULL AND vendor_id != ''
		ORDER BY id DESC
		LIMIT 1
	`).Scan(&lastVendorID)

	nextNumber := 1
	if lastVendorID.Valid {
		re := regexp.MustCompile(`\d+$`)
		match := re.FindString(lastVendorID.String)
		if match != "" {
			n, _ := strconv.Atoi(match)
			nextNumber = n + 1
		}
	}

	newVendorID := fmt.Sprintf("VEND%05d", nextNumber)

	// 3. Business delay
	time.Sleep(5 * time.Second)

	// 4. Update ONLY if still empty
	result, err := db.Exec(`
		UPDATE farmer_details
		SET vendor_id = ?
		WHERE id = ? AND (vendor_id IS NULL OR vendor_id = '')
	`, newVendorID, detailID)

	if err != nil {
		return "", err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Println("❌ Vendor RowsAffected error:", err)
	} else {
		log.Println("✅ Vendor ID update rows affected:", rows)
	}

	return newVendorID, nil
}

// CreateCustomerDetailHandler handles POST /spic_to_erp/customers/:coopId/farmers
// @Summary      Create a new farmer detail
// @Description  Create a new record in the details table
// @Tags         details
// @Accept       json
// @Produce      json
// @Param        coopId  path      string                            true  "Cooperative ID"
// @Param        detail  body      models.CreateDetailSchema          true  "Create Detail Payload"
// @Success      201     {object}  models.CreateSuccessFarmerResponse
// @Router       /spic_to_erp/customers/{coopId}/farmers [post]
func CreateCustomerDetailHandler(c *fiber.Ctx) error {
	// 1. Get CoopID from URL Parameter
	coopId := c.Params("coopId")
	var payload *models.CreateDetailSchema
	var existingFarmer models.FarmerDetails

	// 2. Parse the JSON Body
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	//3. Constraints for the payload check
	if payload.FarmerID == "" {
		return SendCustomerErrorResponse(c, "You must provide a Farmer ID.", payload.FarmerID)
	}

	if payload.FirstName == "" || payload.LastName == "" {
		return SendCustomerErrorResponse(c, "You must provide the first and last name.", payload.FarmerID)
	}

	if payload.FarmerKycID == "" && payload.ClubLeaderFarmerID == "" {
		return SendCustomerErrorResponse(c, "Either farmer_kyc_id or clubLeaderFarmerId must be provided.", payload.FarmerID)
	}

	kycId := initializers.DB.Where("farmer_kyc_id = ?", payload.FarmerKycID).First(&existingFarmer).Error
	if kycId == nil {
		return SendCustomerErrorResponse(c, "Farmer with the given KYC ID "+payload.FarmerKycID+" already exists.", payload.FarmerID)
	}

	if !isCoopAllowed(coopId) {
		return SendCustomerErrorResponse(c, "The indicated cooperative does not exist.", payload.FarmerID)
	}

	farmerId := initializers.DB.Where("farmer_id = ? AND coop_id = ?", payload.FarmerID, coopId).First(&existingFarmer).Error
	if farmerId == nil {
		return SendCustomerErrorResponse(c, "The Farmer ID "+payload.FarmerID+" is already registered in the cooperative "+coopId+".", payload.FarmerID)
	}

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
	sqlDB, err := initializers.DB.DB()
	if err != nil {
		log.Println("❌ Failed to get sql.DB:", err)
	} else {
		go func(detailID uint) {
			if _, err := GenerateAndSetNextCustomerID(sqlDB, detailID); err != nil {
				log.Println("❌ Failed to generate customer ID:", err)
			}
		}(newDetail.ID)
	}

	if result.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": result.Error.Error()})
	}

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

func SendCustomerErrorResponse(c *fiber.Ctx, msg string, farmerId string) error {
	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"success": false,
		"data": fiber.Map{
			"tempERPCustomerId": "0",
			"erpCustomerId":     "",
			"farmerId":          farmerId,
			"createdAt":         now,
			"updatedAt":         now,
			"Message":           msg,
		},
	})
}

// FindDetails handles GET /spic_to_erp/customers/:coopId/farmers
// @Summary      List farmer details
// @Description  Get a paginated list of farmer details for a specific cooperative
// @Tags         details
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
			FarmerId:  f.FarmerID,
			CreatedAt: f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
// @Tags         details
// @Accept       json
// @Produce      json
// @Param        coopId  path      string                            true  "Cooperative ID"
// @Param        detail  body      models.CreateDetailSchema          true  "Create Detail Payload"
// @Success      201     {object}  models.CreateSuccessFarmerResponse
// @Router       /spic_to_erp/vendors/{coopId}/farmers [post]
func CreateVendorDetailHandler(c *fiber.Ctx) error {
	// 1. Get CoopID from URL Parameter
	coopId := c.Params("coopId")
	var payload *models.CreateDetailSchema
	var existingFarmer models.FarmerDetails

	// 2. Parse the JSON Body
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	//3. Constraints for the payload check
	if payload.FarmerID == "" {
		return SendCustomerErrorResponse(c, "You must provide a Farmer ID.", payload.FarmerID)
	}

	if payload.FirstName == "" || payload.LastName == "" {
		return SendCustomerErrorResponse(c, "You must provide the first and last name.", payload.FarmerID)
	}

	if payload.FarmerKycID == "" && payload.ClubLeaderFarmerID == "" {
		return SendCustomerErrorResponse(c, "Either farmer_kyc_id or clubLeaderFarmerId must be provided.", payload.FarmerID)
	}

	kycId := initializers.DB.Where("farmer_kyc_id = ?", payload.FarmerKycID).First(&existingFarmer).Error
	if kycId == nil {
		return SendCustomerErrorResponse(c, "Farmer with the given KYC ID "+payload.FarmerKycID+" already exists.", payload.FarmerID)
	}

	if coopId == "" {
		return SendCustomerErrorResponse(c, "The indicated cooperative does not exist.", payload.FarmerID)
	}

	farmerId := initializers.DB.Where("farmer_id = ? AND coop_id = ?", payload.FarmerID, coopId).First(&existingFarmer).Error
	if farmerId == nil {
		return SendCustomerErrorResponse(c, "The Farmer ID "+payload.FarmerID+" is already registered in the cooperative "+coopId+".", payload.FarmerID)
	}

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
	sqlDB, err := initializers.DB.DB()
	if err != nil {
		log.Println("❌ Failed to get sql.DB:", err)
		return nil
	}

	go func(detailID uint) {
		if _, err := GenerateAndSetNextVendorID(sqlDB, detailID); err != nil {
			log.Println("❌ Vendor ID generation failed:", err)
		}
	}(newDetail.ID)

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
// @Tags         details
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
			FarmerId:  f.FarmerID,
			CreatedAt: f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
// @Tags         details
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        farmerId path      string  true   " "
// @Success      200    {object}  models.FarmerDetailResponse
// @Router       /spic_to_erp/customers/{coopId}/farmers/{farmerId} [get]
func GetCustomerDetailHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")
	farmerId := c.Params("farmerId")

	if !isCoopAllowed(coopId) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Message": "The indicated cooperative does not exist."})
	}
	var farmer models.FarmerDetails
	err := initializers.DB.
		Where("coop_id = ? AND farmer_id = ?", coopId, farmerId).
		First(&farmer).Error

	if err != nil {
		response := models.FarmerDetailResponse{
			FarmerID:           farmerId,
			Name:               "",
			MobileNumber:       "",
			Cooperative:        coopId,
			SettlementID:       0,
			SettlementPartID:   0,
			ZipCode:            "",
			FarmerKycTypeID:    0,
			FarmerKycType:      "",
			FarmerKycID:        "",
			ClubID:             "",
			ClubLeaderFarmerID: "",
			Message:            "",
			EntityID:           "", // or permanent entity ID
			CustomerCode:       "",
			VendorCode:         "",
			CreatedDate:        "1900-01-01T00:00:00",
			UpdatedDate:        "1900-01-01T00:00:00",
			BankDetails: models.BankDetailsInfo{
				IBAN:  "", // ensure field exists
				SWIFT: "", // ensure field exists
			},
		}
		return c.Status(fiber.StatusOK).JSON(response)
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
		EntityID:           farmer.TempID, // or permanent entity ID
		CustomerCode:       farmer.CustomerID,
		VendorCode:         farmer.VendorID,
		CreatedDate:        farmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedDate:        farmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		BankDetails: models.BankDetailsInfo{
			IBAN:  "", // ensure field exists
			SWIFT: "", // ensure field exists
		},
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// FindDetails handles GET /spic_to_erp/vendors/:coopId/farmers/:farmerId
// @Summary      List farmer details
// @Description  Get a paginated list of farmer details for a specific cooperative
// @Tags         details
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        farmerId path      string  true   " "
// @Success      200    {object}  models.FarmerDetailResponse
// @Router       /spic_to_erp/vendors/{coopId}/farmers/{farmerId} [get]
func GetVendorDetailHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")
	farmerId := c.Params("farmerId")

	if !isCoopAllowed(coopId) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Message": "The indicated cooperative does not exist."})
	}
	var farmer models.FarmerDetails

	err := initializers.DB.
		Where("coop_id = ? AND farmer_id = ?", coopId, farmerId).
		First(&farmer).Error

	if err != nil {
		response := models.FarmerDetailResponse{
			FarmerID:           farmerId,
			Name:               "",
			MobileNumber:       "",
			Cooperative:        coopId,
			SettlementID:       0,
			SettlementPartID:   0,
			ZipCode:            "",
			FarmerKycTypeID:    0,
			FarmerKycType:      "",
			FarmerKycID:        "",
			ClubID:             "",
			ClubLeaderFarmerID: "",
			Message:            "",
			EntityID:           "", // or permanent entity ID
			CustomerCode:       "",
			VendorCode:         "",
			CreatedDate:        "1900-01-01T00:00:00",
			UpdatedDate:        "1900-01-01T00:00:00",
			BankDetails: models.BankDetailsInfo{
				IBAN:  "", // ensure field exists
				SWIFT: "", // ensure field exists
			},
		}
		return c.Status(fiber.StatusOK).JSON(response)
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
		EntityID:           farmer.TempID, // or permanent entity ID
		CustomerCode:       farmer.CustomerID,
		VendorCode:         farmer.VendorID,
		CreatedDate:        farmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedDate:        farmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		BankDetails: models.BankDetailsInfo{
			IBAN:  "", // ensure field exists
			SWIFT: "", // ensure field exists
		},
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
