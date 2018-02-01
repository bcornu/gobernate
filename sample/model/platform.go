package model

import "time"

type Platform struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Reward      int64     `json:"reward"`
	Duration    int64     `json:"duration"`
	Description string    `json:"description"`
	ImgUrl      string    `json:"image"`
	Missions    []Mission `json:"missions"`
	Active      bool      `json:"active"`
	Visible     bool      `json:"visible"`
	Private     bool      `json:"private"`
	FromDate    time.Time `json:"fromDate"`
	ToDate      time.Time `json:"toDate"`
}
