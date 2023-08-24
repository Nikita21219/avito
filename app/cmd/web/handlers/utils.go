package handlers

import (
	"golang.org/x/time/rate"
	"net/http"
)

//type NextHandler func(w http.ResponseWriter, r *http.Request, oc *order_complete.OrderCompleteDto)

//func getLimitAndOffset(query url.Values) (int, int, error) {
//	offsets, ok := query["offset"]
//	if !ok || len(offsets) != 1 {
//		offsets = []string{"0"}
//	}
//
//	limits, ok := query["limit"]
//	if !ok || len(limits) != 1 {
//		limits = []string{"1"}
//	}
//
//	offset, err := strconv.Atoi(offsets[0])
//	if err != nil || offset < 0 {
//		return -1, -1, err
//	}
//	limit, err := strconv.Atoi(limits[0])
//	if err != nil || limit < 0 {
//		return -1, -1, err
//	}
//
//	return limit, offset, nil
//}
//
//func getStartDateEndDate(query url.Values) (time.Time, time.Time, error) {
//	startDate, ok := query["start_date"]
//	if !ok || len(startDate) != 1 {
//		return time.Time{}, time.Time{}, fmt.Errorf("start_date not valid")
//	}
//
//	endDate, ok := query["end_date"]
//	if !ok || len(endDate) != 1 {
//		return time.Time{}, time.Time{}, fmt.Errorf("end_date not valid")
//	}
//
//	layout := "2006-01-02"
//	_, err := time.Parse(layout, startDate[0])
//	if err != nil {
//		return time.Time{}, time.Time{}, err
//	}
//	_, err = time.Parse(layout, endDate[0])
//	if err != nil {
//		return time.Time{}, time.Time{}, err
//	}
//
//	start, err := time.Parse(layout, startDate[0])
//	if err != nil {
//		return time.Time{}, time.Time{}, err
//	}
//
//	end, err := time.Parse(layout, endDate[0])
//	if err != nil {
//		return time.Time{}, time.Time{}, err
//	}
//
//	if end.Before(start) {
//		return time.Time{}, time.Time{}, fmt.Errorf("End date after the start date")
//	}
//	return start, end, nil
//}
//
//func IdempotentKeyCheckMiddleware(rdb *redis.Client, next NextHandler) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		idempKey := r.Header.Get("Idempotency-Key")
//		if idempKey == "" {
//			log.Println("Idempotency-Key not found in request headers")
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//
//		stream, err := io.ReadAll(r.Body)
//		if err != nil {
//			log.Println("Error to read request body:", err)
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//		defer r.Body.Close()
//
//		oc := &order_complete.OrderCompleteDto{}
//		err = json.Unmarshal(stream, oc)
//		if err != nil {
//			log.Println("Error unmarshal data from body request:", err)
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		} else if !oc.Valid() {
//			log.Println("Data form request not valid")
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//
//		lastTimeParams := rdb.Get(context.Background(), idempKey)
//		if lastTimeParams.Err() != nil {
//			status := rdb.Set(context.Background(), idempKey, oc, 60*60*time.Second)
//			log.Println("Redis set Idempotency-Key status:", status)
//			next(w, r, oc)
//			return
//		}
//
//		b, err := lastTimeParams.Bytes()
//		if err != nil {
//			log.Println("Error convert data from redis to bytes:", err)
//			w.WriteHeader(http.StatusInternalServerError)
//			return
//		}
//
//		ocLast := order_complete.OrderCompleteDto{}
//		err = json.Unmarshal(b, &ocLast)
//		if err != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			return
//		}
//
//		// If params doesn't change
//		if ocLast == *oc {
//			log.Printf("Request with key %s has already been processed", idempKey)
//			w.WriteHeader(http.StatusConflict)
//			return
//		} else {
//			status := rdb.Set(context.Background(), idempKey, oc, 60*60*time.Second)
//			log.Println("Redis set Idempotency-Key status:", status)
//			next(w, r, oc)
//		}
//	}
//}
//
//func GetRatingCourier(orders []segment.Order, startDate, endDate time.Time, courierRepo user.Repository) (float64, error) {
//	if len(orders) < 1 {
//		return -1, fmt.Errorf("Orders not found")
//	}
//	if !orders[0].CourierId.Valid {
//		return -1, fmt.Errorf("CourierId not valid")
//	}
//	hours := endDate.Sub(startDate).Hours()
//	courierId := int(orders[0].CourierId.Int64)
//	c, err := courierRepo.FindOne(context.Background(), courierId)
//	if err != nil {
//		return -1, err
//	}
//	multiplier := 0.0
//	switch c.CourierType {
//	case "FOOT":
//		multiplier = 3.0
//	case "BIKE":
//		multiplier = 2.0
//	case "AUTO":
//		multiplier = 1.0
//	default:
//		return -1, fmt.Errorf("user type \"%s\" not found", c.CourierType)
//	}
//	return float64(len(orders)) / hours * multiplier, nil
//}

func RateLimiter(next http.HandlerFunc) http.HandlerFunc {
	limiter := rate.NewLimiter(10, 10)
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		} else {
			next(w, r)
		}
	}
}
