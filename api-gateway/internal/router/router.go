package router

import (
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mnntn/ecommerce-project/api-gateway/internal/config"
	"github.com/mnntn/ecommerce-project/api-gateway/internal/middleware"
)

func NewRouter(cfg *config.Config) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.CORS)
	// r.Use(middleware.AuthMiddleware) // TODO: Включить после реализации аутентификации на фронтенде

	// Прокси маршруты для Order Service
	r.HandleFunc("/orders", proxyHandler(cfg.OrderServiceURL)).Methods(http.MethodPost, http.MethodGet, http.MethodOptions)
	r.HandleFunc("/orders/{order_id}", proxyHandler(cfg.OrderServiceURL)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/orders/user/{user_id}", proxyHandler(cfg.OrderServiceURL)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/users/{user_id}/orders", proxyHandler(cfg.OrderServiceURL)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/products", proxyHandler(cfg.OrderServiceURL)).Methods(http.MethodGet)
	r.HandleFunc("/api/orders", proxyHandler(cfg.OrderServiceURL)).Methods(http.MethodPost, http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/orders/{order_id}", proxyHandler(cfg.OrderServiceURL)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/orders/user/{user_id}", proxyHandler(cfg.OrderServiceURL)).Methods(http.MethodGet, http.MethodOptions)

	// Прокси маршруты для Payment Service
	r.HandleFunc("/api/payment/accounts", proxyHandler(cfg.PaymentServiceURL)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/payment/accounts/{user_id}", proxyHandler(cfg.PaymentServiceURL)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/payment/accounts/{user_id}/deposit", proxyHandler(cfg.PaymentServiceURL)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/payment/accounts/{user_id}/withdraw", proxyHandler(cfg.PaymentServiceURL)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/payment/users", proxyHandler(cfg.PaymentServiceURL)).Methods(http.MethodPost, http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/payment/users/{user_id}", proxyHandler(cfg.PaymentServiceURL)).Methods(http.MethodGet, http.MethodOptions)

	return r
}

func proxyHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Map API Gateway paths to service paths
		if strings.HasPrefix(path, "/api/payment") {
			path = strings.TrimPrefix(path, "/api/payment")
		} else if path == "/api/products" {
			path = "/products"
		} else if strings.HasPrefix(path, "/api/orders") {
			path = strings.Replace(path, "/api/orders", "/orders", 1)
		}

		url := target + path
		if r.URL.RawQuery != "" {
			url += "?" + r.URL.RawQuery
		}

		req, err := http.NewRequest(r.Method, url, r.Body)
		if err != nil {
			http.Error(w, "failed to create request", http.StatusInternalServerError)
			return
		}
		for k, v := range r.Header {
			for _, vv := range v {
				req.Header.Add(k, vv)
			}
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "failed to proxy request: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Устанавливаем CORS-заголовки всегда
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")

		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
