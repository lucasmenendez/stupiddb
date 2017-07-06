package stupiddb

import (
	"fmt"
	"os"
	"regexp"

	"github.com/lucasmenendez/stupiddb/globals"
	"github.com/lucasmenendez/stupiddb/dberror"
	"github.com/lucasmenendez/stupiddb/tables"
	"github.com/lucasmenendez/stupiddb/types"
)

//Represents database struct and contains its name, location, table reference
//and compiled regex to decode header information.
type Engine struct {
	Name string
	Location string
	Table *tables.Table
	SizesRgx *regexp.Regexp
	HeaderRgx *regexp.Regexp
}

//Create a 'Engine' by its name. Check if database exists trying no error while
//open database location folder, then compile header and size regex and return
//'Engine' reference. If something was wrong returns 'DBError' with info
//message, else will be nil.
func Instance(schema string) (*Engine, error) {
	var err error
	location := fmt.Sprintf("%s%s/", globals.RootPath, schema)

	var fd *os.File
	if fd, err = os.Open(location); err != nil {
		return nil, dberror.DBError{"Error opening database."}
	}

	defer fd.Close()

	sizes_rgx	:= regexp.MustCompile(globals.SizesRgx)
	header_rgx	:= regexp.MustCompile(globals.HeaderRgx)

	return &Engine{schema, location, nil, sizes_rgx, header_rgx}, nil
}

//Create database structure. First check if stupiddb folder exists, else will
//be created, then check if database folder structure already exists, else will
//be created. If something was wrong, returns 'DBError' with info message,
//else will be nil.
func Create(database string) error {
	var err error
	if _, err = os.Stat(globals.RootPath); err != nil {
		if err = os.Mkdir(globals.RootPath, os.ModePerm); err != nil {
			return dberror.DBError{"Error overwritting database."}
		}
	}

	var database_path string = fmt.Sprintf("%s%s", globals.RootPath, database)
	if _, err = os.Stat(database_path); err != nil {
		if err = os.Mkdir(database_path, os.ModePerm); err != nil {
			return dberror.DBError{"Error creating database."}
		}
	}
	return nil
}

//Remove database. Check if table is opened and remove it, then remove entire
//database folder structure. If something was wrong return 'DBError' with info
//message, else will be nil.
func (db *Engine) Remove() error {
	if db.Table != nil {
		if err := db.Table.Remove(); err != nil {
			return err
		}
	}

	if err := os.RemoveAll(db.Location); err != nil {
		return dberror.DBError{"Error deleting database."}
	}
	return nil
}

//Prepare table provided to be ready. Call 'Use' table function with database
//required attributes. If something was wrong return 'DBError' with info
//message, else will be nil. 
func (db *Engine) UseTable(table string) error {
	var err error
	db.Table, err = tables.Use(table, db.Location, db.SizesRgx, db.HeaderRgx)

	return err
}

//Create table with provided structure. Call 'Create' table function and pass
//name and fields attributes. If something was wrong return 'DBError' with info
//message, else will be nil.
func (db *Engine) NewTable(table string, fields map[string]types.Type) error {
	return tables.Create(db.Location, table, fields)
}
