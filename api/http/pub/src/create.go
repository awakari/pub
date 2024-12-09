package src

import (
	"errors"
	"fmt"
)

type CreatePayload struct {
	Limit LimitPayload `json:"limit,omitempty"`
	Src   SrcPayload   `json:"src"`
}

type LimitPayload struct {
	Freq uint32 `json:"freq,omitempty"`
}

const FreqMin = 1   // once a day
const FreqMax = 288 // every 5 minutes

type SrcPayload struct {
	Addr string `json:"addr"`
	Type string `json:"type,omitempty"`
}

const TypeApub = "apub"
const TypeFeed = "feed"
const TypeSite = "site"
const TypeTgCh = "tgch"
const TypeTgbc = "tgbc"

var errInvalidPayload = errors.New("invalid request payload")

func (cp CreatePayload) validate() (err error) {
	switch cp.Src.Addr {
	case "":
		err = fmt.Errorf("%w: missing source address", errInvalidPayload)
	}
	if err == nil {
		switch cp.Src.Type {
		case TypeFeed:
			if cp.Limit.Freq < FreqMin || cp.Limit.Freq > FreqMax {
				err = fmt.Errorf("%w: missing/invalid feed update frequency: %d per day", errInvalidPayload, cp.Limit.Freq)
			}
		case TypeSite:
		case TypeTgCh:
		case TypeApub:
		default:
			err = fmt.Errorf("%w: unrecognized source type: %s", errInvalidPayload, cp.Src.Type)
		}
	}
	return
}
