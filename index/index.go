package index

import (
	"os"
	"fmt"
	"sync"
	"bufio"
	"io/ioutil"

	"stupiddb/dberror"
)

//Define index struct atributtes that contains column name, location path,
//content and its mutex.
type Index struct {
	Column string
	Location string
	Content []string
	Mutex *sync.Mutex
}

//Create new index for table column provided on its location. If this index its
//the first, create folder tree to store its table index.
//If something fails returns a 'DBError' with info message.
func New(location, table, column string) error {
	var err error
	var index_path string = location + table + "/index"
	if _, err = os.Stat(index_path); err != nil {
		if err = os.Mkdir(index_path, os.ModePerm); err != nil {
			return dberror.DBError{"Error creating table index container."}
		}
	}

	var fd *os.File
	if fd, err = os.Create(index_path + "/" + column); err != nil {
		return dberror.DBError{"Error creating column index."}
	}
	fd.Close()

	return nil
}

//Remove column index file if exists. If something fails returns a 'DBError' with 
//info message.
func (index *Index) Remove() error {
	var index_location = index.Location + index.Column
	if err := os.RemoveAll(index_location); err != nil {
		return dberror.DBError{"Error deleting index."}
	}

	return nil
}

//Returns a Index's pointer from table location provided. First check if table 
//has any index and then returns all index found with its attributues and content.
//If something fails returns a 'DBError' with info message.
func Get(index_path string) ([]*Index, error) {
	var index []*Index
	var err error

	if _, err = os.Stat(index_path); err != nil { //Should not be here never
		return index, dberror.DBError{"Bad formated table: no index. Create it againg."}
	}

	var index_files []os.FileInfo
	if index_files, err = ioutil.ReadDir(index_path); err != nil {
		return index, dberror.DBError{"Error getting table index."}
	}

	for _, file := range index_files {
		var index_name string = file.Name()
		var index_fd *os.File
		var index_content []string
		var err error

		if index_fd, err = os.Open(index_path + index_name); err != nil {
			return index, dberror.DBError{"Error opening index."}
		}
		defer index_fd.Close()

		scanner := bufio.NewScanner(index_fd)
		for scanner.Scan() {
			index_content = append(index_content, scanner.Text())
		}

		var Mutex *sync.Mutex = &sync.Mutex{}
		index = append(index, &Index{index_name, index_path, index_content, Mutex})
	}

	return index, nil
}

//Check if record exists on selected index by content provided encoded previously.
//Return true in record exists.
func (index *Index) Exist(needle string) bool {
	index.Mutex.Lock()
	defer index.Mutex.Unlock()

	for _, value := range index.Content {
		if value == needle {
			return true
		}
	}

	return false
}

//Returns record line number location on selected index by encoded content. If
//something fails, returns a 'DBError' with info message.
func (index *Index) Find(needle string) (int, error) {
	index.Mutex.Lock()
	defer index.Mutex.Unlock()

	for line_number, value := range index.Content {
		if value == needle {
			return line_number, nil
		}
	}

	return 0, dberror.DBError{"Key not found on index."}
}

//Insert new record to selected index on temp memory and commit the change on
//associated file. Lock file while commiting the change. If something fails 
//returns a 'DBError' with info message.
func (index *Index) Append(content string) error {
	var err error
	var index_location = index.Location + index.Column

	index.Mutex.Lock()
	defer index.Mutex.Unlock()

	var stat os.FileInfo
	if stat, err = os.Stat(index_location); err != nil {
		return dberror.DBError{"Error reading index."}
	}

	var old_size int64 = stat.Size()

	var fd *os.File
	if fd, err = os.OpenFile(index_location, os.O_WRONLY | os.O_APPEND, os.ModeAppend); err != nil {
		return dberror.DBError{"Error indexing row."}
	} else {
		var l int
		var record = fmt.Sprintf("%s\n", content)

		if l, err = fd.Write([]byte(record)); err != nil {
			return dberror.DBError{"Indexed error."}
		} else if l != len(record) {
			if err = fd.Truncate(old_size); err != nil {
				return dberror.DBError{"Indexed error. Rollback failed."}
			} else {
				return dberror.DBError{"Indexed error. Rollback done."}
			}
		} else {
			index.Content = append(index.Content, content)
		}

		fd.Close()
	}

	return nil
}

//Delete record from temp memory and form index file. Lock file while committing 
//the change. If something fails returns a 'DBError' with info message.
func (index *Index) Delete(content string) error {
	var err error
	var index_location = index.Location + index.Column

	index.Mutex.Lock()
	defer index.Mutex.Unlock()

	var index_length int = len(index.Content)
	if index.Content[index_length - 1] != content {
		var flag int = index_length
		for i := 0; i < index_length - 1; i++ {
			if index.Content[i] == content {
				flag = i
			}

			if i >= flag {
				index.Content[i] = index.Content[i + 1]
			}
		}
	}
	index.Content = index.Content[:index_length - 1]

	var fd *os.File
	if fd, err = os.OpenFile(index_location, os.O_WRONLY | os.O_APPEND, os.ModeAppend); err != nil {
		return dberror.DBError{"Error indexing row."}
	} else {
		var content string
		for i := range index.Content {
			content = fmt.Sprintf("%s%s\n", content, index.Content[i])
		}

		var l int
		if l, err = fd.WriteString(content); err != nil || l != len(content) {
			return dberror.DBError{"Indexed error."}
		}

		fd.Close()
	}

	return nil
}
