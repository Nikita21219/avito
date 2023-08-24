package segment

import (
	"context"
	"main/pkg"
)

type repository struct {
	client pkg.DBClient
}

func (r *repository) Create(ctx context.Context, segment *Segment) error {
	q := `INSERT INTO segments (segment_name) VALUES ($1) RETURNING segment_id`
	_, err := r.client.Exec(ctx, q, segment.Name)
	return err
}

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

func NewRepo(client pkg.DBClient) Repository {
	return &repository{
		client: client,
	}
}
