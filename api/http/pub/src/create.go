package src

import (
	"errors"
	"fmt"
)

// CreatePayload model info
// @Description A publishing source create payload
type CreatePayload struct {

	// Limit deprecated
	Limit LimitPayload `json:"limit,omitempty"`

	// Src represents a publishing source details
	Src SrcPayload `json:"src"`
}

type LimitPayload struct {
	Freq uint32 `json:"freq,omitempty"`
}

const FreqMin = 1   // once a day
const FreqMax = 288 // every 5 minutes

// SrcPayload model info
// @Description Publishing source details
type SrcPayload struct {

	// Addr source address, e.g. "https://time.com/feed", "https://mastodon.social/@Mastodon", "@proxymtproto"
	Addr string `json:"addr" example:"https://time.com/feed"`

	// Type source type, one of "apub" (ActivityPub), "feed" (web feed), "tgch" (Telegram channel)
	Type string `json:"type,omitempty" example:"feed"`
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
