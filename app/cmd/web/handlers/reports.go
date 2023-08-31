package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/internal/cache"
	"main/internal/config"
	"main/internal/history"
	"main/internal/reportcsv"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	taskPrefix = "report_task_"
)

type Report struct {
	Status     string `json:"status"`
	LinkToFile string `json:"link_to_file"`
}

// genReport is a function that retrieves data from the database and passes it to a function
// responsible for creating a CSV file.
func genReport(ctx context.Context, historyRepo history.Repository, date time.Time, taskId string) error {
	h, err := historyRepo.GetFromDate(ctx, date)
	if err != nil {
		return err
	}

	if err = reportcsv.CreateReport(h, taskId); err != nil {
		return err
	}

	return nil
}

// launchGenReport is an HTTP handler function that initiates the process of generating a report.
// Upon receiving the request, the function starts a new goroutine to generate the report.
// When the report generation is complete, the task status is updated to "success" in the redis.
func launchGenReport(w http.ResponseWriter, r *http.Request, rdb cache.Repository, historyRepo history.Repository, cfg *config.Config) {
	date, err := GetDateQuery(r.URL.Query())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create task and set "progress" status
	taskId := UniqueKey()
	ttl := 5 * time.Hour
	taskKey := taskPrefix + taskId
	report := Report{Status: "progress"}
	b, err := json.Marshal(report)
	if err != nil {
		log.Println("Error marshal report:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := rdb.Set(context.Background(), taskKey, b, ttl)
	if res != nil && res.Err() != nil {
		log.Println("Error to set task in redis:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err = genReport(ctx, historyRepo, date, taskId); err != nil {
			// If we could not get data from the database, then set the status to "fail"
			log.Println("error to get history:", err)
			reportBytes, err := json.Marshal(Report{Status: "fail"})
			if err != nil {
				log.Println("Error to marshal data:", err)
				return
			}

			res = rdb.Set(ctx, taskKey, reportBytes, ttl)
			if res != nil && res.Err() != nil {
				log.Println("Error to set task in redis:", err)
				return
			}
			return
		}

		linkToFile := fmt.Sprintf(
			"%s://%s:%s/download?id=%s",
			cfg.AppCfg.Scheme,
			cfg.AppCfg.Domain,
			cfg.AppCfg.Port,
			taskId,
		)

		report = Report{
			Status:     "success",
			LinkToFile: linkToFile,
		}
		reportBytes, err := json.Marshal(report)
		if err != nil {
			log.Println("Error to marshal data:", err)
			return
		}

		res = rdb.Set(ctx, taskKey, reportBytes, ttl)
		if res != nil && res.Err() != nil {
			log.Println("error to set result in redis:", err)
			return
		}
	}()

	resp := make(map[string]string)
	resp["task_id"] = taskId
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("Error marshal data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(data); err != nil {
		log.Println("Error write data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// checkReport is an HTTP handler function that allows checking the readiness of a report.
// The report can be in one of three stages: "progress" "success" or "fail"
func checkReport(w http.ResponseWriter, r *http.Request, rdb cache.Repository) {
	id, ok := r.URL.Query()["task_id"]
	if !ok || len(id) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	taskId := id[0]

	ctx := context.Background()
	var report Report
	err := rdb.Get(ctx, taskPrefix+taskId, &report)
	if err != nil {
		log.Println("Error get data from redis:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(report)
	if err != nil {
		log.Println("Error marshal data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if _, err = w.Write(data); err != nil {
		log.Println("Error write data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// DownloadFile returns an HTTP handler function that allows downloading a file by its ID.
func DownloadFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := r.URL.Query()["id"]
		if !ok || len(id) != 1 || strings.Contains(id[0], ".") || strings.Contains(id[0], "/") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fileURL := reportcsv.CSVDir + id[0] + ".csv"
		file, err := os.Open(fileURL)
		if err != nil {
			log.Println("Failed to open file:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		fileName := filepath.Base(fileURL)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

		if _, err = io.Copy(w, file); err != nil {
			log.Println("Failed to write file to response:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func Reports(historyRepo history.Repository, rdb cache.Repository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		launchGenReport(w, r, rdb, historyRepo, cfg)
	}
}

func ReportCheck(rdb cache.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checkReport(w, r, rdb)
	}
}
