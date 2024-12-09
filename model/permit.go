package model

type Permit struct {
	Count  uint32
	UserId string

	// JustExhausted represents the given permit has been just exhausted for the 1st time after its reset.
	JustExhausted bool
}
