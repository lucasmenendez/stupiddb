package tables

import (
	"os"
	"fmt"
	"bytes"
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
	Column string
	Location string
	Content []string
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
	var line_length int
	for name, column := range fields {
		var indexable string = ""
		if column.Indexable {
			if err := CreateIndex(location, table, name); err != nil {
				return err
			}
			indexable = "*"
		}

		header += fmt.Sprintf("%s(%d)%s%s;", column.Alias, column.Size, name, indexable)
		line_length += column.Size
	}

	header = fmt.Sprintf("%d;%d\n%s", len(header), line_length, header)
	if _, err = fd.Write([]byte(header)); err != nil {
		return DBError{"Error creating database struct."}
	}

	return nil
}

func CreateIndex(location, table, column string) error {
	var err error
	var index_path string = location + table + "/index"
	if _, err = os.Stat(index_path); err != nil {
		if err = os.Mkdir(index_path, os.ModePerm); err != nil {
			return DBError{"Error creating table index container."}
		}
	}

	var fd *os.File
	if fd, err = os.Create(index_path + "/" + column); err != nil {
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
		if err := os.Remove(index.Location + index.Column); err != nil {
			return DBError{"Error deleting table index."}
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

	var index_path string = location + name + "/index/"
	if _, err = os.Stat(index_path); err != nil {
		return nil, DBError{"Bad formated table: no index. Create it againg."} //Should not be here ever
	}

	var index_files []os.FileInfo
	if index_files, err = ioutil.ReadDir(index_path); err != nil {
		return nil, DBError{"Error getting table index."}
	}

	var index []*Index
	for _, file := range index_files {
		var index_name string = file.Name()
		var index_fd *os.File
		var index_content []string
		var err error

		if index_fd, err = os.Open(index_path + index_name); err != nil {
			return nil, DBError{"Error opening index."}
		}
		defer index_fd.Close()

		scanner := bufio.NewScanner(index_fd)
		for scanner.Scan() {
			index_content = append(index_content, scanner.Text())
		}


		index = append(index, &Index{index_name, index_path, index_content})
	}

	return &Table{name, location, &Header{header_size, columns}, line_size, fd, index}, nil
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

	var columns_to_index map[string]string
	var data []byte
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

	data = bff.Bytes()

	var n int
	if n, err = table.FileDescriptor.Write(data); err != nil || n != table.LineSize {
		return DBError{"Error storing new record."}
	} else if len(columns_to_index) > 0 {
		for column, content := range columns_to_index {
			for _, index := range table.Index {
				if index.Column == column {
					if fd, err := os.Open(index.Location + index.Column); err != nil {
						return DBError{"Error indexing row."}
					} else {
						var l int
						if l, err = fd.Write([]byte(content)); err != nil || l != len(content) {
							return DBError{"Error writting in index."}
						}
					}
				}
			}
		}
	}

	return nil
}

