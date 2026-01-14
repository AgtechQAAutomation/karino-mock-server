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
	models "github.com/shyamsundaar/karino-mock-server/models/farmers"

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

	// If already assigned, return existing ID
	if row.CustomerID != "" {
		return row.CustomerID, nil
	}

	// 2. Get last non-empty customer_id
	last, err := fd.
		Where(q.FarmerDetails.CustomerID.Neq("")).
		Order(q.FarmerDetails.ID.Desc()).
		First()

	// Start from C26000 → numeric = 0
	next := 1

	if err == nil && last.CustomerID != "" {
		// Extract only the last 5 digits
		re := regexp.MustCompile(`(\d{5})$`)
		if m := re.FindString(last.CustomerID); m != "" {
			n, _ := strconv.Atoi(m)
			next = n + 1
		}
	}

	// Generate ID → C26 + 5-digit counter
	newCustomerID := fmt.Sprintf("C26%05d", next)

	// Optional business delay
	time.Sleep(time.Duration(initializers.AppConfig.TimeSeconds) * time.Second)

	// Update only if still empty (safe update)
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
	row, err := fd.
		Where(q.FarmerDetails.ID.Eq(detailID)).
		First()
	if err != nil {
		return "", err
	}

	// If already assigned, return existing VendorID
	if row.VendorID != "" {
		return row.VendorID, nil
	}

	// 2. Get last non-empty vendor_id
	last, err := fd.
		Where(q.FarmerDetails.VendorID.Neq("")).
		Order(q.FarmerDetails.ID.Desc()).
		First()

	// Start counter
	next := 1

	if err == nil && last.VendorID != "" {
		// Extract last 5 digits only
		re := regexp.MustCompile(`(\d{5})$`)
		if m := re.FindString(last.VendorID); m != "" {
			n, _ := strconv.Atoi(m)
			next = n + 1
		}
	}

	// Generate Vendor ID → V26 + 5-digit number
	newVendorID := fmt.Sprintf("F26%05d", next)

	// Optional business delay
	time.Sleep(time.Duration(initializers.AppConfig.TimeSeconds) * time.Second)

	// Update only if still empty (race-safe)
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
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        coopId  path      string                            true  "Cooperative ID"
// @Param        detail  body      models.CreateDetailSchema          true  "Create Detail Payload"
// @Success      201     {object}  models.CreateSuccessFarmerCustomerResponse
// @Router       /spic_to_erp/customers/{coopId}/farmers [post]
func CreateCustomerDetailHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")

	var payload *models.CreateDetailSchema
	var existingFarmer models.FarmerDetails
	var globalFarmer models.FarmerDetails

	// ----------------------------------------------------
	// 1. Parse request body
	// ----------------------------------------------------
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	// ----------------------------------------------------
	// 2. Farmer exists in SAME coop but customer not created
	// ----------------------------------------------------
	err := initializers.DB.
		Where(
			"farmer_id = ? AND coop_id = ? AND (customer_id IS NULL OR customer_id = '')",
			payload.FarmerID,
			coopId,
		).
		First(&existingFarmer).
		Error

	if err == nil {
		if existingFarmer.CustomerID == "" {
			ctx := context.Background()
			q := query.Use(initializers.DB)

			go func(id uint) {
				if _, err := GenerateAndSetNextCustomerIDGen(ctx, q, id); err != nil {
					log.Println("❌ Customer ID generation failed:", err)
				}
			}(existingFarmer.ID)
		}

		return c.Status(fiber.StatusOK).JSON(
			models.CreateSuccessFarmerCustomerResponse{
				Success: true,
				Data: models.CreateFarmerCustomerResponse{
					TempERPCustomerID: existingFarmer.TempID,
					ErpCustomerId:     existingFarmer.CustomerID,
					// ErpVendorId:       existingFarmer.VendorID,
					FarmerId:          existingFarmer.FarmerID,
					CreatedAt:         existingFarmer.CreatedAt.Format(time.RFC3339),
					UpdatedAt:         existingFarmer.UpdatedAt.Format(time.RFC3339),
					Message:           "Farmer detail created successfully",
				},
			},
		)
	}

	// ----------------------------------------------------
	// 3. BASIC VALIDATIONS
	// ----------------------------------------------------
	if payload.FarmerID == "" {
		return SendCustomerErrorResponse(c, "You must provide a Farmer ID.", "")
	}

	if payload.FirstName == "" || payload.LastName == "" {
		return SendCustomerErrorResponse(c, "You must provide the first and last name.", payload.FarmerID)
	}

	if payload.FarmerKycID == "" && payload.ClubLeaderFarmerID == "" {
		return SendCustomerErrorResponse(
			c,
			"Either farmer_kyc_id or clubLeaderFarmerId must be provided.",
			payload.FarmerID,
		)
	}

	if !isCoopAllowed(coopId) {
		return SendCustomerErrorResponse(c, "The indicated cooperative does not exist.", payload.FarmerID)
	}

	// ----------------------------------------------------
	// 4. CHECK IF FARMER EXISTS GLOBALLY
	// ----------------------------------------------------
	farmerExistsGlobally := initializers.DB.
		Where("farmer_id = ?", payload.FarmerID).
		First(&globalFarmer).
		Error == nil

	// ----------------------------------------------------
	// 5. KYC UNIQUENESS → ONLY IF FARMER IS NEW
	// ----------------------------------------------------
	if !farmerExistsGlobally && payload.FarmerKycID != "" {
		var kycFarmer models.FarmerDetails

		err := initializers.DB.
			Where("farmer_kyc_id = ?", payload.FarmerKycID).
			First(&kycFarmer).
			Error

		if err == nil {
			return SendCustomerErrorResponse(
				c,
				"Farmer with the given KYC ID "+payload.FarmerKycID+" already exists.",
				payload.FarmerID,
			)
		}
	}

	// ----------------------------------------------------
	// 6. BLOCK SAME FARMER IN SAME COOP
	// ----------------------------------------------------
	err = initializers.DB.
		Where("farmer_id = ? AND coop_id = ?", payload.FarmerID, coopId).
		First(&existingFarmer).
		Error

	if err == nil {
		return SendCustomerErrorResponse(
			c,
			"The Farmer ID "+payload.FarmerID+" is already registered in the cooperative "+coopId+".",
			payload.FarmerID,
		)
	}

	// ----------------------------------------------------
	// 7. CREATE NEW FARMER RECORD
	// ----------------------------------------------------
	newDetail := models.FarmerDetails{
		CoopID: coopId,
	}

	// copy from payload
	newDetail.FarmerID = payload.FarmerID
	newDetail.FirstName = payload.FirstName
	newDetail.LastName = payload.LastName
	newDetail.MobileNumber = payload.MobileNumber
	newDetail.RegionID = payload.RegionID
	newDetail.RegionPartID = payload.RegionPartID
	newDetail.SettlementID = payload.SettlementID
	newDetail.SettlementPartID = payload.SettlementPartID
	newDetail.ZipCode = payload.ZipCode
	newDetail.FarmerKycTypeID = payload.FarmerKycTypeID
	newDetail.FarmerKycType = payload.FarmerKycType
	newDetail.FarmerKycID = payload.FarmerKycID
	newDetail.ClubID = payload.ClubID
	newDetail.ClubName = payload.ClubName
	newDetail.ClubLeaderFarmerID = payload.ClubLeaderFarmerID
	newDetail.RaithuCreatedDate = payload.RaithuCreatedDate
	newDetail.RaithuUpdatedAt = payload.RaithuUpdatedAt

	// // Perform the mapping
	// if err := mapper.Map(payload, &newDetail); err != nil {
	// 	log.Fatalf("failed to map payload to FarmerDetails: %v", err)
	// }

	newDetail.CustomGeographyStructure1ID = payload.CustomGeo1ID
	newDetail.CustomGeographyStructure2ID = payload.CustomGeo2ID

	if err := initializers.DB.Create(&newDetail).Error; err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// ----------------------------------------------------
	// 8. ASYNC CUSTOMER ID GENERATION
	// ----------------------------------------------------
	ctx := context.Background()
	q := query.Use(initializers.DB)

	go func(id uint) {
		if _, err := GenerateAndSetNextCustomerIDGen(ctx, q, id); err != nil {
			log.Println("❌ Customer ID generation failed:", err)
		}
	}(newDetail.ID)

	// ----------------------------------------------------
	// 9. RESPONSE
	// ----------------------------------------------------
	return c.Status(fiber.StatusOK).JSON(
		models.CreateSuccessFarmerCustomerResponse{
			Success: true,
			Data: models.CreateFarmerCustomerResponse{
				TempERPCustomerID: newDetail.TempID,
				ErpCustomerId:     newDetail.CustomerID,
				// ErpVendorId:       newDetail.VendorID,
				FarmerId:          newDetail.FarmerID,
				CreatedAt:         newDetail.CreatedAt.Format(time.RFC3339),
				UpdatedAt:         newDetail.UpdatedAt.Format(time.RFC3339),
				Message:           "Farmer detail created successfully",
			},
		},
	)
}

func SendCustomerErrorResponse(c *fiber.Ctx, msg string, farmerId string) error {
	now := time.Now().UTC()
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
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        updatedFrom   query     string  false  " "
// @Param        updatedTo     query     string  false  " "
// @Param        page          query     int     false  "Page number"    default(1)
// @Param        perPage         query     int     false  "Items per page" default(10)
// @Success      200    {object}  models.ListFarmersCustomersResponse
// @Router       /spic_to_erp/customers/{coopId}/farmers [get]
func FindCustomerDetailsHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")
	updatedFrom := c.Query("updatedFrom")
	updatedTo := c.Query("updatedTo")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("perPage", "10"))
	if perPage <= 0 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	var farmers []models.FarmerDetails
	var totalRecords int64

	query := initializers.DB.
		Model(&models.FarmerDetails{}).
		Where("coop_id = ? AND customer_id IS NOT NULL AND customer_id != '' ", coopId)
	
		if !isCoopAllowed(coopId){
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"Message": "The indicated cooperative does not exist",
			})
	}

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

		query = query.Where("cust_id_update_at>= ? AND cust_id_update_at<= ? ", fromTime, toTime)
	}
	query.Count(&totalRecords)

	if err := query.
		Limit(perPage).
		Offset(offset).
		Find(&farmers).Error; err != nil {

		return c.Status(fiber.StatusBadGateway).JSON(models.ErrorFarmerResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	// ✅ Map DB → RESPONSE MODEL
	data := make([]models.FarmerCustomerResponse, 0)
	for _, f := range farmers {
		data = append(data, models.FarmerCustomerResponse{
			ErpCustomerId:     f.CustomerID,
			TempERPCustomerID: f.TempID,
			// ErpVendorId:       f.VendorID,
			// TempErpVendorId:   f.TempVendorID, // if exists
			FarmerId:  f.FarmerID,
			CreatedAt: f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.ListFarmersCustomersResponse{
		Data: data,
		Pagination: models.PaginationInfo{
			Page:        page,
			Limit:       perPage,
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
// @Tags         vendors
// @Accept       json
// @Produce      json
// @Param        coopId  path      string                            true  "Cooperative ID"
// @Param        detail  body      models.CreateDetailSchema          true  "Create Detail Payload"
// @Success      201     {object}  models.CreateSuccessFarmerVendorResponse
// @Router       /spic_to_erp/vendors/{coopId}/farmers [post]
func CreateVendorDetailHandler(c *fiber.Ctx) error {
	// 1. Get CoopID from URL Parameter
	coopId := c.Params("coopId")

	var payload *models.CreateDetailSchema
	var existingFarmer models.FarmerDetails
	var globalFarmer models.FarmerDetails

	// 2. Parse the JSON Body
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	// ----------------------------------------------------
	// 3. Farmer exists in SAME coop but vendor not created
	// ----------------------------------------------------
	err := initializers.DB.
		Where(
			"farmer_id = ? AND coop_id = ? AND (vendor_id IS NULL OR vendor_id = '')",
			payload.FarmerID,
			coopId,
		).
		First(&existingFarmer).
		Error

	if err == nil {
		if existingFarmer.VendorID == "" {
			ctx := context.Background()
			q := query.Use(initializers.DB)

			go func(id uint) {
				if _, err := GenerateAndSetNextVendorIDGen(ctx, q, id); err != nil {
					log.Println("❌ Vendor ID generation failed:", err)
				}
			}(existingFarmer.ID)
		}

		return c.Status(fiber.StatusOK).JSON(
			models.CreateSuccessFarmerVendorResponse{
				Success: true,
				Data: models.CreateFarmerVendorResponse{
					TempERPCustomerID: existingFarmer.TempID,
					// ErpCustomerId:     existingFarmer.CustomerID,
					ErpVendorId:       existingFarmer.VendorID,
					FarmerId:          existingFarmer.FarmerID,
					CreatedAt:         existingFarmer.CreatedAt.Format(time.RFC3339),
					UpdatedAt:         existingFarmer.UpdatedAt.Format(time.RFC3339),
					Message:           "Farmer detail created successfully",
				},
			},
		)
	}

	// ----------------------------------------------------
	// 4. BASIC VALIDATIONS
	// ----------------------------------------------------
	if payload.FarmerID == "" {
		return SendVendorErrorResponse(c, "You must provide a Farmer ID.", "")
	}

	if payload.FirstName == "" || payload.LastName == "" {
		return SendVendorErrorResponse(
			c,
			"You must provide the first and last name.",
			payload.FarmerID,
		)
	}

	if payload.FarmerKycID == "" && payload.ClubLeaderFarmerID == "" {
		return SendVendorErrorResponse(
			c,
			"Either farmer_kyc_id or clubLeaderFarmerId must be provided.",
			payload.FarmerID,
		)
	}

	if !isCoopAllowed(coopId) {
		return SendVendorErrorResponse(c, "The indicated cooperative does not exist.", payload.FarmerID)
	}

	// ----------------------------------------------------
	// 5. CHECK IF FARMER EXISTS GLOBALLY (ANY COOP)
	// ----------------------------------------------------
	farmerExistsGlobally := initializers.DB.
		Where("farmer_id = ?", payload.FarmerID).
		First(&globalFarmer).
		Error == nil

	// ----------------------------------------------------
	// 6. KYC UNIQUENESS → ONLY IF FARMER IS NEW
	// ----------------------------------------------------
	if !farmerExistsGlobally && payload.FarmerKycID != "" {
		var kycFarmer models.FarmerDetails

		err := initializers.DB.
			Where("farmer_kyc_id = ?", payload.FarmerKycID).
			First(&kycFarmer).
			Error

		if err == nil {
			return SendVendorErrorResponse(
				c,
				"Farmer with the given KYC ID "+payload.FarmerKycID+" already exists.",
				payload.FarmerID,
			)
		}
	}

	// ----------------------------------------------------
	// 7. BLOCK SAME FARMER IN SAME COOP
	// ----------------------------------------------------
	err = initializers.DB.
		Where("farmer_id = ? AND coop_id = ?", payload.FarmerID, coopId).
		First(&existingFarmer).
		Error

	if err == nil {
		return SendVendorErrorResponse(
			c,
			"The Farmer ID "+payload.FarmerID+
				" is already registered in the cooperative "+coopId+".",
			payload.FarmerID,
		)
	}

	// ----------------------------------------------------
	// 8. CREATE NEW FARMER RECORD
	// ----------------------------------------------------
	newDetail := models.FarmerDetails{
		CoopID:                      coopId,
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
		RaithuCreatedDate:           payload.RaithuCreatedDate,
		RaithuUpdatedAt:             payload.RaithuUpdatedAt,
	}

	if err := initializers.DB.Create(&newDetail).Error; err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// ----------------------------------------------------
	// 9. ASYNC VENDOR ID GENERATION
	// ----------------------------------------------------
	ctx := context.Background()
	q := query.Use(initializers.DB)

	go func(id uint) {
		if _, err := GenerateAndSetNextVendorIDGen(ctx, q, id); err != nil {
			log.Println("❌ Vendor ID gen failed:", err)
		}
	}(newDetail.ID)

	// ----------------------------------------------------
	// 10. RESPONSE
	// ----------------------------------------------------
	return c.Status(fiber.StatusOK).JSON(
		models.CreateSuccessFarmerVendorResponse{
			Success: true,
			Data: models.CreateFarmerVendorResponse{
				TempERPCustomerID: newDetail.TempID,
				// ErpCustomerId:     newDetail.CustomerID,
				ErpVendorId:       newDetail.VendorID,
				FarmerId:          newDetail.FarmerID,
				CreatedAt:         newDetail.CreatedAt.Format(time.RFC3339),
				UpdatedAt:         newDetail.UpdatedAt.Format(time.RFC3339),
				Message:           "Farmer detail created successfully",
			},
		},
	)
}

func SendVendorErrorResponse(c *fiber.Ctx, msg string, farmerId string) error {
	now := time.Now().UTC()
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"success": false,
		"data": fiber.Map{
			"tempERPCustomerId": "0",
			"erpVendorId":     "",
			"farmerId":          farmerId,
			"createdAt":         now,
			"updatedAt":         now,
			"Message":           msg,
		},
	})
}

// FindDetails handles GET /spic_to_erp/vendors/:coopId/farmers
// @Summary      List farmer details
// @Description  Get a paginated list of farmer details for a specific cooperative
// @Tags         vendors
// @Accept       json
// @Produce      json
// @Param        coopId path      string  true   " "
// @Param        updatedFrom   query     string  false  " "
// @Param        updatedTo     query     string  false  " "
// @Param        page          query     int     false  "Page number"    default(1)
// @Param        perPage         query     int     false  "Items per page" default(10)
// @Success      200    {object}  models.ListFarmersVendorsResponse
// @Router       /spic_to_erp/vendors/{coopId}/farmers [get]
func FindVendorDetailsHandler(c *fiber.Ctx) error {
	coopId := c.Params("coopId")
	updatedFrom := c.Query("updatedFrom")
	updatedTo := c.Query("updatedTo")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("perPage", "10"))
	if perPage <= 0 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	var farmers []models.FarmerDetails
	var totalRecords int64

	query := initializers.DB.
		Model(&models.FarmerDetails{}).
		Where("coop_id = ? AND vendor_id IS NOT NULL AND vendor_id != ''", coopId)
	
	if !isCoopAllowed(coopId){
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"Message": "The indicated cooperative does not exist",
			})
	}

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

		query = query.Where("vendor_id_update_at>= ? AND vendor_id_update_at<= ? ", fromTime, toTime)
	}

	query.Count(&totalRecords)

	if err := query.
		Limit(perPage).
		Offset(offset).
		Find(&farmers).Error; err != nil {

		return c.Status(fiber.StatusBadGateway).JSON(models.ErrorFarmerResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	// ✅ Map DB → RESPONSE MODEL
	data := make([]models.FarmerVendorResponse, 0)
	for _, f := range farmers {
		data = append(data, models.FarmerVendorResponse{
			// ErpCustomerId:     f.CustomerID,
			TempERPCustomerID: f.TempID,
			ErpVendorId:       f.VendorID,
			// TempErpVendorId:   f.TempVendorID, // if exists
			FarmerId:  f.FarmerID,
			CreatedAt: f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.ListFarmersVendorsResponse{
		Data: data,
		Pagination: models.PaginationInfo{
			Page:        page,
			Limit:       perPage,
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
// @Tags         customers
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
		Where("coop_id = ? AND farmer_id = ? AND customer_id IS NOT NULL AND customer_id != '' ", coopId, farmerId).
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
// @Tags         vendors
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
		Where("coop_id = ? AND farmer_id = ? AND vendor_id IS NOT NULL AND vendor_id != '' ", coopId, farmerId).
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
