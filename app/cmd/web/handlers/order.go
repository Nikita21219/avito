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
//	limit, offset, e := getLimitAndOffset(r.URL.Query())
//	if e != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		log.Println("Error parse query string:", e)
//		return
//	}
//
//	ctx := context.Background()
//	orders, e := orderRepo.FindByLimitAndOffset(ctx, limit, offset)
//	if e != nil {
//		log.Println("Error to get orders from db", e)
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//	ordersDto := make([]segment.OrderDto, 0, len(orders))
//	for _, o := range orders {
//		t, e := o.CompletedTime.Value()
//		if e != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			log.Println("Time not valid:", e)
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
//	data, e := json.Marshal(ordersDto)
//	if e != nil {
//		log.Println("Error marshal data:", e)
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//	w.Header().Set("Content-Type", "application/json")
//	_, e = w.Write(data)
//	if e != nil {
//		log.Println("Error write data:", e)
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//}
//
//func pushOrders(w http.ResponseWriter, r *http.Request, orderRepo segment.Repository) {
//	var ordersDto []*segment.OrderDto
//
//	body, e := io.ReadAll(r.Body)
//	if e != nil {
//		w.WriteHeader(500)
//		fmt.Println(e)
//		return
//	}
//	defer r.Body.Close()
//
//	e = json.Unmarshal(body, &ordersDto)
//	if e != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		fmt.Println(e)
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
//	e = orderRepo.CreateAll(ctx, orders)
//	if e != nil {
//		log.Println("Error to create orders:", e)
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
//		id, e := strconv.Atoi(mux.Vars(r)["id"])
//		if e != nil {
//			w.WriteHeader(http.StatusBadRequest)
//			log.Println("Error to convert ascii to int:", e)
//			return
//		}
//
//		ctx := context.Background()
//		o, e := orderRepo.FindOne(ctx, id)
//		if e != nil {
//			log.Println("Error to get segment from db:", e)
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//
//		t, e := o.CompletedTime.Value()
//		if e != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			log.Println("Time not valid:", e)
//		}
//
//		var completedTime string
//		if t != nil {
//			completedTime = o.CompletedTime.Time.Format("2006-01-02")
//		}
//
//		data, e := json.Marshal(segment.OrderDto{
//			Id:            o.Id,
//			Weight:        o.Weight,
//			Region:        o.Weight,
//			DeliveryTime:  o.DeliveryTime,
//			Price:         o.Price,
//			CompletedTime: completedTime,
//		})
//		if e != nil {
//			log.Println("Error marshal data:", e)
//			w.WriteHeader(http.StatusInternalServerError)
//			return
//		}
//		w.Header().Set("Content-Type", "application/json")
//		_, e = w.Write(data)
//		if e != nil {
//			log.Println("Error write data:", e)
//			w.WriteHeader(http.StatusInternalServerError)
//			return
//		}
//	}
//}
//
//func OrderComplete(orderRepo segment.Repository, rdb *redis.Client) http.HandlerFunc {
//	return IdempotentKeyCheckMiddleware(rdb, func(w http.ResponseWriter, r *http.Request, oc *order_complete.OrderCompleteDto) {
//		o, e := orderRepo.FindOne(context.Background(), oc.OrderId)
//		if e != nil {
//			w.WriteHeader(http.StatusBadRequest)
//			log.Println("Error to complete segment:", e)
//			return
//		}
//
//		if o.CourierId.Valid {
//			w.WriteHeader(http.StatusBadRequest)
//			log.Println("Error to complete segment: this segment already has a user")
//			return
//		}
//
//		e = orderRepo.Update(context.Background(), o, oc)
//		if e != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			log.Printf("Error to update segment with id %d: %s", o.Id, e)
//			return
//		}
//		log.Printf("Order with id %d updated successfully", o.Id)
//	})
//}
