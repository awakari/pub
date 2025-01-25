package model

type Subject int

const (
	SubjectUndefined Subject = iota
	SubjectInterests
	SubjectPublishHourly
	SubjectPublishDaily
)

func (s Subject) String() string {
	return [...]string{
		"SubjectUndefined",
		"SubjectInterests",
		"SubjectPublishHourly",
		"SubjectPublishDaily",
	}[s]
}
