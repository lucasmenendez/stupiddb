package query

type Query struct {
	Table string
	Filters map[string]string
	Data map[string]string
}

func NewQuery() *Query {
	return &Query{}
}

func (q *Query) SetTable(name string) *Query {
	q.Table = name
	return q
}

func (q *Query) SetFilters(filters map[string]string) *Query {
	q.Filters = filters
	return q
}

func (q *Query) SetData(data map[string]string) *Query {
	q.Data = data
	return q
}
