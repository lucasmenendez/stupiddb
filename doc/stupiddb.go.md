PACKAGE DOCUMENTATION

package stupiddb
    import "stupiddb"


FUNCTIONS

func Create(database string) error
    Create database structure. First check if stupiddb folder exists, else
    will be created, then check if database folder structure already exists,
    else will be created. If something was wrong, returns 'DBError' with
    info message, else will be nil.

TYPES

type Engine struct {
    Name      string
    Location  string
    Table     *tables.Table
    SizesRgx  *regexp.Regexp
    HeaderRgx *regexp.Regexp
}
    Represents database struct and contains its name, location, table
    reference and compiled regex to decode header information.

func Instance(schema string) (*Engine, error)
    Create a 'Engine' by its name. Check if database exists trying no error
    while open database location folder, then compile header and size regex
    and return 'Engine' reference. If something was wrong returns 'DBError'
    with info message, else will be nil.

func (db *Engine) NewTable(table string, fields map[string]types.Type) error
    Create table with provided structure. Call 'Create' table function and
    pass name and fields attributes. If something was wrong return 'DBError'
    with info message, else will be nil.

func (db *Engine) Remove() error
    Remove database. Check if table is opened and remove it, then remove
    entire database folder structure. If something was wrong return
    'DBError' with info message, else will be nil.

func (db *Engine) UseTable(table string) error
    Prepare table provided to be ready. Call 'Use' table function with
    database required attributes. If something was wrong return 'DBError'
    with info message, else will be nil.

SUBDIRECTORIES

	dberror
	decoder
	encoder
	globals
	index
	tables
	types
