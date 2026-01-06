package sales

// OrderItem represents each item inside order_items
type OrderItem struct {
	OrderItemID           string  `json:"order_item_id"`
	OrderItemNumber       string  `json:"order_item_number"`
	StockKeepingUnit      string  `json:"stock_keeping_unit"`
	ProductGroup          string  `json:"product_group"`
	InputItemID           string  `json:"input_item_id"`
	InputItemName         string  `json:"input_item_name"`
	InputItemNameCaption  string  `json:"input_item_name_caption"`
	Quantity              float64 `json:"quantity"`
	QuantityUnitKey       string  `json:"quantity_unit_key"`
	UnitPrice             float64 `json:"unit_price"`
	Price                 string  `json:"price"`
	PriceUnitKey          string  `json:"price_unit_key"`
	NumberOfUnits         int     `json:"number_of_units"`
}