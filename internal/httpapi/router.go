package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"cargo/internal/config"
	"cargo/internal/service"
)

// Handler — клиентские и auth-эндпоинты.
type Handler struct {
	svc *service.Service
}

// New собирает HTTP-роутер.
func New(svc *service.Service, cfg *config.Config) http.Handler {
	h := &Handler{svc: svc}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Use(cors.Handler(cors.Options{
		// TMA открывается из доменов Telegram; для MVP разрешаем все origin.
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/telegram", h.authTelegram)
		if cfg.Env == "dev" {
			r.Post("/auth/dev", h.authDev)
		}

		r.Group(func(r chi.Router) {
			r.Use(ClientAuth(cfg.JWTSecret))
			r.Get("/me", h.me)
			r.Patch("/me", h.updatePhone)
			r.Get("/warehouse-address", h.warehouseAddress)
			r.Get("/tariffs", h.tariffs)
			r.Get("/tariffs/estimate", h.estimate)
			r.Post("/orders", h.createOrder)
			r.Get("/orders", h.listOrders)
			r.Get("/orders/{id}", h.getOrder)
			r.Post("/orders/{id}/cancel", h.cancelOrder)
		})

		// --- Админ-API ---
		r.Post("/admin/auth/login", h.adminLogin)

		r.Group(func(r chi.Router) {
			r.Use(AdminAuth(cfg.JWTSecret))
			r.Get("/admin/orders", h.adminListOrders)
			r.Get("/admin/orders/{id}", h.adminGetOrder)
			r.Patch("/admin/orders/{id}", h.adminPatchOrder)
			r.Post("/admin/orders/{id}/status", h.adminChangeStatus)
			r.Post("/admin/orders/{id}/tracks", h.adminAddTrack)
			r.Delete("/admin/orders/{id}/tracks/{trackID}", h.adminRemoveTrack)
			r.Get("/admin/tariffs", h.adminListTariffs)
			r.Get("/admin/clients", h.adminListClients)
			r.Get("/admin/clients/{id}", h.adminGetClient)
			r.Get("/admin/settings", h.adminGetSettings)

			// Правки справочников — только роль admin.
			r.Group(func(r chi.Router) {
				r.Use(RequireAdminRole)
				r.Patch("/admin/tariffs/{id}", h.adminPatchTariff)
				r.Put("/admin/settings/{key}", h.adminPutSetting)
			})
		})
	})

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return r
}
