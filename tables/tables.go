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

	"github.com/lucasmenendez/stupiddb/dberror"
	"github.com/lucasmenendez/stupiddb/types"
	"github.com/lucasmenendez/stupiddb/index"
)

//Table contains table attributes and reference auxiliar structs.
type Table struct {
	Name string
	Location string
	Header *Header
	LineSize int
	FileDescriptor *os.File
	Index []*index.Index
	mutex *sync.Mutex
}

//Header contain table size and columns definition.
type Header struct {
	Size int
	Columns map[string]types.Type
}

//Generate part of table definition by formated concatenation of
//name, type and length of each column and returns []byte with.
func encodeHeader(fields map[string]types.Type) ([]byte, error) {
	var line_length int
	var header *bytes.Buffer	= bytes.NewBuffer([]byte{})
	var columns *bytes.Buffer	= bytes.NewBuffer([]byte{})

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

	fmt.Fprintf(header, "%d\n%s", line_length, columns.Bytes())

	return header.Bytes(), nil
}

//Search on index by key value tuple and return line number where is located.
//Iterate over key index content to find by value provided.
func (table *Table) getLine(key string, value interface{}) (int, error) {
	var filter types.Type = table.Header.Columns[key]
	if !filter.Indexable || filter.Empty() {
		return 0, dberror.DBError{"Column not found or not indexable."}
	}

	var index *index.Index
	for _, i := range table.Index {
		if key == i.Column {
			index = i
			break
		}
	}

	if index == nil {
		return 0, dberror.DBError{"Index not found."}
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
		return 0, dberror.DBError{"Row not found."}
	}

	return line_number, nil
}

//Seek table file descriptor to initial position and read all table content
//to search into them. This is horrible. I'm working on.
//If something fails returns a 'DBError' with info message.
func (table *Table) getContent() (string, error) {
	var err error
	var content string

	if _, err = table.FileDescriptor.Seek(0, 0); err != nil {
		return content, dberror.DBError{"Error seeking table file descriptor."}
	}

	var scanner *bufio.Scanner = bufio.NewScanner(table.FileDescriptor)
	if !scanner.Scan() {
		return content, dberror.DBError{"Error or empty table."}
	}
	content = scanner.Text();

	return content, nil
}

//Create file structure with index, table info with table definition
//and data files. Check if user define one indexable filed at least,
//else will be created one with '_id' key. If table exists or something
//fails returns error with info message.
func Create(location, table string, fields map[string]types.Type) error {
	var fd *os.File
	var err error

	var table_path string	= fmt.Sprintf("%s%s", location, table)
	var table_data string	= fmt.Sprintf("%s/data", table_path)
	var table_index string	= fmt.Sprintf("%s/index", table_path)
	var table_info string	= fmt.Sprintf("%s/info", table_path)

	if err = os.Mkdir(table_path, os.ModePerm); err != nil {
		return dberror.DBError{"Error creating database file."}
	} else {
		if fd, err = os.Create(table_data); err != nil {
			return dberror.DBError{"Error table data."}
		}
		fd.Close()

		if fd, err = os.Create(table_info); err != nil {
			return dberror.DBError{"Error table info."}
		}
		defer fd.Close()

		if err = os.Mkdir(table_index, os.ModePerm); err != nil {
			return dberror.DBError{"Error table index folder."}
		}
	}

	var indexables []string
	for name, col := range fields {
		if col.Indexable {
			indexables = append(indexables, name)
		}
	 }

	if len(indexables) == 0 {
		fields["_id"] = types.Int(true)
		indexables = append(indexables, "_id")
	}

	for _, name := range indexables {
		if err := index.New(location, table, name); err != nil {
			return err
		}
	}

	var info []byte
	if info, err = encodeHeader(fields); err != nil {
		return err
	}

	if _, err = fd.Write(info); err != nil {
		return dberror.DBError{"Error creating database info."}
	}

	return nil
}

//Remove table with all index and data. Close database and index file 
//descriptors, then delete entire table folder.
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
		return dberror.DBError{"Error deleting table data."}
	}

	return nil
}

//Prepare selected database to use it. Create file descriptors and get
//database info and structure. Instance index and header references and return
//'Table' struct. If something was wrong returns 'DBError'.
func Use(name, location string, sizes_rgx, header_rgx *regexp.Regexp) (*Table, error) {
	var err error
	var fd_data *os.File
	var fd_info *os.File

	var data_path string = fmt.Sprintf("%s%s/data", location, name)
	if fd_data, err = os.OpenFile(data_path, os.O_RDWR, os.ModePerm); err != nil {
		return nil, dberror.DBError{"Table not found."}
	}

	var info_path string = fmt.Sprintf("%s%s/info", location, name)
	if fd_info, err = os.OpenFile(info_path, os.O_RDWR, os.ModePerm); err != nil {
		return nil, dberror.DBError{"Table not found."}
	}
	defer fd_info.Close()

	var scanner *bufio.Scanner = bufio.NewScanner(fd_info)
	if !scanner.Scan() {
		return nil, dberror.DBError{"Bad formated table."}
	}

	var line_size int
	sizes_line := scanner.Text()
	res := sizes_rgx.FindString(sizes_line)
	if line_size, err = strconv.Atoi(res); err != nil {
		return nil, err
	}

	if !scanner.Scan() {
		return nil, dberror.DBError{"Bad formated table."}
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

	var index_path string = fmt.Sprintf("%s%s/index/", location, name)
	var table_index []*index.Index
	if table_index, err = index.Get(index_path); err != nil {
		return nil, err
	}

	var header *Header		= &Header{0, columns}
	var mutex *sync.Mutex	= &sync.Mutex{}

	if _, err = fd_info.Seek(0, 2); err != nil {
		return nil, dberror.DBError{"Error seeking table file descriptor."}
	}

	return &Table{name, location, header, line_size, fd_data, table_index, mutex}, nil
}

//Close database instance. Commit database changes and close file descriptor.
//If something was wrong returns 'DBError', else nil.
func (table *Table) Close() error {
	table.mutex.Lock()
	defer table.mutex.Unlock()

	if err := table.FileDescriptor.Sync(); err != nil {
		return dberror.DBError{"Error commiting table record."}
	}

	if table.FileDescriptor != nil {
		if err := table.FileDescriptor.Close(); err != nil {
			return dberror.DBError{"Error closing table"}
		}
	}

	return nil
}

//Create new record on table. First format new record and check if exists 
//indexable columns or key already exists on its index, then write new
//record on database row content and update index. If something was wrong 
//returns 'DBError', else nil.
func (table *Table) Add(row map[string]interface{}) error {
	var err error

	if len(row) != len(table.Header.Columns) {
		return dberror.DBError{"Wrong data to insert."}
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
				return dberror.DBError{"Error storing new record."}
			}
		}

		if t.Indexable {
			columns_to_index[column] = string(t.Content.([]byte))
		}
	}

	if len(columns_to_index) < 1 {
		return dberror.DBError{"No indexable columns provided"}
	}

	for column, content := range columns_to_index {
		for _, index := range table.Index {
			if index.Column == column {
				if index.Exist(content) {
					var message string = fmt.Sprintf("Constrain violation on '%s' column.", column)
					return dberror.DBError{message}
				} else {
					if err := index.Append(content); err != nil {
						return err
					}
				}
			}
		}
	}

	table.mutex.Lock()
	defer table.mutex.Unlock()

	var n int
	data = bff.Bytes()
	if n, err = table.FileDescriptor.Write(data); err != nil || n != table.LineSize {
		return dberror.DBError{"Error storing new record."}
	}

	if err = table.FileDescriptor.Sync(); err != nil {
		return dberror.DBError{"Error commiting new record."}
	}

	return nil
}

//Update row by record provided. One of columns must be indexed column at 
//least, the rest can be updated. Firs check if one of columns provided are
//indexable, then get record offset and limit, update it and store again in
//same position. If something was wrong returns 'DBError', else nil.
func (table *Table) Edit(row map[string]interface{}) error {
	var err error

	var key, indexed_key string
	var value interface{}

	OUTER:
	for _, index := range table.Index {
		for key, value = range row {
			if index.Column == key {
				indexed_key = key
				break OUTER
			}
		}
	}

	if len(indexed_key) < 1 {
		return dberror.DBError{"No primary key provided."}
	}

	var line_number int
	if line_number, err = table.getLine(indexed_key, value); err != nil {
		return err
	}

	table.mutex.Lock()
	defer table.mutex.Unlock()

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

	var row_offset int		= line_number * table.LineSize
	var row_end int			= row_offset + table.LineSize
	var row_content string	= content[row_offset:row_end]

	var col_offset int = 0
	for _, col := range columns {
		var data types.Type	= table.Header.Columns[col]
		var col_end int		= col_offset + data.Size

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
				return dberror.DBError{"Error storing new record."}
			}
		}
	}

	var offset int64 = int64(row_offset)
	if _, err = table.FileDescriptor.Seek(offset, 0); err != nil {
		return dberror.DBError{"Error deleting row. Rollback seek failed."}
	}

	data = bff.Bytes()
	var l int
	if l, err = table.FileDescriptor.Write(data); err != nil {
		return err
	} else if l != len(data) {
		if err = table.FileDescriptor.Truncate(0); err != nil {
			return dberror.DBError{"Error deleting row. Rollback truncate failed."}
		}

		if _, err = table.FileDescriptor.Seek(0, 0); err != nil {
			return dberror.DBError{"Error deleting row. Rollback seek failed."}
		}

		var n int
		if n, err = table.FileDescriptor.WriteString(content); err != nil || n != len(content) {
			return dberror.DBError{"Error deleting row. Rollback write failed."}
		}
	}

	if err = table.FileDescriptor.Sync(); err != nil {
		return dberror.DBError{"Error commiting new record."}
	}

	return nil
}

//Remove record from table by key value tuple. First get record line number
//on index, then join previous and next records of query result and store 
//on table. If something was worng return a 'DBError' with info message,
//else nil.
func (table *Table) Delete(key string, value interface{}) error {
	var err error

	var line_num int
	if line_num, err = table.getLine(key, value); err != nil {
		return err
	}

	table.mutex.Lock()
	defer table.mutex.Unlock()
	var content string
	if content, err = table.getContent(); err != nil {
		return err
	}

	var offset int			= line_num * table.LineSize + table.LineSize
	var after_line string	= content[offset:]

	var stat os.FileInfo
	var path string = fmt.Sprintf("%s%s/data", table.Location, table.Name)
	if stat, err = os.Stat(path); err != nil {
		return dberror.DBError{"Error reading table."}
	}

	var truncate_limit int64 = stat.Size() - int64(len(after_line) + table.LineSize)

	if err = table.FileDescriptor.Truncate(truncate_limit); err != nil {
		return dberror.DBError{"Error truncate table file descriptor."}
	}

	if _, err = table.FileDescriptor.Seek(0, 2); err != nil {
		return dberror.DBError{"Error seeking table descriptor.."}
	}

	var l int
	if l, err = table.FileDescriptor.WriteString(after_line); err != nil {
		return dberror.DBError{"Error deleting row."}
	} else if l != len(after_line) {
		if err = table.FileDescriptor.Truncate(0); err != nil {
			return dberror.DBError{"Error deleting row. Rollback truncate failed."}
		}

		if _, err = table.FileDescriptor.Seek(0, 0); err != nil {
			return dberror.DBError{"Error deleting row. Rollback seek failed."}
		}

		var n int
		if n, err = table.FileDescriptor.WriteString(content); err != nil || n != len(content) {
			return dberror.DBError{"Error deleting row. Rollback write failed."}
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

	return nil
}

//Get all table records formated. First obteins row table content and
//format each reacord by table definition structure. If something was wrong
//return a 'DBError' with info message, else list of 'Types'.
func (table *Table) Get() ([]map[string]types.Type, error) {
	var err error

	table.mutex.Lock()
	var content string
	if content, err = table.getContent(); err != nil {
		table.mutex.Unlock()
		return nil, err
	}
	table.mutex.Unlock()

	var columns []string
	for col := range table.Header.Columns {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	var results []map[string]types.Type
	var cursor int64		= 0
	var file_length int64	= int64(len(content) - 1)

	for cursor <= file_length {
		var row_end int64		= cursor + int64(table.LineSize)
		var row_content string	= content[cursor:row_end]
		cursor					= row_end

		var col_offset int				= 0
		var row map[string]types.Type	= make(map[string]types.Type, len(columns))
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

//Search and return single record from current table by key value tuple.
//First obtain record line number, then calculate offset and limit to search
//on raw database content and format record by table definition. If something 
//was wrong return a 'DBError' with info message, else 'Type'.
func (table *Table) GetOne(key string, value interface{}) (map[string]types.Type, error) {
	var err error

	var line_number int
	if line_number, err = table.getLine(key, value); err != nil {
		return nil, err
	}

	table.mutex.Lock()
	var content string
	if content, err = table.getContent(); err != nil {
		table.mutex.Unlock()
		return nil, err
	}
	table.mutex.Unlock()

	var columns []string
	for col := range table.Header.Columns {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	var result map[string]types.Type = make(map[string]types.Type, len(columns))

	var row_offset int		= line_number * table.LineSize
	var row_end int			= row_offset + table.LineSize
	var row_content string	= content[row_offset:row_end]

	var col_offset int = 0
	for _, col := range columns {
		var data types.Type	= table.Header.Columns[col]
		var col_end int		= col_offset + data.Size

		data.Content = row_content[col_offset:col_end]
		data.Decoder()

		result[col] = data
		col_offset += data.Size
	}

	return result, nil
}
