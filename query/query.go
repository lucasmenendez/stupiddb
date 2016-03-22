package query

type Query struct {
	T string
	F map[string]string
	D map[string]string
}

func New() *Query {
	return &Query{}
}

func (q *Query) Table(name string) *Query {
	q.T = name
	return q
}

func (q *Query) Data(data map[string]string) *Query {
	q.D = data
	return q
}

func (q *Query) Filters(filters map[string]string) *Query {
	q.F = filters
	return q
}
