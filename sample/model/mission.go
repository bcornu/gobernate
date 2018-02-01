package model

import "time"

type Mission struct {
	Id       int64        `json:"id"`
	Name     string       `json:"name"`
	Type     *MissionType `json:"type"`
	Inherit  bool         `json:"inherit"`
	Active   bool         `json:"active"`
	Visible  bool         `json:"visible"`
	Private  bool         `json:"private"`
	FromDate time.Time    `json:"fromDate"`
	ToDate   time.Time    `json:"toDate"`
}
