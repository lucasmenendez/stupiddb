package tables

import (
	"os"
	"fmt"
	"sort"
	"sync"
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
	mutex *sync.Mutex
}

type Header struct {
	Size int
	Columns map[string]types.Type
}

func encodeHeader(fields map[string]types.Type) ([]byte, error) {
	var line_length int
	var header *bytes.Buffer = bytes.NewBuffer([]byte{})
	var columns *bytes.Buffer = bytes.NewBuffer([]byte{})

	var keys []string
	for name := range fields {
		keys = append(keys, name)
	}

	sort.Strings(keys)

	for _, name := range keys {
		var col types.Type = fields[name]

		var index string = ""
		if col.Indexable {
			index = "*"
		}

		fmt.Fprintf(columns, "%s(%d)%s%s;", col.Alias, col.Size, name, index)
		line_length += col.Size
	}

	fmt.Fprintf(header, "%d;%d\n%s", columns.Len(), line_length, columns.Bytes())

	var header_size int = header.Len()
	var new_header *bytes.Buffer = bytes.NewBuffer([]byte{})
	fmt.Fprintf(new_header, "%d;%d\n%s", header_size, line_length, columns.Bytes())

	var new_header_size int = new_header.Len()
	for header_size != new_header_size {
		header_size = new_header_size
		fmt.Fprintf(new_header, "%d;%d\n%s", header_size, line_length, columns.Bytes())

		new_header_size = new_header.Len()
	}

	return new_header.Bytes(), nil
}

func (table *Table) getLine(key string, value interface{}) (int, error) {
	var filter types.Type = table.Header.Columns[key]
	if !filter.Indexable || filter.Empty() {
		return 0, DBError{"Column not found or not indexable."}
	}

	var index *index.Index
	for _, i := range table.Index {
		if key == i.Column {
			index = i
			break
		}
	}

	if index == nil {
		return 0, DBError{"Index not found."}
	}

	filter.Content = value
	filter.Encoder()
	var needle string = string(filter.Content.([]byte))

	index.Mutex.Lock()
	var line_number int = -1
	for line, id := range index.Content {
		if id == needle {
			line_number = line
			break
		}
	}
	index.Mutex.Unlock()

	if line_number == -1 {
		return 0, DBError{"Row not found."}
	}

	return line_number, nil
}

func (table *Table) getContent() (string, error) {
	var err error
	var content string

	if _, err = table.FileDescriptor.Seek(0, 0); err != nil {
		return content, DBError{"Error seeking table file descriptor."}
	}

	var scanner *bufio.Scanner = bufio.NewScanner(table.FileDescriptor)
	if !scanner.Scan() {
		return content, DBError{"Bad formated table."}
	}

	if !scanner.Scan() {
		return content, DBError{"Bad formated table."}
	}
	content = scanner.Text();

	return content, nil
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

	var header []byte
	if header, err = encodeHeader(fields); err != nil {
		return err
	}

	if _, err = fd.Write(header); err != nil {
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

	var path string = fmt.Sprintf("%s%s/data", location, name)
	if fd, err = os.OpenFile(path, os.O_RDWR, os.ModePerm); err != nil {
		return nil, DBError{"Table not found."}
	}

	var scanner *bufio.Scanner = bufio.NewScanner(fd)
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

	fmt.Println(header_size)

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
	var mutex *sync.Mutex = &sync.Mutex{}

	if _, err = fd.Seek(0, 2); err != nil {
		return nil, DBError{"Error seeking table file descriptor."}
	}

	return &Table{name, location, header, line_size, fd, table_index, mutex}, nil
}

func (table *Table) Close() error {
	table.mutex.Lock()
	defer table.mutex.Unlock()

	if err := table.FileDescriptor.Sync(); err != nil {
		return DBError{"Error commiting table record."}
	}

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

	//Sorting map keys to iterate
	var columns []string
	for column := range row {
		columns = append(columns, column)
	}
	sort.Strings(columns)

	for _, column := range columns {
		var t types.Type = table.Header.Columns[column]
		t.Content = row[column]

		if err := t.Encoder(); err != nil {
			return err
		} else { //Ordered writting
			value := t.Content.([]byte)
			if _, err = bff.Write(value); err != nil {
				fmt.Println(err)
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
					return DBError{"Constrain violation on '"+column+"' column"}
				} else {
					if err := index.Append(content); err != nil {
						return err
					}
				}
			}
		}
	}

	table.mutex.Lock()

	var n int
	data = bff.Bytes()
	if n, err = table.FileDescriptor.Write(data); err != nil || n != table.LineSize {
		return DBError{"Error storing new record."}
	}

	if err = table.FileDescriptor.Sync(); err != nil {
		return DBError{"Error commiting new record."}
	}

	table.mutex.Unlock()

	return nil
}

func (table *Table) Edit(row map[string]interface{}) error {
	var err error

	var key string
	var value interface{}

	OUTER:
	for _, index := range table.Index {
		for key, value = range row {
			if index.Column == key {
				break OUTER
			}
		}
	}

	if len(key) < 1 {
		return DBError{"No primary key provided."}
	}

	var line_number int
	if line_number, err = table.getLine(key, value); err != nil {
		return err
	}

	table.mutex.Lock()
	var content string
	if content, err = table.getContent(); err != nil {
		return err
	}

	var columns []string
	for col := range table.Header.Columns {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	var current map[string]types.Type = make(map[string]types.Type, len(columns))

	var row_offset int = line_number * table.LineSize + table.Header.Size
	var row_end int = row_offset + table.LineSize
	var row_content string = content[row_offset:row_end]

	var col_offset int = 0
	for _, col := range columns {
		var data types.Type = table.Header.Columns[col]
		var col_end int = col_offset + data.Size

		data.Content = row_content[col_offset:col_end]
		data.Decoder()

		current[col] = data
		col_offset += data.Size
	}

	var data []byte
	bff := bytes.NewBuffer(data)

	for _, col := range columns {
		var new_col types.Type = current[col]
		for key, value := range row {
			if key != "id" && col == key {
				new_col.Content = value
				break
			}
		}

		if err := new_col.Encoder(); err != nil {
			return err
		} else { //Ordered writting
			value := new_col.Content.([]byte)
			if _, err = bff.Write(value); err != nil {
				return DBError{"Error storing new record."}
			}
		}
	}

	fmt.Println(row_content)
	fmt.Println(len(row_content))
	fmt.Println(bff.String())
	fmt.Println(len(bff.String()))

	var offset int64 = int64(row_offset)
	if _, err = table.FileDescriptor.Seek(offset, 0); err != nil {
		return DBError{"Error deleting row. Rollback seek failed."}
	}

	data = bff.Bytes()
	var l int
	if l, err = table.FileDescriptor.Write(data); err != nil {
		return err
	} else if l != len(data) {
		if err = table.FileDescriptor.Truncate(0); err != nil {
			return DBError{"Error deleting row. Rollback truncate failed."}
		}

		if _, err = table.FileDescriptor.Seek(0, 0); err != nil {
			return DBError{"Error deleting row. Rollback seek failed."}
		}

		var n int
		if n, err = table.FileDescriptor.WriteString(content); err != nil || n != len(content) {
			return DBError{"Error deleting row. Rollback write failed."}
		}
	}

	if err = table.FileDescriptor.Sync(); err != nil {
		return DBError{"Error commiting new record."}
	}

	table.mutex.Unlock()
	return nil
}

func (table *Table) Delete(key string, value interface{}) error {
	var err error

	var line_num int
	if line_num, err = table.getLine(key, value); err != nil {
		return err
	}

	table.mutex.Lock()
	var content string
	if content, err = table.getContent(); err != nil {
		return err
	}

	var offset int = line_num * table.LineSize + table.Header.Size + table.LineSize
	var after_line string = content[offset:]

	var stat os.FileInfo
	var path string = fmt.Sprintf("%s%s/data", table.Location, table.Name)
	if stat, err = os.Stat(path); err != nil {
		return DBError{"Error reading table."}
	}

	var truncate_limit int64 = stat.Size() - int64(len(after_line) + table.LineSize)

	if err = table.FileDescriptor.Truncate(truncate_limit); err != nil {
		return DBError{"Error truncate table file descriptor."}
	}

	if _, err = table.FileDescriptor.Seek(0, 2); err != nil {
		return DBError{"Error seeking table descriptor.."}
	}

	var l int
	if l, err = table.FileDescriptor.WriteString(after_line); err != nil {
		return DBError{"Error deleting row."}
	} else if l != len(after_line) {
		if err = table.FileDescriptor.Truncate(0); err != nil {
			return DBError{"Error deleting row. Rollback truncate failed."}
		}

		if _, err = table.FileDescriptor.Seek(0, 0); err != nil {
			return DBError{"Error deleting row. Rollback seek failed."}
		}

		var n int
		if n, err = table.FileDescriptor.WriteString(content); err != nil || n != len(content) {
			return DBError{"Error deleting row. Rollback write failed."}
		}
	}

	var filter types.Type = table.Header.Columns[key]

	var index *index.Index
	for _, i := range table.Index {
		if key == i.Column {
			index = i
			break
		}
	}

	filter.Content = value
	filter.Encoder()

	var needle string = string(filter.Content.([]byte))
	if err = index.Delete(needle); err != nil {
		return err
	}

	table.mutex.Unlock()

	return nil
}

func (table *Table) Get() ([]map[string]types.Type, error) {
	var err error

	table.mutex.Lock()
	var content string
	if content, err = table.getContent(); err != nil {
		return nil, err
	}
	table.mutex.Unlock()

	var columns []string
	for col := range table.Header.Columns {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	var results []map[string]types.Type
	var row_offset int64 = int64(table.Header.Size)
	var file_length int64 = int64(len(content) - 1)

	for row_offset <= file_length {
		var row_end int64 = row_offset + int64(table.LineSize)
		var row_content string = content[row_offset:row_end]
		row_offset = row_end

		var row map[string]types.Type = make(map[string]types.Type, len(columns))
		var col_offset int = 0
		for _, col := range columns {
			var data types.Type = table.Header.Columns[col]
			var col_end int = col_offset + data.Size

			data.Content = row_content[col_offset:col_end]
			data.Decoder()

			row[col] = data
			col_offset += data.Size
		}
		results = append(results, row)
	}
	return results, err
}

func (table *Table) GetOne(key string, value interface{}) (map[string]types.Type, error) {
	var err error

	var line_number int
	if line_number, err = table.getLine(key, value); err != nil {
		return nil, err
	}

	table.mutex.Lock()
	var content string
	if content, err = table.getContent(); err != nil {
		return nil, err
	}
	table.mutex.Unlock()

	var columns []string
	for col := range table.Header.Columns {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	var result map[string]types.Type = make(map[string]types.Type, len(columns))

	var row_offset int = line_number * table.LineSize + table.Header.Size
	var row_end int = row_offset + table.LineSize
	var row_content string = content[row_offset:row_end]

	var col_offset int = 0
	for _, col := range columns {
		var data types.Type = table.Header.Columns[col]
		var col_end int = col_offset + data.Size

		data.Content = row_content[col_offset:col_end]
		data.Decoder()

		result[col] = data
		col_offset += data.Size
	}

	return result, nil
}
