package tables

import (
	"os"
	"fmt"
	"bufio"
	"regexp"
	"strconv"
	"io/ioutil"
	"stupiddb/types"
)


type DBError struct {
	Message string
}

func (err DBError) Error() string {
	return fmt.Sprintf("DBError: %v", err.Message)
}


type Query struct {
	Filters map[string]string
	Data map[string]string
}

func NewQuery() *Query {
	return &Query{}
}

func (q *Query) SetFilters(filters map[string]string) {
	q.Filters = filters
	return
}

func (q *Query) SetData(data map[string]string) {
	q.Data = data
	return
}


type Table struct {
	Name string
	Location string
	Header *Header
	LineSize int
	FileDescriptor *os.File
	Index []*Index
}

type Header struct {
	Size int
	Columns map[string]types.Type
}

type Index struct {
	Cursor int
	Location string
	FileDescriptor *os.File
}

/*
 *	TODO:
 *	- Check one primery key min
 *	- Create other id column else
 */

func Create(location, table string, fields map[string]types.Type) error {
	fd, err := os.Create(location + "data/" + table)
	if err != nil {
		return DBError{"Error creating database file."}
	}
	defer fd.Close()

	var header string
	var line_length int
	for name, column := range fields {
		header += fmt.Sprintf("%s(%d)%s;", column.Alias, column.Size, name)
		line_length += column.Size + 1 //add separator character
		if column.Indexable {
			if err := CreateIndex(location, table, name); err != nil {
				return err
			}
		}
	}
	header = fmt.Sprintf("%d;%d\n%s", len(header), line_length, header)

	if _, err = fd.Write([]byte(header)); err != nil {
		return DBError{"Error creating database struct."}
	}

	return nil
}

func CreateIndex(location, table, column string) error {
	var err error
	indexContainer := location + "index/" + table
	if _, err = os.Stat(indexContainer); err != nil {
		if err = os.Mkdir(indexContainer, os.ModePerm); err != nil {
			return DBError{"Error creating table index container."}
		}
	}

	fd, err := os.Create(indexContainer + "/" + column)
	if err != nil {
		return DBError{"Error creating column index."}
	}
	fd.Close()
	return nil
}

func (table *Table) Remove() error {
	if table.FileDescriptor != nil {
		table.FileDescriptor.Close()
	}

	for _, index := range table.Index {
		if err := os.Remove(index.Location); err != nil {
			return DBError{"Error deleting table index."}
		}
	}

	if err := os.Remove(table.Location); err != nil {
		return DBError{"Error deleting table."}
	}

	return nil
}

func Use(name, location string, sizes_rgx, header_rgx *regexp.Regexp) (*Table, error) {
	var err error

	fd, err := os.OpenFile(location + "data/" + name, os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, DBError{"Table not found."}
	}

	scanner := bufio.NewScanner(fd)
	if !scanner.Scan() {
		return nil, DBError{"Bad formated table."}
	}

	var line_size, header_size int

	sizes_line := scanner.Text()
	res := sizes_rgx.FindStringSubmatch(sizes_line)
	if line_size, err = strconv.Atoi(res[1]); err != nil {
		return nil, err
	}
	if header_size, err = strconv.Atoi(res[2]); err != nil {
		return nil, err
	}

	if !scanner.Scan() {
		return nil, DBError{"Bad formated table."}
	}

	var columns map[string]types.Type = make(map[string]types.Type)
	var header_line string = scanner.Text()
	var header_items [][]string = header_rgx.FindAllStringSubmatch(header_line, -1)

	for _, column := range header_items {
		var column_size int
		if column_size, err = strconv.Atoi(column[2]); err != nil {
			return nil, err
		}
		columns[column[3]] = types.Type{column[1], false, column_size, nil}
	}

	index_path := location + "index/" + name
	//var index []Index
	if _, err = os.Stat(index_path); err != nil {
		return nil, DBError{"Bad formated table: no index. Create it againg."} //Should not be here ever
	}

	index_files, err := ioutil.ReadDir(index_path)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("%#v", index_files)

	return &Table{name, location + name, &Header{header_size, columns}, line_size, fd, nil}, nil
}

func (table *Table) Close() error {
	if table.FileDescriptor != nil {
		if err := table.FileDescriptor.Close(); err != nil {
			return DBError{"Error closing table"}
		}
	}
	return nil
}
