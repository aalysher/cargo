package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"cargo/internal/auth"
	"cargo/internal/repo"
	"cargo/internal/service"
)

// authTelegram — обмен initData на JWT-сессию.
func (h *Handler) authTelegram(w http.ResponseWriter, r *http.Request) {
	var req authTelegramReq
	if !h.readJSON(w, r, &req) {
		return
	}
	cfg := h.svc.Config()

	tg, err := auth.ValidateInitData(req.InitData, cfg.BotToken, cfg.InitDataMaxAge)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid_init_data", "откройте приложение через бота")
		return
	}
	cl, err := h.svc.UpsertTelegramClient(r.Context(), tg)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	token, err := auth.NewClientToken(cfg.JWTSecret, cl.ID)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, authResp{Token: token, Profile: toProfile(cl)})
}

// authDev — dev-вход без Telegram (только APP_ENV=dev).
func (h *Handler) authDev(w http.ResponseWriter, r *http.Request) {
	var req devAuthReq
	if !h.readJSON(w, r, &req) {
		return
	}
	if req.TelegramID == 0 {
		writeErr(w, http.StatusBadRequest, "validation_error", "нужен telegram_id")
		return
	}
	tg := auth.TelegramUser{ID: req.TelegramID, FirstName: req.FirstName, Username: req.Username}

	cl, err := h.svc.UpsertTelegramClient(r.Context(), tg)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	token, err := auth.NewClientToken(h.svc.Config().JWTSecret, cl.ID)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, authResp{Token: token, Profile: toProfile(cl)})
}

// me — профиль клиента.
func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	id, _ := clientIDFrom(r.Context())
	cl, err := h.svc.GetProfile(r.Context(), id)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toProfile(cl))
}

// updatePhone — сохранить телефон.
func (h *Handler) updatePhone(w http.ResponseWriter, r *http.Request) {
	id, _ := clientIDFrom(r.Context())
	var req updatePhoneReq
	if !h.readJSON(w, r, &req) {
		return
	}
	if err := h.svc.UpdatePhone(r.Context(), id, req.Phone); err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// warehouseAddress — адрес склада + личный код.
func (h *Handler) warehouseAddress(w http.ResponseWriter, r *http.Request) {
	id, _ := clientIDFrom(r.Context())
	wa, cl, err := h.svc.GetWarehouseAddress(r.Context(), id)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, warehouseAddressResp{
		Contact:    wa.Contact,
		Address:    wa.Address,
		Phone:      wa.Phone,
		CityCN:     wa.CityCN,
		ClientCode: cl.ClientCode,
	})
}

// tariffs — активные тарифы.
func (h *Handler) tariffs(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListTariffs(r.Context())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	out := make([]tariffResp, 0, len(list))
	for _, t := range list {
		out = append(out, tariffResp{
			TransportType:   t.TransportType,
			PricePerKg:      t.PricePerKg,
			Currency:        t.Currency,
			MinChargeableKg: t.MinChargeableKg,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// estimate — расчёт стоимости.
func (h *Handler) estimate(w http.ResponseWriter, r *http.Request) {
	transport := r.URL.Query().Get("transport")
	weight, err := strconv.ParseFloat(r.URL.Query().Get("weight"), 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation_error", "weight должен быть числом")
		return
	}
	res, err := h.svc.Estimate(r.Context(), transport, weight)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// createOrder — новая заявка.
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := clientIDFrom(r.Context())
	var req createOrderReq
	if !h.readJSON(w, r, &req) {
		return
	}
	ord, err := h.svc.CreateOrder(r.Context(), id, service.CreateOrderInput{
		TransportType:    req.TransportType,
		Contents:         req.Contents,
		DeclaredWeightKg: req.DeclaredWeightKg,
		Tracks:           req.Tracks,
		Comment:          req.Comment,
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toOrderResp(ord, nil, nil))
}

// listOrders — список заказов клиента.
func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	id, _ := clientIDFrom(r.Context())
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "active"
	}
	cursor, _ := strconv.ParseInt(r.URL.Query().Get("cursor"), 10, 64)

	orders, next, err := h.svc.ListOrders(r.Context(), id, scope, cursor)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	items := make([]orderResp, 0, len(orders))
	for _, o := range orders {
		items = append(items, toOrderResp(o, nil, nil))
	}
	writeJSON(w, http.StatusOK, listResp{Items: items, NextCursor: next})
}

// getOrder — деталь заказа с таймлайном и треками.
func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := clientIDFrom(r.Context())
	orderID, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation_error", "некорректный id заказа")
		return
	}
	ord, tracks, events, err := h.svc.GetOrder(r.Context(), id, orderID)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toOrderResp(ord, tracks, events))
}

// cancelOrder — отмена заявки клиентом.
func (h *Handler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := clientIDFrom(r.Context())
	orderID, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation_error", "некорректный id заказа")
		return
	}
	ord, err := h.svc.CancelOrder(r.Context(), id, orderID)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toOrderResp(ord, nil, nil))
}

// pathID — числовой параметр из пути.
func pathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

// toProfile — маппинг клиента в DTO.
func toProfile(cl repo.Client) profileResp {
	return profileResp{
		ID:         cl.ID,
		TelegramID: cl.TelegramID,
		FirstName:  cl.FirstName,
		LastName:   cl.LastName,
		Username:   cl.Username,
		PhotoURL:   cl.PhotoURL,
		Phone:      cl.Phone,
		ClientCode: cl.ClientCode,
	}
}

// toOrderResp — маппинг заказа в DTO (треки/события опционально).
func toOrderResp(ord repo.Order, tracks []repo.OrderTrack, events []repo.OrderEvent) orderResp {
	resp := orderResp{
		ID:                 ord.ID,
		OrderNumber:        ord.OrderNumber,
		Status:             ord.Status,
		TransportType:      ord.TransportType,
		Contents:           ord.Contents,
		DeclaredWeightKg:   ord.DeclaredWeightKg,
		ActualWeightKg:     ord.ActualWeightKg,
		ChargeableWeightKg: ord.ChargeableWeightKg,
		EstimatedCost:      ord.EstimatedCost,
		FinalCost:          ord.FinalCost,
		Currency:           ord.Currency,
		Comment:            ord.Comment,
		CreatedAt:          ord.CreatedAt.Format(time.RFC3339),
	}
	if tracks != nil {
		resp.Tracks = make([]trackResp, 0, len(tracks))
		for _, t := range tracks {
			resp.Tracks = append(resp.Tracks, trackResp{ID: t.ID, TrackingNumber: t.TrackingNumber})
		}
	}
	if events != nil {
		resp.Events = make([]eventResp, 0, len(events))
		for _, e := range events {
			resp.Events = append(resp.Events, eventResp{
				Type:      e.EventType,
				Status:    e.Status,
				Title:     e.Title,
				CreatedAt: e.CreatedAt.Format(time.RFC3339),
			})
		}
	}
	return resp
}
