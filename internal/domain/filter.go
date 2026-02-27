package domain

type Operator string

const (
	FilterEq Operator = "eq"
	FilterGte Operator = "gte"
)

type Filter struct {
	Column string
	Operator Operator
	Value string
}