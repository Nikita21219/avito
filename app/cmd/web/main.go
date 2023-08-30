package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"main/cmd/web/handlers"
	"main/internal/cache"
	"main/internal/config"
	"main/internal/history"
	"main/internal/segment"
	"main/internal/user"
	"main/pkg"
	"main/pkg/utils"
	"net/http"
)

var cfg *config.Config

func init() {
	cfg = utils.LoadConfig("./config/app.yaml")
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
	userRepo := user.NewRepo(psqlClient)
	segmentRepo := segment.NewRepo(psqlClient)
	cacheRepo := cache.NewRepo(redisClient)
	historyRepo := history.NewRepo(psqlClient)

	// Launch cache update
	ctx := context.Background()
	cacheRepo.UpdateCache(ctx, userRepo)

	// Launch delete user segments (ttl)
	userRepo.DeleteSegmentsEveryDay(ctx, historyRepo)

	r := mux.NewRouter()

	r.HandleFunc("/segment", handlers.RateLimiter(
		handlers.Segments(segmentRepo, cacheRepo)),
	).Methods("POST", "DELETE")

	r.HandleFunc("/segment/user", handlers.RateLimiter(
		handlers.Users(userRepo, cacheRepo, historyRepo)),
	).Methods("POST", "GET")

	r.HandleFunc("/report", handlers.RateLimiter(
		handlers.Reports(historyRepo, cacheRepo, cfg)),
	).Methods("GET")

	r.HandleFunc("/report_check", handlers.RateLimiter(
		handlers.ReportCheck(cacheRepo)),
	).Methods("GET")

	r.HandleFunc("/download", handlers.RateLimiter(
		handlers.DownloadFile()),
	).Methods("GET")

	http.Handle("/", r)

	addr := fmt.Sprintf("%s:%s", cfg.AppCfg.Host, cfg.AppCfg.Port)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalln("Error launch web server:", err)
	}
}
