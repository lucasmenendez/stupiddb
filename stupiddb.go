package stupiddb

import (
	"fmt"
	"os"
	"os/user"
	"regexp"
	"stupiddb/tables"
	"stupiddb/types"
)



type DBError struct {
	Message string
}

func (err DBError) Error() string {
	return fmt.Sprintf("DBError: %v", err.Message)
}



type Query interface {
	NewQuery() *tables.Query
	SetFilters(filters map[string]string)
	SetData(data map[string]string)
}



type Engine struct {
	Name string
	Location string
	Table *tables.Table
	SizesRgx *regexp.Regexp
	HeaderRgx *regexp.Regexp
}

func Instance(schema string) (*Engine, error) {
	user, err := user.Current()
	if err != nil {
		return nil, DBError{"Error getting username."}
	}

	location := user.HomeDir + "/.stupiddb/" + schema + "/"

	var fd *os.File
	if fd, err = os.Open(location); err != nil {
		return nil, DBError{"Error opening database."}
	}

	defer fd.Close()

	sizes_rgx := regexp.MustCompile(`([0-9]*);([0-9]*)`)
	header_rgx := regexp.MustCompile(`(int|float|string|bool)\(([0-9]*)\)([A-Za-z]*);`)

	return &Engine{schema, location, nil, sizes_rgx, header_rgx}, nil
}

func Create(database string) error {
	user, err := user.Current()
	if err != nil {
		return DBError{"Error getting username."}
	}

	path := user.HomeDir + "/.stupiddb/"
	if _, err = os.Stat(path); err != nil {
		if err = os.Mkdir(path, os.ModePerm); err != nil {
			return DBError{"Error overwritting database."}
		}
	}

	if _, err = os.Stat(path + database); err != nil {
		if err = os.Mkdir(path+database, os.ModePerm); err != nil {
			return DBError{"Error creating database."}
		} else {
			if err = os.Mkdir(path + database + "/data", os.ModePerm); err != nil {
				return DBError{"Error data folder."}
			}
			if err = os.Mkdir(path + database + "/index", os.ModePerm); err != nil {
				return DBError{"Error index folder."}
			}
		}
	}
	return nil
}

func (db *Engine) Remove() error {
	if db.Table != nil {
		if err := db.Table.Remove(); err != nil {
			return err
		}
	}

	if err := os.RemoveAll(db.Location); err != nil {
		return DBError{"Error deleting database."}
	}
	return nil
}


func (db *Engine) UseTable(table string) {
	var err error
	if db.Table, err = tables.Use(table, db.Location, db.SizesRgx, db.HeaderRgx); err != nil {
		fmt.Println(err)
		return
	}

	return
}

func (db *Engine) NewTable(table string, fields map[string]types.Type) error {
	return tables.Create(db.Location, table, fields)
}
