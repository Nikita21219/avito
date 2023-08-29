package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"main/cmd/web/handlers"
	"main/internal/cache"
	"main/internal/config"
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

	// Launch cache update
	ctx := context.Background()
	cacheRepo.UpdateCache(ctx, userRepo)

	// Launch delete user segments (ttl)
	userRepo.DeleteSegmentsEveryDay(ctx)

	r := mux.NewRouter()

	r.HandleFunc("/segment", handlers.RateLimiter(
		handlers.Segments(segmentRepo, cacheRepo)),
	).Methods("POST", "DELETE")

	r.HandleFunc("/segment/user", handlers.RateLimiter(
		handlers.Users(userRepo, cacheRepo)),
	).Methods("POST", "GET")

	http.Handle("/", r)

	addr := fmt.Sprintf("%s:%s", cfg.AppCfg.Host, cfg.AppCfg.Port)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalln("Error launch web server:", err)
	}
}
