package stupiddb

import (
	"fmt"
	"os"
	"os/user"
	"regexp"
	//"stupiddb/query"
	"stupiddb/tables"
	"stupiddb/types"
)

type DBError struct {
	Message string
}

func (err DBError) Error() string {
	return fmt.Sprintf("DBError: %v", err.Message)
}

type Query struct {
	Table string
	Filters map[string]string
	Data map[string]string
}

type Engine struct {
	Location string
	Table *tables.Table
	SizesRgx *regexp.Regexp
	HeaderRgx *regexp.Regexp
}

func CreateInstance(database string) error {
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
		}
		return nil
	}
	return nil
}

func Instance(schema string) (*Engine, error) {
	user, err := user.Current()
	if err != nil {
		return nil, DBError{"Error getting username."}
	}

	location := user.HomeDir + "/.stupiddb/" + schema

	var fd *os.File
	if fd, err = os.Open(location); err != nil {
		return nil, DBError{"Error opening database."}
	}

	defer fd.Close()

	sizes_rgx := regexp.MustCompile(`([0-9]*);([0-9]*)`)
	header_rgx := regexp.MustCompile(`(int|float|string|bool)\(([0-9]*)\)([A-Za-z]*);`)

	return &Engine{location + "/", nil, sizes_rgx, header_rgx}, nil
}

func (db *Engine) UseTable(table string) {
	fd, err := os.OpenFile(db.Location + table, os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		fmt.Println(DBError{"Table not found."})
		return
	}

	db.Table, err = tables.Use(table, db.Location, fd, db.SizesRgx, db.HeaderRgx)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("%#v", db.Table)
	return
}

func (db *Engine) CloseTable() {
	db.Table.FileDescriptor.Close()
	return
}

func (db *Engine) CreateTable(table string, fields map[string]types.Type) error {
	fd, err := os.Create(db.Location + table)
	if err != nil {
		return DBError{"Error creating database file."}
	}
	defer fd.Close()

	var header string
	var line_length int
	for name, column := range fields {
		header += fmt.Sprintf("%s(%d)%s;", column.Alias, column.Size, name)
		line_length += column.Size
	}
	header = fmt.Sprintf("%d;%d\n%s", len(header), line_length, header)

	if _, err = fd.Write([]byte(header)); err != nil {
		return DBError{"Error creating database struct."}
	}
	return nil
}
