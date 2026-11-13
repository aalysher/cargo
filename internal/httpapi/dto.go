package httpapi

// Request/response структуры клиентского API.

type authTelegramReq struct {
	InitData string `json:"init_data"`
}

type devAuthReq struct {
	TelegramID int64  `json:"telegram_id"`
	FirstName  string `json:"first_name"`
	Username   string `json:"username"`
}

type authResp struct {
	Token   string      `json:"token"`
	Profile profileResp `json:"profile"`
}

type profileResp struct {
	ID         int64  `json:"id"`
	TelegramID int64  `json:"telegram_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Username   string `json:"username"`
	PhotoURL   string `json:"photo_url"`
	Phone      string `json:"phone"`
	ClientCode string `json:"client_code"`
}

type updatePhoneReq struct {
	Phone string `json:"phone"`
}

type warehouseAddressResp struct {
	Contact    string `json:"contact"`
	Address    string `json:"address"`
	Phone      string `json:"phone"`
	CityCN     string `json:"city_cn"`
	ClientCode string `json:"client_code"`
}

// telegramUpdate — входящее обновление Bot API (из webhook).
type telegramUpdate struct {
	UpdateID int64 `json:"update_id"`
	Message  *struct {
		MessageID int64 `json:"message_id"`
		Chat      struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		PaymentSuccessful *struct {
			TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`
			InvoicePayload          string `json:"invoice_payload"`
		} `json:"payment_successful"`
	} `json:"message"`
}

type tariffResp struct {
	TransportType   string  `json:"transport_type"`
	PricePerKg      float64 `json:"price_per_kg"`
	Currency        string  `json:"currency"`
	MinChargeableKg float64 `json:"min_chargeable_kg"`
}

type createOrderReq struct {
	TransportType    string   `json:"transport_type"`
	Contents         string   `json:"contents"`
	DeclaredWeightKg float64  `json:"declared_weight_kg"`
	Tracks           []string `json:"tracks"`
	Comment          string   `json:"comment"`
}

type orderResp struct {
	ID                 int64       `json:"id"`
	OrderNumber        string      `json:"order_number"`
	Status             string      `json:"status"`
	TransportType      string      `json:"transport_type"`
	Contents           string      `json:"contents"`
	DeclaredWeightKg   float64     `json:"declared_weight_kg"`
	ActualWeightKg     *float64    `json:"actual_weight_kg"`
	ChargeableWeightKg *float64    `json:"chargeable_weight_kg"`
	EstimatedCost      *float64    `json:"estimated_cost"`
	FinalCost          *float64    `json:"final_cost"`
	Currency           string      `json:"currency"`
	Comment            string      `json:"comment"`
	CreatedAt          string      `json:"created_at"`
	Tracks             []trackResp `json:"tracks,omitempty"`
	Events             []eventResp `json:"events,omitempty"`
}

type trackResp struct {
	ID             int64  `json:"id"`
	TrackingNumber string `json:"tracking_number"`
}

type eventResp struct {
	Type      string `json:"type"`
	Status    string `json:"status,omitempty"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

type listResp struct {
	Items      []orderResp `json:"items"`
	NextCursor int64       `json:"next_cursor"`
}

type errResp struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
}
