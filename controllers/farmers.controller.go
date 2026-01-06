package controllers

import (
	"context"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shyamsundaar/karino-mock-server/initializers"
	"github.com/shyamsundaar/karino-mock-server/models"

	// "karino-mock-server/query"
	"github.com/shyamsundaar/karino-mock-server/query"
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

func GenerateAndSetNextCustomerIDGen(
	ctx context.Context,
	q *query.Query,
	detailID uint,
) (string, error) {

	fd := q.FarmerDetails.WithContext(ctx)

	// 1. Get current row
	row, err := fd.Where(q.FarmerDetails.ID.Eq(detailID)).First()
	if err != nil {
		return "", err
	}

	if row.CustomerID != "" {
		return row.CustomerID, nil
	}

	// 2. Get last non-empty customer_id
	last, err := fd.
		Where(q.FarmerDetails.CustomerID.Neq("")).
		Order(q.FarmerDetails.ID.Desc()).
		First()

	next := 1
	if err == nil && last.CustomerID != "" {
		re := regexp.MustCompile(`\d+$`)
		if m := re.FindString(last.CustomerID); m != "" {
			n, _ := strconv.Atoi(m)
			next = n + 1
		}
	}

	newCustomerID := fmt.Sprintf("CUST%05d", next)

	// 3. Business delay
	time.Sleep(5 * time.Second)

	// 4. Update only if still empty
	_, err = fd.
		Where(
			q.FarmerDetails.ID.Eq(detailID),
			q.FarmerDetails.CustomerID.Eq(""),
		).
		UpdateColumnSimple(
			q.FarmerDetails.CustomerID.Value(newCustomerID),
			q.FarmerDetails.CustIDUpdateAt.Value(time.Now()),
		)

	if err != nil {
		return "", err
	}

	return newCustomerID, nil
}

func GenerateAndSetNextVendorIDGen(
	ctx context.Context,
	q *query.Query,
	detailID uint,
) (string, error) {

	fd := q.FarmerDetails.WithContext(ctx)

	// 1. Get current row
	row, err := fd.Where(q.FarmerDetails.ID.Eq(detailID)).First()
	if err != nil {
		return "", err
	}

	if row.VendorID != "" {
		return row.VendorID, nil
	}

	// 2. Get last non-empty vendor_id
	last, err := fd.
		Where(q.FarmerDetails.VendorID.Neq("")).
		Order(q.FarmerDetails.ID.Desc()).
		First()

	next := 1
	if err == nil && last.VendorID != "" {
		re := regexp.MustCompile(`\d+$`)
		if m := re.FindString(last.VendorID); m != "" {
			n, _ := strconv.Atoi(m)
			next = n + 1
		}
	}

	newVendorID := fmt.Sprintf("VEND%05d", next)

	// 3. Business delay
	time.Sleep(5 * time.Second)

	// 4. Update only if still empty
	_, err = fd.
		Where(
			q.FarmerDetails.ID.Eq(detailID),
			q.FarmerDetails.VendorID.Eq(""),
		).
		UpdateColumnSimple(
			q.FarmerDetails.VendorID.Value(newVendorID),
			q.FarmerDetails.VendorIDUpdateAt.Value(time.Now()),
		)

	if err != nil {
		return "", err
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
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}
	err := initializers.DB.Where("farmer_id = ? AND coop_id = ? AND (customer_id IS NULL OR customer_id = '')",
		payload.FarmerID,
		coopId,
	).First(&existingFarmer).Error

	if err == nil {
		// If customer_id is empty → generate & update
		if existingFarmer.CustomerID == "" {
			ctx := context.Background()
			q := query.Use(initializers.DB)

			go func(id uint) {
				if _, err := GenerateAndSetNextCustomerIDGen(ctx, q, id); err != nil {
					log.Println("❌ Customer ID generation failed:", err)
				}
			}(existingFarmer.ID)
		}
		response := models.CreateSuccessFarmerResponse{
			Success: true,
			Data: models.FarmerResponse{
				TempERPCustomerID: existingFarmer.TempID,
				ErpCustomerId:     existingFarmer.CustomerID,
				ErpVendorId:       existingFarmer.VendorID,
				FarmerId:          existingFarmer.FarmerID,
				CreatedAt:         existingFarmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:         existingFarmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
				Message:           "Farmer detail created successfully",
			},
		}

		return c.Status(fiber.StatusCreated).JSON(response)

	}
	// 2. Parse the JSON Body

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
	ctx := context.Background()
	q := query.Use(initializers.DB)

	go func(id uint) {
		var err error

		_, err = GenerateAndSetNextCustomerIDGen(ctx, q, id)
		if err != nil {
			log.Println("❌ Vendor ID gen failed:", err)
		}
	}(newDetail.ID)

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
	updatedFrom := c.Query("updatedFrom")
	updatedTo := c.Query("updatedTo")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	var farmers []models.FarmerDetails
	var totalRecords int64

	query := initializers.DB.
		Model(&models.FarmerDetails{}).
		Where("coop_id = ? AND AND (customer_id IS NULL OR customer_id = '')", coopId)

	if updatedFrom != "" {
		fromTime, err := time.Parse(time.RFC3339, updatedFrom)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid updatedFrom format. Use ISO8601 (YYYY-MM-DDTHH:MM:SSZ)"})
		}
		query = query.Where("updated_at >= ?", fromTime)
	}

	if updatedTo != "" {
		toTime, err := time.Parse(time.RFC3339, updatedTo)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid updatedTo format. Use ISO8601 (YYYY-MM-DDTHH:MM:SSZ)"})
		}
		query = query.Where("updated_at <= ?", toTime)
	}
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
	data := make([]models.FarmerResponse, 0)
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

	err := initializers.DB.Where("farmer_id = ? AND coop_id = ? AND (vendor_id IS NULL OR vendor_id = '')",
		payload.FarmerID,
		coopId,
	).First(&existingFarmer).Error

	if err == nil {
		// If customer_id is empty → generate & update
		if existingFarmer.VendorID == "" {
			ctx := context.Background()
			q := query.Use(initializers.DB)

			go func(id uint) {
				if _, err := GenerateAndSetNextVendorIDGen(ctx, q, id); err != nil {
					log.Println("❌ Vendor ID generation failed:", err)
				}
			}(existingFarmer.ID)
		}
		response := models.CreateSuccessFarmerResponse{
			Success: true,
			Data: models.FarmerResponse{
				TempERPCustomerID: existingFarmer.TempID,
				ErpCustomerId:     existingFarmer.CustomerID,
				ErpVendorId:       existingFarmer.VendorID,
				FarmerId:          existingFarmer.FarmerID,
				CreatedAt:         existingFarmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:         existingFarmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
				Message:           "Farmer detail created successfully",
			},
		}

		return c.Status(fiber.StatusCreated).JSON(response)

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
	ctx := context.Background()
	q := query.Use(initializers.DB)

	go func(id uint) {
		var err error

		_, err = GenerateAndSetNextVendorIDGen(ctx, q, id)
		if err != nil {
			log.Println("❌ Vendor ID gen failed:", err)
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
	updatedFrom := c.Query("updatedFrom")
	updatedTo := c.Query("updatedTo")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	var farmers []models.FarmerDetails
	var totalRecords int64

	query := initializers.DB.
		Model(&models.FarmerDetails{}).
		Where("coop_id = ? AND AND (vendor_id IS NULL OR vendor_id = '')", coopId)

	if updatedFrom != "" {
		fromTime, err := time.Parse(time.RFC3339, updatedFrom)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid updatedFrom format. Use ISO8601 (YYYY-MM-DDTHH:MM:SSZ)"})
		}
		query = query.Where("updated_at >= ?", fromTime)
	}

	if updatedTo != "" {
		toTime, err := time.Parse(time.RFC3339, updatedTo)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid updatedTo format. Use ISO8601 (YYYY-MM-DDTHH:MM:SSZ)"})
		}
		query = query.Where("updated_at <= ?", toTime)
	}

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
	data := make([]models.FarmerResponse, 0)
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
