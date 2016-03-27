package engine

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
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
		return err
	}

	path := user.HomeDir + "/.godb/"
	if _, err = os.Stat(path); err != nil {
		if err = os.Mkdir(path, os.ModePerm); err != nil {
			return err
		}
	}

	if _, err = os.Stat(path + database); err != nil {
		if err = os.Mkdir(path+database, os.ModePerm); err != nil {
			return err
		}
		return nil
	} else {
		return err
	}
}

//Return instance with its attributes
func Instance(schema string) (*engine, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	path := user.HomeDir + "/.godb/"

	if _, err := os.Stat(path + schema); err != nil {
		return nil, err
	}

	location := path + schema

	var database []os.FileInfo
	if database, err = ioutil.ReadDir(location); err != nil {
		return nil, err
	}

	return &engine{database, location + "/"}, nil
}

//Create table
func (db *engine) CreateTable(table string, fields []string) error {
	fd, err := os.Create(db.location + table)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.Write([]byte(strings.Join(fields, "<;;>") + "<;;>\n"))
	return err
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
		_, err = fd.Write([]byte(data + "\n"))
		return err
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

	old := strings.Join(content, "<;;>")
	for i, hd := range headers {
		for key, value := range q.d {
			if hd == key {
				content[i] = value
			}
		}
	}

	new := []byte(strings.Join(content, "<;;>"))

	var data []byte
	if data, err = ioutil.ReadFile(db.location + q.t); err != nil {
		return err
	}

	data = bytes.Replace(data, []byte(old), new, -1)

	if err = ioutil.WriteFile(db.location+q.t, data, os.ModePerm); err != nil {
		return err
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

	old := strings.Join(content, "<;;>") + "\n"
	for i, hd := range headers {
		for key, value := range q.d {
			if hd == key {
				content[i] = value
			}
		}
	}

	var data []byte
	if data, err = ioutil.ReadFile(db.location + q.t); err != nil {
		return err
	}

	data = bytes.Replace(data, []byte(old), []byte(""), -1)

	if err = ioutil.WriteFile(db.location+q.t, data, os.ModePerm); err != nil {
		return err
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
