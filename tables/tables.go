package tables

import (
	"os"
	"fmt"
	"bytes"
	"bufio"
	"regexp"
	"strconv"
	"stupiddb/types"
	"stupiddb/index"
)


type DBError struct {
	Message string
}

func (err DBError) Error() string {
	return fmt.Sprintf("DBError: %v", err.Message)
}


type Table struct {
	Name string
	Location string
	Header *Header
	LineSize int
	FileDescriptor *os.File
	Index []*index.Index
}

type Header struct {
	Size int
	Columns map[string]types.Type
}

func formatHeader(fields map[string]types.Type) (string, error) {
	var header string
	var line_length int
	for name, col := range fields {
		var index string = ""
		if col.Indexable {
			index = "*"
		}

		header += fmt.Sprintf("%s(%d)%s%s;", col.Alias, col.Size, name, index)
		line_length += col.Size
	}

	return fmt.Sprintf("%d;%d\n%s", len(header), line_length, header), nil
}

/*
 *	TODO:
 *	- Check one primery key min
 *	- Create other id column else
 */

func Create(location, table string, fields map[string]types.Type) error {
	var fd *os.File
	var err error

	var table_path string = location + table
	var table_data string = table_path + "/data"
	var table_index string = table_path + "/index"

	if err = os.Mkdir(table_path, os.ModePerm); err != nil {
		return DBError{"Error creating database file."}
	} else {
		if fd, err = os.Create(table_data); err != nil {
			return DBError{"Error data folder."}
		}
		defer fd.Close()

		if err = os.Mkdir(table_index, os.ModePerm); err != nil {
			return DBError{"Error index folder."}
		}
	}

	var header string
	if header, err = formatHeader(fields); err != nil {
		return err
	}

	if _, err = fd.Write([]byte(header)); err != nil {
		return DBError{"Error creating database struct."}
	}

	for name, col := range fields {
		if col.Indexable {
			if err := index.New(location, table, name); err != nil {
				return err
			}
		}
	 }

	return nil
}

func (table *Table) Remove() error {
	if table.FileDescriptor != nil {
		table.FileDescriptor.Close()
	}

	for _, index := range table.Index {
		if err := index.Remove(); err != nil {
			return err
		}
	}

	if err := os.RemoveAll(table.Location + table.Name); err != nil {
		return DBError{"Error deleting table data."}
	}

	return nil
}

func Use(name, location string, sizes_rgx, header_rgx *regexp.Regexp) (*Table, error) {
	var err error
	var fd *os.File

	if fd, err = os.OpenFile(location + name + "/data", os.O_RDWR|os.O_APPEND, os.ModePerm); err != nil {
		return nil, DBError{"Table not found."}
	}

	scanner := bufio.NewScanner(fd)
	if !scanner.Scan() {
		return nil, DBError{"Bad formated table."}
	}

	var line_size, header_size int

	sizes_line := scanner.Text()
	res := sizes_rgx.FindStringSubmatch(sizes_line)
	if line_size, err = strconv.Atoi(res[2]); err != nil {
		return nil, err
	}
	if header_size, err = strconv.Atoi(res[1]); err != nil {
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

		var indexable bool = false
		if column[4] == "*" {
			indexable = true
		}

		columns[column[3]] = types.Type{column[1], indexable, column_size, nil}
	}

	var table_index []*index.Index
	if table_index, err = index.Get(location + name + "/index/"); err != nil {
		return nil, err
	}
	var header *Header = &Header{header_size, columns}

	return &Table{name, location, header, line_size, fd, table_index}, nil
}

func (table *Table) Close() error {
	if table.FileDescriptor != nil {
		if err := table.FileDescriptor.Close(); err != nil {
			return DBError{"Error closing table"}
		}
	}
	return nil
}

func (table *Table) Add(row map[string]interface{}) error {
	var err error

	if len(row) != len(table.Header.Columns) {
		return DBError{"Wrong data to insert."}
	}

	var data []byte
	columns_to_index := make(map[string]string, len(table.Index))
	bff := bytes.NewBuffer(data)
	for column, content := range row {
		var t types.Type = table.Header.Columns[column]
		t.Content = content

		if err := t.Encoder(); err != nil {
			return err
		} else { //Ordered writting
			value := t.Content.([]byte)
			if _, err = bff.Write(value); err != nil {
				return DBError{"Error storing new record."}
			}
		}

		if t.Indexable {
			columns_to_index[column] = string(t.Content.([]byte))
		}
	}

	for column, content := range columns_to_index {
		for _, index := range table.Index {
			if index.Column == column {
				if index.Exist(content) {
					return DBError{"Row with '"+column+"' equal to '"+content+"' already exists."}
				} else {
					if err := index.Append(content); err != nil {
						return err
					}
				}
			}
		}
	}

	var n int
	data = bff.Bytes()
	if n, err = table.FileDescriptor.Write(data); err != nil || n != table.LineSize {
		return DBError{"Error storing new record."}
	}

	return nil
}
