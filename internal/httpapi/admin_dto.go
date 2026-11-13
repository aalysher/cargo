package httpapi

// Request/response структуры админского API.

type adminLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type adminProfileResp struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type adminAuthResp struct {
	Token   string           `json:"token"`
	Profile adminProfileResp `json:"profile"`
}

type adminOrderResp struct {
	ID                 int64            `json:"id"`
	OrderNumber        string           `json:"order_number"`
	Status             string           `json:"status"`
	TransportType      string           `json:"transport_type"`
	Contents           string           `json:"contents"`
	DeclaredWeightKg   float64          `json:"declared_weight_kg"`
	ActualWeightKg     *float64         `json:"actual_weight_kg"`
	ChargeableWeightKg *float64         `json:"chargeable_weight_kg"`
	EstimatedCost      *float64         `json:"estimated_cost"`
	FinalCost          *float64         `json:"final_cost"`
	Currency           string           `json:"currency"`
	Comment            string           `json:"comment"`
	CreatedAt          string           `json:"created_at"`
	UpdatedAt          string           `json:"updated_at"`
	Client             adminClientBrief `json:"client"`
	Tracks             []trackResp      `json:"tracks,omitempty"`
	Events             []eventResp      `json:"events,omitempty"`
}

type adminClientBrief struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Phone     string `json:"phone"`
	Code      string `json:"code"`
}

type adminOrderPatchReq struct {
	ActualWeightKg *float64 `json:"actual_weight_kg,omitempty"`
	FinalCost      *float64 `json:"final_cost,omitempty"`
	Comment        *string  `json:"comment,omitempty"`
}

type adminStatusReq struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

type adminTrackReq struct {
	TrackingNumber string `json:"tracking_number"`
}

type adminTariffResp struct {
	ID              int64   `json:"id"`
	TransportType   string  `json:"transport_type"`
	PricePerKg      float64 `json:"price_per_kg"`
	Currency        string  `json:"currency"`
	MinChargeableKg float64 `json:"min_chargeable_kg"`
	IsActive        bool    `json:"is_active"`
}

type adminTariffPatchReq struct {
	PricePerKg    *float64 `json:"price_per_kg,omitempty"`
	MinChargeable *float64 `json:"min_chargeable_kg,omitempty"`
	IsActive      *bool    `json:"is_active,omitempty"`
}

type adminClientResp struct {
	ID          int64             `json:"id"`
	TelegramID  int64             `json:"telegram_id"`
	Username    string            `json:"username"`
	FirstName   string            `json:"first_name"`
	LastName    string            `json:"last_name"`
	Phone       string            `json:"phone"`
	Code        string            `json:"code"`
	OrdersCount int               `json:"orders_count"`
	Orders      []orderResp       `json:"orders,omitempty"`
}

type adminSettingsResp struct {
	Settings map[string]any `json:"settings"`
}

type adminSettingReq struct {
	Value map[string]any `json:"value"`
}

type adminListResp struct {
	Items      []adminOrderResp `json:"items"`
	NextCursor int64            `json:"next_cursor"`
}
