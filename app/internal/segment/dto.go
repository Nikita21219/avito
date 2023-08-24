package segment

import "main/pkg/utils"

type OrderDto struct {
	Id            int    `json:"id"`
	Weight        int    `json:"weight"`
	Region        int    `json:"region"`
	DeliveryTime  string `json:"delivery_time"`
	Price         int    `json:"price"`
	CompletedTime string `json:"completed_time"`
}

func (o *OrderDto) Valid() bool {
	return o.Weight >= 0 && o.Region >= 0 && o.Price >= 0 && utils.ValidTime(o.DeliveryTime)
}
