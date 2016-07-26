package index

import (
	"os"
	"fmt"
	"bufio"
	"io/ioutil"
)

type DBError struct {
	Message string
}

func (err DBError) Error() string {
	return fmt.Sprintf("DBError: %v", err.Message)
}


type Index struct {
	Column string
	Location string
	Content []string
}

func New(location, table, column string) error {
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

func (index *Index) Remove() error {
	var index_location = index.Location + index.Column
	if err := os.RemoveAll(index_location); err != nil {
		return DBError{"Error deleting index."}
	}

	return nil
}

func Get(index_path string) ([]*Index, error) {
	var index []*Index
	var err error

	if _, err = os.Stat(index_path); err != nil {
		return index, DBError{"Bad formated table: no index. Create it againg."} //Should not be here ever
	}

	var index_files []os.FileInfo
	if index_files, err = ioutil.ReadDir(index_path); err != nil {
		return index, DBError{"Error getting table index."}
	}

	for _, file := range index_files {
		var index_name string = file.Name()
		var index_fd *os.File
		var index_content []string
		var err error

		if index_fd, err = os.Open(index_path + index_name); err != nil {
			return index, DBError{"Error opening index."}
		}
		defer index_fd.Close()

		scanner := bufio.NewScanner(index_fd)
		for scanner.Scan() {
			index_content = append(index_content, scanner.Text())
		}


		index = append(index, &Index{index_name, index_path, index_content})
	}

	return index, nil
}

func (index *Index) Exist(needle string) bool {
	for _, value := range index.Content {
		if value == needle {
			return true
		}
	}

	return false
}

func (index *Index) Find(needle string) (int, error) {
	for line_number, value := range index.Content {
		if value == needle {
			return line_number, nil
		}
	}

	return 0, DBError{"Key not found on index."}
}

func (index *Index) Append (content string) error {
	var index_location = index.Location + index.Column

	if fd, err := os.OpenFile(index_location, os.O_WRONLY | os.O_APPEND, os.ModeAppend); err != nil {
		return DBError{"Error indexing row."}
	} else {
		var l int
		var record = fmt.Sprintf("%s\n", content)

		if l, err = fd.Write([]byte(record)); err != nil {
			return DBError{"Error writting in index."}
		} else if l != len(record) {
			//TODO: Delete last index line
		} else {
			index.Content = append(index.Content, content)
		}

		fd.Close()
	}

	return nil
}
