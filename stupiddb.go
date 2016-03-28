package stupiddb

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/user"
	"strings"
)

//Custom error
type DBError struct {
	Message string
}

func (err DBError) Error() string {
	return fmt.Sprintf("%v", err.Message)
}

//Query struct
type query struct {
	t string
	f map[string]string
	d map[string]string
}

func Query() *query {
	return &query{}
}

func (q *query) Table(name string) *query {
	q.t = name
	return q
}

func (q *query) Data(data map[string]string) *query {
	q.d = data
	return q
}

func (q *query) Filters(filters map[string]string) *query {
	q.f = filters
	return q
}

//Engine struct
type engine struct {
	database []os.FileInfo
	location string
}

//Create database file on filesystem
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

//Return instance with its attributes
func Instance(schema string) (*engine, error) {
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

	var database []os.FileInfo
	if database, err = fd.Readdir(-1); err != nil {
		return nil, DBError{"Error reading table headers."}
	}

	return &engine{database, location + "/"}, nil
}

//Create table
func (db *engine) CreateTable(table string, fields []string) error {
	fd, err := os.Create(db.location + table)
	if err != nil {
		return DBError{"Error creating database file."}
	}
	defer fd.Close()

	if _, err = fd.Write([]byte(strings.Join(fields, "<;;>") + "<;;>\n")); err != nil {
		return DBError{"Error creating database struct."}
	}
	return nil
}

//String mask to set unique restriction to column name
func Unique(data string) string {
	return "U%" + data
}

//Add row with query information
func (db *engine) Add(q *query) error {
	fd, err := os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return DBError{"Table not found."}
	}

	scanner := bufio.NewScanner(fd)
	if !scanner.Scan() {
		return DBError{"Table not found."}
	}

	header := strings.Split(scanner.Text(), "<;;>")
	headers := header[:len(header)-1]

	defer fd.Close()

	if len(headers) == len(q.d) {
		for scanner.Scan() {
			fields := strings.Split(scanner.Text(), "<;;>")
			for i, hd := range headers {
				if hd[:2] == "U%" && q.d[hd[2:]] == fields[i] {
					return DBError{"Unique row '" + hd + "' restriction violated."}
				}
			}
		}

		var data string
		for _, hd := range headers {
			isField := false
			for column, value := range q.d {
				if hd[:2] == "U%" {
					hd = hd[2:]
				}

				if column == hd {
					data += value + "<;;>"
					isField = true
				} else {
					isField = isField || false
				}
			}
			if isField != true {
				return DBError{"Table struct and data attributes doesn't match."}
			}

		}
		if _, err = fd.Write([]byte(data + "\n")); err != nil {
			return DBError{"Error writting table."}
		}
		return nil
	}

	return DBError{"Attribute length mismatch."}
}

//Edit, if exists, the row specified on query with its replacement
func (db *engine) Edit(q *query) error {
	fd, err := os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return DBError{"Table not found."}
	}

	scanner := bufio.NewScanner(fd)
	if !scanner.Scan() {
		return DBError{"Table not found."}
	}

	header := strings.Split(scanner.Text(), "<;;>")
	headers := header[:len(header)-1]

	filters := make(map[int]string)
	for i, hd := range headers {
		header := hd
		if header[:2] == "U%" {
			header = header[2:]
		}

		for filter, value := range q.f {
			if header == filter {
				filters[i] = value
			}
		}
	}

	var content []string
	for scanner.Scan() {
		content = strings.Split(scanner.Text(), "<;;>")
		match := true
		for pos, value := range filters {
			match = value == content[pos]
			if !match {
				break
			}
		}
		if match {
			break
		}
	}

	fd.Close()

	old := strings.Join(content, "<;;>")
	for i, hd := range headers {
		for key, value := range q.d {
			if hd[:2] == "U%" && hd[2:] == key {
				return DBError{"Unique column cannot be modificated."}
			}
			if hd == key {
				content[i] = value
			}
		}
	}

	new := []byte(strings.Join(content, "<;;>"))

	fd, err = os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return DBError{"Error opening table."}
	}

	fileinfo, _ := fd.Stat()
	data := make([]byte, fileinfo.Size())
	if _, err = fd.Read(data); err != nil {
		return DBError{"Error reading table."}
	}

	data = bytes.Replace(data, []byte(old), new, -1)
	if err = fd.Truncate(0); err != nil {
		return DBError{"Error truncating table."}
	}

	defer fd.Close()

	if _, err = fd.Write(data); err != nil {
		return DBError{"Error writting table."}
	}
	return nil
}

//Delete a row according to query data
func (db *engine) Delete(q *query) error {
	fd, err := os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return DBError{"Table not found."}
	}

	scanner := bufio.NewScanner(fd)
	if !scanner.Scan() {
		return DBError{"Table not found."}
	}

	header := strings.Split(scanner.Text(), "<;;>")
	headers := header[:len(header)-1]

	filters := make(map[int]string)
	for i, hd := range headers {
		if hd[:2] == "U%" {
			hd = hd[2:]
		}

		for filter, value := range q.f {
			if hd == filter {
				filters[i] = value
			}
		}
	}

	var content []string
	for scanner.Scan() {
		content = strings.Split(scanner.Text(), "<;;>")
		match := true
		for pos, value := range filters {
			match = value == content[pos]
			if !match {
				break
			}
		}
		if match {
			break
		}
	}

	fd.Close()

	old := strings.Join(content, "<;;>") + "\n"
	for i, hd := range headers {
		for key, value := range q.d {
			if hd == key {
				content[i] = value
			}
		}
	}

	fd, err = os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return DBError{"Error opening table."}
	}

	fileinfo, _ := fd.Stat()
	data := make([]byte, fileinfo.Size())
	if _, err = fd.Read(data); err != nil {
		return DBError{"Error reading table."}
	}

	data = bytes.Replace(data, []byte(old), []byte(""), -1)
	if err = fd.Truncate(0); err != nil {
		return DBError{"Error truncating table."}
	}

	defer fd.Close()

	if _, err = fd.Write(data); err != nil {
		return DBError{"Error writting table."}
	}
	return nil
}

//Get all rows requested on query data
func (db *engine) Get(q *query) ([]map[string]string, error) {
	fd, err := os.OpenFile(db.location+q.t, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, DBError{"Table not found."}
	}

	scanner := bufio.NewScanner(fd)
	if !scanner.Scan() {
		return nil, DBError{"Table not found."}
	}

	header := strings.Split(scanner.Text(), "<;;>")
	headers := header[:len(header)-1]

	defer fd.Close()

	filters := make(map[int]string)
	for i, hd := range headers {
		if hd[:2] == "U%" {
			hd = hd[2:]
		}

		for filter, value := range q.f {
			if hd == filter {
				filters[i] = value
			}
		}
	}

	var result []map[string]string
	for scanner.Scan() {
		content := strings.Split(scanner.Text(), "<;;>")
		match := true
		for pos, value := range filters {
			match = value == content[pos]
			if !match {
				break
			}
		}
		if match {
			item := make(map[string]string)
			for i, hd := range headers {
				if hd[:2] == "U%" {
					hd = hd[2:]
				}
				item[hd] = content[i]
			}
			result = append(result, item)
		}
	}

	err = DBError{"No result."}
	if len(result) > 0 {
		err = nil
	}
	return result, err

}

//Get the first row with information of query data
func (db *engine) GetOne(q *query) (map[string]string, error) {
	fd, err := os.OpenFile(db.location+q.t, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, DBError{"Table not found."}
	}

	scanner := bufio.NewScanner(fd)
	if !scanner.Scan() {
		return nil, DBError{"Table not found."}
	}

	header := strings.Split(scanner.Text(), "<;;>")
	headers := header[:len(header)-1]

	defer fd.Close()

	filters := make(map[int]string)
	for i, hd := range headers {
		if hd[:2] == "U%" {
			hd = hd[2:]
		}

		for filter, value := range q.f {
			if hd == filter {
				filters[i] = value
			}
		}
	}

	for scanner.Scan() {
		content := strings.Split(scanner.Text(), "<;;>")
		match := true
		for pos, value := range filters {
			match = value == content[pos]
			if !match {
				break
			}
		}
		if match {
			item := make(map[string]string)
			for i, hd := range headers {
				if hd[:2] == "U%" {
					hd = hd[2:]
				}
				item[hd] = content[i]
			}
			return item, nil
		}
	}

	return nil, DBError{"No result."}
}
