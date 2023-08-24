package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"
	"log"
	"main/cmd/web/handlers"
	"main/internal/config"
	"main/internal/courier"
	"main/internal/order"
	"main/pkg"
	"net/http"
	"os"
)

var cfg *config.Config

func init() {
	cfg = LoadConfig()
}

func LoadConfig() *config.Config {
	confStream, err := os.ReadFile("./config/app.yaml")
	if err != nil {
		log.Fatalln("Error to open read config file:", err)
	}

	conf := config.NewConfig()
	err = yaml.Unmarshal(confStream, conf)
	if err != nil {
		log.Fatalln("Error to unmarshal data from config file:", err)
	}
	return conf
}

func main() {
	// Create postgres client
	psqlClient, err := pkg.NewPsqlClient(context.Background(), cfg)
	if err != nil {
		log.Fatalln("Error create db client:", err)
	}

	// Create redis client
	redisClient, err := pkg.NewRedisClient(context.Background(), cfg)
	if err != nil {
		log.Fatalln("Error create redis client:", err)
	}

	// Init repositories
	courierRepo := courier.NewRepo(psqlClient)
	orderRepo := order.NewRepo(psqlClient)

	r := mux.NewRouter()

	// Couriers
	r.HandleFunc("/couriers", handlers.RateLimiter(handlers.Couriers(courierRepo))).Methods("GET", "POST")
	r.HandleFunc("/couriers/{id:[0-9]+}", handlers.RateLimiter(handlers.CourierId(courierRepo))).Methods("GET")
	r.HandleFunc(
		"/couriers/meta-info/{id:[0-9]+}",
		handlers.RateLimiter(handlers.CourierRating(orderRepo, courierRepo)),
	).Methods("GET")

	// Orders
	r.HandleFunc("/orders", handlers.RateLimiter(handlers.Orders(orderRepo))).Methods("GET", "POST")
	r.HandleFunc("/orders/{id:[0-9]+}", handlers.RateLimiter(handlers.OrderId(orderRepo))).Methods("GET")
	r.HandleFunc(
		"/orders/complete",
		handlers.RateLimiter(handlers.OrderComplete(orderRepo, redisClient)),
	).Methods("POST")

	http.Handle("/", r)

	addr := fmt.Sprintf("%s:%s", cfg.AppCfg.Host, cfg.AppCfg.Port)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalln("Error launch web server:", err)
	}
}
