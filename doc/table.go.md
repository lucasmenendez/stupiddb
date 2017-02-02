PACKAGE DOCUMENTATION

package tables
    import "./tables/"


FUNCTIONS

func Create(location, table string, fields map[string]types.Type) error
    Create file structure with index, table info with table definition and
    data files. Check if user define one indexable filed at least, else will
    be created one with '_id' key. If table exists or something fails
    returns error with info message.

TYPES

type Header struct {
    Size    int
    Columns map[string]types.Type
}
    Header contain table size and columns definition.

type Table struct {
    Name           string
    Location       string
    Header         *Header
    LineSize       int
    FileDescriptor *os.File
    Index          []*index.Index
    // contains filtered or unexported fields
}
    Table contains table attributes and reference auxiliar structs.

func Use(name, location string, sizes_rgx, header_rgx *regexp.Regexp) (*Table, error)
    Prepare selected database to use it. Create file descriptors and get
    database info and structure. Instance index and header references and
    return 'Table' struct. If something was wrong returns 'DBError'.

func (table *Table) Add(row map[string]interface{}) error
    Create new record on table. First format new record and check if exists
    indexable columns or key already exists on its index, then write new
    record on database row content and update index. If something was wrong
    returns 'DBError', else nil.

func (table *Table) Close() error
    Close database instance. Commit database changes and close file
    descriptor. If something was wrong returns 'DBError', else nil.

func (table *Table) Delete(key string, value interface{}) error
    Remove record from table by key value tuple. First get record line
    number on index, then join previous and next records of query result and
    store on table. If something was worng return a 'DBError' with info
    message, else nil.

func (table *Table) Edit(row map[string]interface{}) error
    Update row by record provided. One of columns must be indexed column at
    least, the rest can be updated. Firs check if one of columns provided
    are indexable, then get record offset and limit, update it and store
    again in same position. If something was wrong returns 'DBError', else
    nil.

func (table *Table) Get() ([]map[string]types.Type, error)
    Get all table records formated. First obteins row table content and
    format each reacord by table definition structure. If something was
    wrong return a 'DBError' with info message, else list of 'Types'.

func (table *Table) GetOne(key string, value interface{}) (map[string]types.Type, error)
    Search and return single record from current table by key value tuple.
    First obtain record line number, then calculate offset and limit to
    search on raw database content and format record by table definition. If
    something was wrong return a 'DBError' with info message, else 'Type'.

func (table *Table) Remove() error
    Remove table with all index and data. Close database and index file
    descriptors, then delete entire table folder.
