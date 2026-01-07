package sales

type SalesOrderListResponse struct {
    OrderID     string `json:"order_id"`
    OrderNumber string `json:"order_number"`
    FarmerID    string `json:"farmer_id"`
    FarmerName  string `json:"farmer_name"`
    ClubID      string `json:"club_id"`
    ClubName    string `json:"club_name"`
    CreatedAt   string `json:"created_at"`
    UpdatedAt   string `json:"updated_at"`
}

type ListSalesOrderResponse struct {
    Data       []SalesOrderListResponse `json:"data"`
    Pagination PaginationInfo           `json:"pagination"`
}

// PaginationInfo matches the required pagination format
type PaginationInfo struct {
    Page         int  `json:"page"`
    Limit        int  `json:"limit"`
    TotalItems   int  `json:"total_items"`
    TotalPages   int  `json:"total_pages"`
    HasPrevious  bool `json:"has_previous"`
    HasNext      bool `json:"has_next"`
}