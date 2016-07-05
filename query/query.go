package query

type Query struct {
	t string
	f map[string]string
	d map[string]string
}

func NewQuery() *Query {
	return &Query{}
}

func (q *Query) Table(name string) *Query {
	q.t = name
	return q
}

func (q *Query) Data(data map[string]string) *Query {
	q.d = data
	return q
}

func (q *Query) Filters(filters map[string]string) *Query {
	q.f = filters
	return q
}
