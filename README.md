# stupiddb
Ridiculously easy.

## API

##### ```func engine.CreateInstance(name string) error {}```
Create a database instance with name provided.

##### ```func engine.Instance(name string) (*Instance, error) {}```
Initialize a database instance.

##### ```func (database *Instance) database.CreateTable(name string, columns []string) error {}``` 
Create table with name and columns provided. Columns can be UNIQUE.
Example:
```
	columns := []string {
		engine.Unique("col1"),
		"col2",
		"col3",
	}
```

##### ```func (database *Instance) database.Add(q *query.Query) error {}```
The ```Query``` struct contains three attributes:
* ```T```: Table refered name.
* ```F```: Filters to applay to query results.
* ```D```: Data provided to use them into query exec.
And ```Add``` functions insert query data into table provided.

##### ```func (database *Instance) database.Get(q *query.Query) ([]map[string]string, error) {}```
Get all itemss stored on table provided, applying the filters.

##### ```func (database *Instance) database.GetOne(q *query.Query) (map[string]string, error) {}```
Return the first item stored on table provided, applying the filters.
