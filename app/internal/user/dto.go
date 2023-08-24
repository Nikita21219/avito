package user

import (
	"main/pkg/utils"
)

type CourierType string

const (
	Foot CourierType = "FOOT"
	Bike CourierType = "BIKE"
	Auto CourierType = "AUTO"
)

type CourierDto struct {
	Id           int         `json:"id"`
	CourierType  CourierType `json:"courier_type"`
	Regions      []int       `json:"regions"`
	WorkingHours []string    `json:"working_hours"`
}

func (c *CourierDto) Valid() (bool, error) {
	switch c.CourierType {
	case Foot, Bike, Auto:
	default:
		return false, nil
	}

	if c.WorkingHours == nil || c.Regions == nil || len(c.Regions) == 0 || c.WorkingHours == nil || len(c.WorkingHours) == 0 {
		return false, nil
	}

	for _, hours := range c.WorkingHours {
		if !utils.ValidTime(hours) {
			return false, nil
		}
	}

	for _, region := range c.Regions {
		if region < 0 {
			return false, nil
		}
	}

	return true, nil
}

type CourierRatingDto struct {
	Rating float64 `json:"rating"`
}
