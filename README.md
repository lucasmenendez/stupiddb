# stupiddb
Ridiculously easy.

## Setup enviroment
Only you need is create folder `.stupiddb` on your HOME directory.

## API

#### Query
Used to make queries to database. Specifies referenced table, data use on query execution and filters to limit results.

###### Methods
- `func Query() *query {}`
Return new query pointer.

- `func (q *query) Table(name string) *query {}`
That assign database to query based on name.

- `func (q *query) Data(data map[string]string) *query {}`
Set key-value data relation to query.

- `func (q *query) Filters(filters map[string]string) *query {}`
Set key-value attributes to query filters.

###### Example
```	
	d := map[string]string {
		"column": "value",
	}
	f := map[string]string {
		"column": "row value",
	}
	query := stupiddb.Query().Table("example").Data(d).Filters(f)

```

#### Engine
Contains database information and is used to make queries.

###### Methods & examples
- `func CreateInstance(database string) error {}`
Create files on filesystem used as database.
```
	if err := stupiddb.CreateInstance("demo_db"); err != nil {
		fmt.Println(err)
		return
	}
```

- `func Instance(schema string) (*engine, error) {}`
Return instance pointer with its attributes.
```
	db, err := stupiddb.Instance("demo_db")
	if err != nil {
		fmt.Println(err)
		return
	}
```

- `func (db *engine) CreateTable(table string, fields []string) error {}`
Create table with name and fields provided. Also you can set a column as unique with mask `stupiddb.Unique(column string)` when set column name on fields string slice daclaration.
```
	fields := []string{
		stupiddb.Unique("col1"),
		stupiddb.Unique("col2"),
		"col3",
	}
	if err := db.CreateTable("demo_table", fields); err != nil {
		fmt.Println(err)
		return
	}
```

- `func (db *engine) Add(q *query) error {}`
Add new row on table specified on query with its data.
```
	data := map[string]string{
		"col1": "row1",
		"col2": "row1",
		"col3": "row1",
	}
	query := stupiddb.Query().Table("demo_table").Data(data1)
	if err := db.Add(query); err != nil {
		fmt.Println(err)
		return
	}
```

- `func (db *engine) Edit(q *query) error {}`
Replace data from query on row specified by query table and filters.
```
	data := map[string]string{
		"col3": "newrow1",
	}

	filters := map[string]string{
		"col1": "row1",
		"col2": "row2",
	}
	query := stupiddb.Query().Table("demo_table").Filters(filters).Data(data)
	if err := db.Edit(query); err != nil {
		fmt.Println(err)
		return
	}
```

- `func (db *engine) Delete(q *query) error {}`
Delete row based on query table and filters.
```
	filters := map[string]string{
		"col1": "row1",
	}

	query := stupiddb.Query().Table("demo_table").Filters(filters)
	if err := db.Delete(query); err != nil {
		fmt.Println(err)
	}
```

- `func (db *engine) Get(q *query) ([]map[string]string, error) {}`
Return all ocurrences on database based on query table and filters.
```
	filters := map[string]string{
		"col1": "row1",
	}
	query := stupiddb.Query().Table("demo_table").Filters(filters)
	
	res, err := db.Get(query)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(res)
```

- `func (db *engine) GetOne(q *query) (map[string]string, error) {}`
Return first ocurrence on database based on query table and filters.
``` 
	filters := map[string]string{
		"col1": "row1",
	}
	query := stupiddb.Query().Table("demo_table").Filters(filters)
	
	res, err := db.GetOne(query)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(res)
```
