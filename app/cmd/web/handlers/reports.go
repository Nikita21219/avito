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
	"time"
)

var (
	taskPrefix = "report_task_"
)

type Report struct {
	Status     string `json:"status"`
	LinkToFile string `json:"link_to_file"`
}

func (r *Report) MarshalBinary() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Report) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}

// TODO fill doc
func genReport(ctx context.Context, historyRepo history.Repository, date time.Time, taskId string) error {
	// TODO check context timeout

	h, err := historyRepo.GetFromDate(ctx, date)
	if err != nil {
		return err
	}

	if err = reportcsv.CreateReport(h, taskId); err != nil {
		return err
	}

	return nil
}

// TODO fill doc
func launchGenReport(w http.ResponseWriter, r *http.Request, rdb cache.Repository, historyRepo history.Repository, cfg *config.Config) {
	date, ok := r.URL.Query()["date"]
	if !ok || len(date) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	t, err := time.Parse("2006-01-02 15:04", date[0])
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

	err = rdb.Set(context.Background(), taskKey, b, ttl).Err()
	if err != nil {
		log.Println("Error to set task in redis:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err = genReport(ctx, historyRepo, t, taskId); err != nil {
			// If we could not get data from the database, then set the status to "fail"
			report = Report{Status: "fail"}
			reportBytes, err := json.Marshal(report)
			if err != nil {
				log.Println("Error to marshal data:", err)
				return
			}
			rdb.Set(ctx, taskKey, reportBytes, ttl)
			log.Println("error to get history:", err)
			return
		}

		linkToFile := fmt.Sprintf("http://%s:%s/download?id=%s", cfg.AppCfg.Host, cfg.AppCfg.Port, taskId)
		report = Report{
			Status:     "success",
			LinkToFile: linkToFile,
		}
		reportBytes, err := json.Marshal(report)
		if err != nil {
			log.Println("Error to marshal data:", err)
			return
		}

		err = rdb.Set(ctx, taskKey, reportBytes, ttl).Err()
		if err != nil {
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
	_, err = w.Write(data)
	if err != nil {
		log.Println("Error write data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func DownloadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := r.URL.Query()["id"]
		if !ok || len(id) != 1 {
			// TODO check, if id has dots. For example ../../filePath
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fileURL := reportcsv.CSVDir + id[0] + ".csv"
		file, err := os.Open(fileURL)
		if err != nil {
			log.Println("Failed to open file:", err)
			w.WriteHeader(http.StatusInternalServerError)
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

// TODO fill doc
func Reports(historyRepo history.Repository, rdb cache.Repository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		launchGenReport(w, r, rdb, historyRepo, cfg)
	}
}

// TODO fill doc
func ReportCheck(rdb cache.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checkReport(w, r, rdb)
	}
}
