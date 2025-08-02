package src

import "time"

// ReadPayload model info
// @Description An existing publishing source details
type ReadPayload struct {

	// Addr publishing source address
	Addr string `json:"addr" example:"https://time.com/feed"`

	// GroupId owner group id
	GroupId string `json:"groupId" example:"default"`

	// UserId owner user id, the source is dedicated if not set
	UserId string `json:"userId,omitempty" example:"tg://user?id=1234567890"`

	// LastUpdate time of last event from the source or time of last check for updates
	LastUpdate time.Time `json:"lastUpdate" example:"2022-03-20T00:00:00Z"`

	// NextUpdate the scheduled time of the next check for updates (polling-based sources only)
	NextUpdate time.Time `json:"nextUpdate" example:"2022-03-20T00:00:00Z"`

	// Usage publishing limits utilization details
	Usage UsagePayload `json:"usage"`

	// Push defines whether the source submits new events using push, the source is being checked with polls when false
	Push bool `json:"push" example:"false"`

	// Counts frequency statistic used to schedule the next check time for polling-based sources
	Counts map[uint32]int64 `json:"counts"`

	// Name a source name
	Name string `json:"name" example:"Time News"`

	// Accepted defines whether a source explicitly confirmed their registration
	Accepted bool `json:"accepted" example:"true"`

	// Created time when source was registered
	Created time.Time `json:"created" example:"2022-03-20T00:00:00Z"`

	// Query user query that caused this source registration
	Query string `json:"query" example:"news"`
}

// UsagePayload model info
// @Description publishing limits utilization details
type UsagePayload struct {

	// Type may be 1 for shared (dedicated) sources or 2 for a private
	Type UsageType `json:"type" example:"2"`

	// Count deprecated
	Count int64 `json:"count"` // deprecated = CountHourly

	// CountHourly number of events published during 1 hour
	CountHourly int64 `json:"countHourly" example:"5"`

	// CountDaily number of events published during 1 day
	CountDaily int64 `json:"countDaily" example:"10"`

	// Total number of published events since the registration
	Total int64 `json:"total" example:"42"`

	// Limit deprecated
	Limit int64 `json:"limit"` // deprecated = LimitHourly

	// LimitHourly max number of events allowed to publish during 1 hour
	LimitHourly int64 `json:"limitHourly" example:"10"`

	// LimitDaily max number of events allowed to publish during 1 day
	LimitDaily int64 `json:"limitDaily" example:"100"`
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
