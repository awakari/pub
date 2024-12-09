package src

import "time"

type ReadPayload struct {
	Addr         string           `json:"addr"`
	GroupId      string           `json:"groupId"`
	UserId       string           `json:"userId,omitempty"`
	LastUpdate   time.Time        `json:"lastUpdate"`
	NextUpdate   time.Time        `json:"nextUpdate"`
	UpdatePeriod time.Duration    `json:"updatePeriod"`
	Usage        UsagePayload     `json:"usage"`
	Push         bool             `json:"push"`
	Counts       map[uint32]int64 `json:"counts"`
	Name         string           `json:"name"`
	Accepted     bool             `json:"accepted"`
	Created      time.Time        `json:"created"`
	Query        string           `json:"query"`
}

type UsagePayload struct {
	Type  UsageType `json:"type"`
	Count int64     `json:"count"`
	Total int64     `json:"total"`
	Limit int64     `json:"limit"`
}

type UsageType int

const (
	UsageTypeUndefined UsageType = iota
	UsageTypeShared
	UsageTypePrivate
)

func (ut UsageType) String() string {
	return [...]string{
		"Undefined",
		"Shared",
		"Private",
	}[ut]
}
