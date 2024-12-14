package model

import "time"

type BlacklistValue struct {
	CreatedAt time.Time
	Reason    string
}

type BlacklistEntry struct {
	Prefix string
	Value  BlacklistValue
}
