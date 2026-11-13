package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"cargo/internal/repo"
)

// adminLogin — POST /admin/auth/login (публичный).
func (h *Handler) adminLogin(w http.ResponseWriter, r *http.Request) {
	var req adminLoginReq
	if !h.readJSON(w, r, &req) {
		return
	}
	u, token, err := h.svc.AdminLogin(r.Context(), req.Username, req.Password)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, adminAuthResp{
		Token: token,
		Profile: adminProfileResp{
			ID:          u.ID,
			Username:    u.Username,
			DisplayName: u.DisplayName,
			Role:        u.Role,
		},
	})
}

// adminListOrders — GET /admin/orders?status=&q=&cursor=.
func (h *Handler) adminListOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("q")
	cursor, _ := strconv.ParseInt(r.URL.Query().Get("cursor"), 10, 64)

	orders, next, err := h.svc.ListAdminOrders(r.Context(), status, search, cursor)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	items := make([]adminOrderResp, 0, len(orders))
	for _, ao := range orders {
		items = append(items, toAdminOrderResp(ao, nil, nil))
	}
	writeJSON(w, http.StatusOK, adminListResp{Items: items, NextCursor: next})
}

// adminGetOrder — GET /admin/orders/{id}.
func (h *Handler) adminGetOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation_error", "некорректный id заказа")
		return
	}
	ao, tracks, events, err := h.svc.GetAdminOrder(r.Context(), id)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toAdminOrderResp(ao, tracks, events))
}

// adminPatchOrder — PATCH /admin/orders/{id} (вес / финальная стоимость / комментарий).
func (h *Handler) adminPatchOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation_error", "некорректный id заказа")
		return
	}
	var req adminOrderPatchReq
	if !h.readJSON(w, r, &req) {
		return
	}
	actorID := adminIDFrom(r.Context())

	if req.ActualWeightKg != nil {
		ao, err := h.svc.SetOrderWeight(r.Context(), id, *req.ActualWeightKg, actorID)
		if err != nil {
			h.fail(w, r, err)
			return
		}
		ao, tracks, events, err := h.adminOrderFull(r.Context(), ao)
		if err != nil {
			h.fail(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, toAdminOrderResp(ao, tracks, events))
		return
	}

	ao, err := h.svc.UpdateOrderMeta(r.Context(), id, req.FinalCost, req.Comment, actorID)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	ao, tracks, events, err := h.adminOrderFull(r.Context(), ao)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toAdminOrderResp(ao, tracks, events))
}

// adminOrderFull — дотягивает треки и таймлайн к изменённому заказу,
// чтобы клиент админки после мутации получал полную карточку.
func (h *Handler) adminOrderFull(ctx context.Context, ao repo.AdminOrder) (repo.AdminOrder, []repo.OrderTrack, []repo.OrderEvent, error) {
	tracks, err := h.svc.Orders.TracksByOrder(ctx, ao.ID)
	if err != nil {
		return repo.AdminOrder{}, nil, nil, err
	}
	events, err := h.svc.Events.ListForOrder(ctx, ao.ID)
	if err != nil {
		return repo.AdminOrder{}, nil, nil, err
	}
	return ao, tracks, events, nil
}

// adminChangeStatus — POST /admin/orders/{id}/status.
func (h *Handler) adminChangeStatus(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation_error", "некорректный id заказа")
		return
	}
	var req adminStatusReq
	if !h.readJSON(w, r, &req) {
		return
	}
	ao, err := h.svc.AdminChangeStatus(r.Context(), id, req.Status, req.Note, adminIDFrom(r.Context()))
	if err != nil {
		h.fail(w, r, err)
		return
	}
	ao, tracks, events, err := h.adminOrderFull(r.Context(), ao)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toAdminOrderResp(ao, tracks, events))
}

// adminAddTrack — POST /admin/orders/{id}/tracks.
func (h *Handler) adminAddTrack(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation_error", "некорректный id заказа")
		return
	}
	var req adminTrackReq
	if !h.readJSON(w, r, &req) {
		return
	}
	ao, err := h.svc.AdminAddTrack(r.Context(), id, req.TrackingNumber, adminIDFrom(r.Context()))
	if err != nil {
		h.fail(w, r, err)
		return
	}
	tracks, err := h.svc.Orders.TracksByOrder(r.Context(), ao.ID)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toAdminOrderResp(ao, tracks, nil))
}

// adminRemoveTrack — DELETE /admin/orders/{id}/tracks/{trackID}.
func (h *Handler) adminRemoveTrack(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation_error", "некорректный id заказа")
		return
	}
	trackID, err := strconv.ParseInt(chi.URLParam(r, "trackID"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation_error", "некорректный id трека")
		return
	}
	ao, err := h.svc.AdminRemoveTrack(r.Context(), id, trackID, adminIDFrom(r.Context()))
	if err != nil {
		h.fail(w, r, err)
		return
	}
	tracks, err := h.svc.Orders.TracksByOrder(r.Context(), ao.ID)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toAdminOrderResp(ao, tracks, nil))
}

// adminListTariffs — GET /admin/tariffs.
func (h *Handler) adminListTariffs(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListAdminTariffs(r.Context())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	out := make([]adminTariffResp, 0, len(list))
	for _, t := range list {
		out = append(out, toAdminTariffResp(t))
	}
	writeJSON(w, http.StatusOK, out)
}

// adminPatchTariff — PATCH /admin/tariffs/{id} (только admin).
func (h *Handler) adminPatchTariff(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation_error", "некорректный id тарифа")
		return
	}
	var req adminTariffPatchReq
	if !h.readJSON(w, r, &req) {
		return
	}
	tf, err := h.svc.UpdateTariff(r.Context(), id, req.PricePerKg, req.MinChargeable, req.IsActive)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toAdminTariffResp(tf))
}

// adminListClients — GET /admin/clients?q=.
func (h *Handler) adminListClients(w http.ResponseWriter, r *http.Request) {
	clients, err := h.svc.ListClientsAdmin(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		h.fail(w, r, err)
		return
	}
	out := make([]adminClientResp, 0, len(clients))
	for _, c := range clients {
		out = append(out, toAdminClientResp(c, nil))
	}
	writeJSON(w, http.StatusOK, out)
}

// adminGetClient — GET /admin/clients/{id}: клиент + последние заказы.
func (h *Handler) adminGetClient(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation_error", "некорректный id клиента")
		return
	}
	cl, orders, err := h.svc.GetClientAdmin(r.Context(), id)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	clientOrders := make([]orderResp, 0, len(orders))
	for _, o := range orders {
		clientOrders = append(clientOrders, toOrderResp(o, nil, nil))
	}
	writeJSON(w, http.StatusOK, toAdminClientResp(cl, clientOrders))
}

// adminGetSettings — GET /admin/settings.
func (h *Handler) adminGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.svc.GetSettingsAdmin(r.Context())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	out := map[string]any{}
	for k, raw := range settings {
		var v any
		if json.Unmarshal(raw, &v) == nil {
			out[k] = v
		}
	}
	writeJSON(w, http.StatusOK, adminSettingsResp{Settings: out})
}

// adminPutSetting — PUT /admin/settings/{key} (только admin).
func (h *Handler) adminPutSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var req adminSettingReq
	if !h.readJSON(w, r, &req) {
		return
	}
	raw, err := json.Marshal(req.Value)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation_error", "значение должно быть JSON-объектом")
		return
	}
	if err := h.svc.PutSettingAdmin(r.Context(), key, raw); err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// toAdminOrderResp — маппинг заказа+клиента в DTO.
func toAdminOrderResp(ao repo.AdminOrder, tracks []repo.OrderTrack, events []repo.OrderEvent) adminOrderResp {
	resp := adminOrderResp{
		ID:                 ao.ID,
		OrderNumber:        ao.OrderNumber,
		Status:             ao.Status,
		TransportType:      ao.TransportType,
		Contents:           ao.Contents,
		DeclaredWeightKg:   ao.DeclaredWeightKg,
		ActualWeightKg:     ao.ActualWeightKg,
		ChargeableWeightKg: ao.ChargeableWeightKg,
		EstimatedCost:      ao.EstimatedCost,
		FinalCost:          ao.FinalCost,
		Currency:           ao.Currency,
		Comment:            ao.Comment,
		CreatedAt:          ao.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          ao.UpdatedAt.Format(time.RFC3339),
		Client: adminClientBrief{
			ID:        ao.ClientID,
			FirstName: ao.ClientFirstName,
			LastName:  ao.ClientLastName,
			Username:  ao.ClientUsername,
			Phone:     ao.ClientPhone,
			Code:      ao.ClientCode,
		},
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

// toAdminTariffResp — маппинг тарифа в DTO.
func toAdminTariffResp(t repo.Tariff) adminTariffResp {
	return adminTariffResp{
		ID:              t.ID,
		TransportType:   t.TransportType,
		PricePerKg:      t.PricePerKg,
		Currency:        t.Currency,
		MinChargeableKg: t.MinChargeableKg,
		IsActive:        t.IsActive,
	}
}

// toAdminClientResp — маппинг клиента в DTO (заказы опционально).
func toAdminClientResp(cl repo.Client, orders []orderResp) adminClientResp {
	resp := adminClientResp{
		ID:         cl.ID,
		TelegramID: cl.TelegramID,
		Username:   cl.Username,
		FirstName:  cl.FirstName,
		LastName:   cl.LastName,
		Phone:      cl.Phone,
		Code:       cl.ClientCode,
		Orders:     orders,
	}
	if orders != nil {
		resp.OrdersCount = len(orders)
	}
	return resp
}
