package model

type Subject int

const (
	SubjectUndefined Subject = iota
	SubjectInterests
	SubjectPublishEvents
)

func (s Subject) String() string {
	return [...]string{
		"SubjectUndefined",
		"SubjectInterests",
		"SubjectPublishEvents",
	}[s]
}
