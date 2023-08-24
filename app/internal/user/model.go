package user

type Courier struct {
	Id           int         `json:"id"`
	CourierType  CourierType `json:"courier_type"`
	Regions      []int       `json:"regions"`
	WorkingHours []string    `json:"working_hours"`
}
