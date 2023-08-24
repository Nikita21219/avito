package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/internal/segment"
	"net/http"
)

func createSegment(w http.ResponseWriter, r *http.Request, segmentRepo segment.Repository) {
	ctx := context.Background()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	defer r.Body.Close()

	var s *segment.SegmentDto
	err = json.Unmarshal(body, &s)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	err = segmentRepo.Create(ctx, &segment.Segment{
		Name: s.Name,
	})
	if err != nil {
		log.Println("Error to create segment:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func deleteSegment(w http.ResponseWriter, r *http.Request, segmentRepo segment.Repository) {
	//var cours []*user.CourierDto
	//
	//body, err := io.ReadAll(r.Body)
	//if err != nil {
	//	w.WriteHeader(500)
	//	fmt.Println(err)
	//	return
	//}
	//defer r.Body.Close()
	//
	//err = json.Unmarshal(body, &cours)
	//if err != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	fmt.Println(err)
	//	return
	//}
	//
	//couriers := make([]*user.Courier, 0, len(cours))
	//for _, cour := range cours {
	//	ok, _ := cour.Valid()
	//	if !ok {
	//		w.WriteHeader(http.StatusBadRequest)
	//		return
	//	}
	//	couriers = append(couriers, &user.Courier{
	//		Id:           cour.Id,
	//		CourierType:  cour.CourierType,
	//		Regions:      cour.Regions,
	//		WorkingHours: cour.WorkingHours,
	//	})
	//}
	//
	//ctx := context.Background()
	//err = segmentRepo.CreateAll(ctx, couriers)
	//if err != nil {
	//	log.Println("Error to create couriers:", err)
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}
}

func Segments(segmentRepo segment.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			createSegment(w, r, segmentRepo)
		} else if r.Method == "DELETE" {
			deleteSegment(w, r, segmentRepo)
		}
	}
}

//func CourierId(courierRepo user.Repository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		id, err := strconv.Atoi(mux.Vars(r)["id"])
//		if err != nil {
//			w.WriteHeader(http.StatusBadRequest)
//			log.Println("Error to convert ascii to int:", err)
//			return
//		}
//
//		ctx := context.Background()
//		c, err := courierRepo.FindOne(ctx, id)
//		if err != nil {
//			log.Println("Error to get user from db", err)
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//		data, err := json.Marshal(c)
//		if err != nil {
//			log.Println("Error marshal data:", err)
//			w.WriteHeader(http.StatusInternalServerError)
//			return
//		}
//		w.Header().Set("Content-Type", "application/json")
//		_, err = w.Write(data)
//		if err != nil {
//			log.Println("Error write data:", err)
//			w.WriteHeader(http.StatusInternalServerError)
//			return
//		}
//	}
//}
//
//func CourierRating(orderRepo segment.Repository, courierRepo user.Repository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		startDate, endDate, err := getStartDateEndDate(r.URL.Query())
//		if err != nil {
//			w.WriteHeader(http.StatusBadRequest)
//			log.Println("Error to parse request query string:", err)
//			return
//		}
//
//		courierId, err := strconv.Atoi(mux.Vars(r)["id"])
//		if err != nil {
//			w.WriteHeader(http.StatusBadRequest)
//			log.Println("Error to get id from path request:", err)
//			return
//		}
//
//		orders, err := orderRepo.FindAllInTimeInterval(context.Background(), startDate, endDate, courierId)
//		if err != nil {
//			log.Println("Error to find all orders in time interval:", err)
//			return
//		}
//
//		if len(orders) == 0 {
//			log.Println("Orders not found")
//			w.WriteHeader(http.StatusOK)
//			return
//		}
//		rating, err := GetRatingCourier(orders, startDate, endDate, courierRepo)
//		if err != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			log.Printf("Error to get rating user with id %d: %s\n", courierId, err)
//			return
//		}
//
//		b, err := json.Marshal(user.CourierRatingDto{
//			Rating: rating,
//		})
//
//		w.Header().Set("Content-Type", "application/json")
//		_, err = w.Write(b)
//		if err != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			log.Println("Error to write data:", err)
//			return
//		}
//	}
//}
