package segment

//
//import (
//	"context"
//	"database/sql"
//	"fmt"
//	"main/pkg"
//	"strings"
//	"time"
//)
//
//type repository struct {
//	client pkg.DBClient
//}
//
//func (r *repository) CreateAll(ctx context.Context, orders []*Order) error {
//	q := `INSERT INTO "segment" (weight, region, delivery_time, price) VALUES %s`
//
//	values := make([]string, 0, 4)
//	params := make([]interface{}, 0, len(orders))
//
//	for _, o := range orders {
//		paramsLength := len(params)
//		rowValues := fmt.Sprintf(
//			"($%d, $%d, $%d, $%d)",
//			paramsLength+1,
//			paramsLength+2,
//			paramsLength+3,
//			paramsLength+4,
//		)
//		values = append(values, rowValues)
//		params = append(params, o.Weight, o.Region, o.DeliveryTime, o.Price)
//	}
//
//	q = fmt.Sprintf(q, strings.Join(values, ","))
//	_, err := r.client.Exec(ctx, q, params...)
//	return err
//}
//
//func (r *repository) FindOne(ctx context.Context, id int) (Order, error) {
//	q := `SELECT id, courier_id, weight, region, delivery_time, price, complete_ FROM "segment" WHERE id = $1`
//
//	var completedTime sql.NullTime
//	var o Order
//	if err := r.client.QueryRow(ctx, q, id).Scan(&o.Id, &o.CourierId, &o.Weight, &o.Region, &o.DeliveryTime, &o.Price, &completedTime); err != nil {
//		return Order{}, err
//	}
//	if completedTime.Valid {
//		o.CompletedTime = completedTime
//	}
//
//	return o, nil
//}
//
//func (r *repository) FindByLimitAndOffset(ctx context.Context, l, o int) ([]Order, error) {
//	q := `SELECT id, weight, region, delivery_time, price, complete_time FROM "segment" ORDER BY id LIMIT $1 OFFSET $2`
//	rows, err := r.client.Query(ctx, q, l, o)
//	if err != nil {
//		return nil, err
//	}
//
//	orders := make([]Order, 0, l)
//
//	for rows.Next() {
//		var order Order
//		var completedTime sql.NullTime
//		err = rows.Scan(&order.Id, &order.Weight, &order.Region, &order.DeliveryTime, &order.Price, &completedTime)
//		if completedTime.Valid {
//			order.CompletedTime = completedTime
//		}
//		if err != nil {
//			return nil, err
//		}
//		orders = append(orders, order)
//	}
//	if err = rows.Err(); err != nil {
//		return nil, err
//	}
//
//	return orders, nil
//}
//
//func (r *repository) Update(ctx context.Context, o Order, oc *order_complete.OrderCompleteDto) error {
//	q := `UPDATE "segment" SET courier_id = $1, complete_time = $2 WHERE id = $3`
//	completeTime, err := time.Parse("2006-01-02", oc.OrderCompleteTime)
//	if err != nil {
//		return err
//	}
//	_, err = r.client.Exec(ctx, q, oc.CourierId, completeTime, o.Id)
//	return err
//}
//
//func (r *repository) FindAllInTimeInterval(ctx context.Context, startDate, endDate time.Time, courierId int) ([]Order, error) {
//	q := `
//	SELECT id, courier_id, weight, region, delivery_time, price, complete_time
//	FROM "segment"
//	WHERE complete_time >= $1 AND complete_time < $2 AND courier_id = $3
//	`
//
//	rows, err := r.client.Query(ctx, q, startDate, endDate, courierId)
//	if err != nil {
//		return nil, err
//	}
//
//	orders := make([]Order, 0)
//
//	for rows.Next() {
//		var o Order
//		err = rows.Scan(&o.Id, &o.CourierId, &o.Weight, &o.Region, &o.DeliveryTime, &o.Price, &o.CompletedTime)
//		if err != nil {
//			return nil, err
//		}
//		orders = append(orders, o)
//	}
//	if err = rows.Err(); err != nil {
//		return nil, err
//	}
//
//	return orders, nil
//}
//
//func NewRepo(client pkg.DBClient) Repository {
//	return &repository{
//		client: client,
//	}
//}
