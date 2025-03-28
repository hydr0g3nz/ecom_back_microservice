package dto

// ShippingDTO represents a shipping data transfer object
type ShippingDTO struct {
	ID                string   `json:"id"`
	OrderID           string   `json:"order_id"`
	Carrier           string   `json:"carrier"`
	TrackingNumber    string   `json:"tracking_number"`
	Status            string   `json:"status"`
	EstimatedDelivery *string  `json:"estimated_delivery,omitempty"`
	ShippedAt         *string  `json:"shipped_at,omitempty"`
	DeliveredAt       *string  `json:"delivered_at,omitempty"`
	ShippingMethod    string   `json:"shipping_method"`
	ShippingCost      float64  `json:"shipping_cost"`
	Notes             string   `json:"notes"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

// UpdateShippingInput represents the input for updating shipping information
type UpdateShippingInput struct {
	OrderID           string   `json:"order_id"`
	Carrier           string   `json:"carrier"`
	TrackingNumber    string   `json:"tracking_number"`
	Status            string   `json:"status"`
	EstimatedDelivery *string  `json:"estimated_delivery,omitempty"`
	ShippingMethod    string   `json:"shipping_method"`
	ShippingCost      float64  `json:"shipping_cost"`
	Notes             string   `json:"notes"`
}

// UpdateTrackingInput represents the input for updating tracking information
type UpdateTrackingInput struct {
	ID             string `json:"id"`
	Carrier        string `json:"carrier"`
	TrackingNumber string `json:"tracking_number"`
}

// ShippingStatusUpdateInput represents the input for updating a shipping status
type ShippingStatusUpdateInput struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
