package pub

// PublishResponse model info
// @Description info about events submitted
type PublishResponse struct {

	// AckCount represents the number of successfully submitted events. Order is preserved.
	AckCount uint32 `json:"ackCount"`
}
