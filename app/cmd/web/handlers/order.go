package handlers

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"github.com/gorilla/mux"
//	"github.com/redis/go-redis/v9"
//	"io"
//	"log"
//	"main/internal/order_complete"
//	"main/internal/segment"
//	"net/http"
//	"strconv"
//)
//
//func getOrders(w http.ResponseWriter, r *http.Request, orderRepo segment.Repository) {
//	limit, offset, err := getLimitAndOffset(r.URL.Query())
//	if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		log.Println("Error parse query string:", err)
//		return
//	}
//
//	ctx := context.Background()
//	orders, err := orderRepo.FindByLimitAndOffset(ctx, limit, offset)
//	if err != nil {
//		log.Println("Error to get orders from db", err)
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//	ordersDto := make([]segment.OrderDto, 0, len(orders))
//	for _, o := range orders {
//		t, err := o.CompletedTime.Value()
//		if err != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			log.Println("Time not valid:", err)
//			return
//		}
//		var completedTime string
//		if t != nil {
//			completedTime = o.CompletedTime.Time.Format("2006-01-02")
//		}
//
//		ordersDto = append(ordersDto, segment.OrderDto{
//			Id:            o.Id,
//			Weight:        o.Weight,
//			Region:        o.Region,
//			DeliveryTime:  o.DeliveryTime,
//			Price:         o.Price,
//			CompletedTime: completedTime,
//		})
//	}
//
//	data, err := json.Marshal(ordersDto)
//	if err != nil {
//		log.Println("Error marshal data:", err)
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//	w.Header().Set("Content-Type", "application/json")
//	_, err = w.Write(data)
//	if err != nil {
//		log.Println("Error write data:", err)
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//}
//
//func pushOrders(w http.ResponseWriter, r *http.Request, orderRepo segment.Repository) {
//	var ordersDto []*segment.OrderDto
//
//	body, err := io.ReadAll(r.Body)
//	if err != nil {
//		w.WriteHeader(500)
//		fmt.Println(err)
//		return
//	}
//	defer r.Body.Close()
//
//	err = json.Unmarshal(body, &ordersDto)
//	if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		fmt.Println(err)
//		return
//	}
//
//	orders := make([]*segment.Order, 0, len(ordersDto))
//	for _, o := range ordersDto {
//		if !o.Valid() {
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//
//		orders = append(orders, &segment.Order{
//			Weight:       o.Weight,
//			Region:       o.Region,
//			DeliveryTime: o.DeliveryTime,
//			Price:        o.Price,
//		})
//	}
//
//	ctx := context.Background()
//	err = orderRepo.CreateAll(ctx, orders)
//	if err != nil {
//		log.Println("Error to create orders:", err)
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//}
//
//func Orders(orderRepo segment.Repository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		if r.Method == "GET" {
//			getOrders(w, r, orderRepo)
//			return
//		} else if r.Method == "POST" {
//			pushOrders(w, r, orderRepo)
//		}
//	}
//}
//
//func OrderId(orderRepo segment.Repository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		id, err := strconv.Atoi(mux.Vars(r)["id"])
//		if err != nil {
//			w.WriteHeader(http.StatusBadRequest)
//			log.Println("Error to convert ascii to int:", err)
//			return
//		}
//
//		ctx := context.Background()
//		o, err := orderRepo.FindOne(ctx, id)
//		if err != nil {
//			log.Println("Error to get segment from db:", err)
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//
//		t, err := o.CompletedTime.Value()
//		if err != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			log.Println("Time not valid:", err)
//		}
//
//		var completedTime string
//		if t != nil {
//			completedTime = o.CompletedTime.Time.Format("2006-01-02")
//		}
//
//		data, err := json.Marshal(segment.OrderDto{
//			Id:            o.Id,
//			Weight:        o.Weight,
//			Region:        o.Weight,
//			DeliveryTime:  o.DeliveryTime,
//			Price:         o.Price,
//			CompletedTime: completedTime,
//		})
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
//func OrderComplete(orderRepo segment.Repository, rdb *redis.Client) http.HandlerFunc {
//	return IdempotentKeyCheckMiddleware(rdb, func(w http.ResponseWriter, r *http.Request, oc *order_complete.OrderCompleteDto) {
//		o, err := orderRepo.FindOne(context.Background(), oc.OrderId)
//		if err != nil {
//			w.WriteHeader(http.StatusBadRequest)
//			log.Println("Error to complete segment:", err)
//			return
//		}
//
//		if o.CourierId.Valid {
//			w.WriteHeader(http.StatusBadRequest)
//			log.Println("Error to complete segment: this segment already has a user")
//			return
//		}
//
//		err = orderRepo.Update(context.Background(), o, oc)
//		if err != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			log.Printf("Error to update segment with id %d: %s", o.Id, err)
//			return
//		}
//		log.Printf("Order with id %d updated successfully", o.Id)
//	})
//}
