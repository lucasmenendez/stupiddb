package stupiddb

import (
	//"bufio"
	//"bytes"
	"fmt"
	"os"
	"os/user"
	//"strings"

	//"stupiddb/query"
	"stupiddb/types"
)

type DBError struct {
	Message string
}

func (err DBError) Error() string {
	return fmt.Sprintf("DBError: %v", err.Message)
}


type engine struct {
	database []os.FileInfo
	location string
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


func (db *engine) CreateTable(table string, fields map[string]types.Type) error {
	fmt.Println(db.location + table)
	fd, err := os.Create(db.location + table)
	if err != nil {
		return DBError{"Error creating database file."}
	}
	defer fd.Close()

	var header string
	var line_length int
	for column, t := range fields {
		header += fmt.Sprintf("%s(%d)%s;", t.Alias, t.Size, column)
		line_length += t.Size
	}
	header = fmt.Sprintf("%d;%d\n%s", len(header), line_length, header)

	if _, err = fd.Write([]byte(header)); err != nil {
		return DBError{"Error creating database struct."}
	}
	return nil
}

//func (db *engine) Add(query *query.Query) error {
//	fd, err := os.OpenFile(db.location+query.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
//	if err != nil {
//		return DBError{"Table not found."}
//	}
//
//	scanner := bufio.NewScanner(fd)
//	if !scanner.Scan() {
//		return DBError{"Table not found."}
//	}
//
//	header := strings.Split(scanner.Text(), "<;;>")
//	headers := header[:len(header)-1]
//
//	defer fd.Close()
//
//	if len(headers) == len(query.Data) {
//		for scanner.Scan() {
//			fields := strings.Split(scanner.Text(), "<;;>")
//			for i, hd := range headers {
//				if hd[:2] == "U%" && query.Data[hd[2:]] == fields[i] {
//					return DBError{"Unique row '" + hd + "' restriction violated."}
//				}
//			}
//		}
//
//		var data string
//		for _, hd := range headers {
//			isField := false
//			for column, value := range query.Data {
//				if hd[:2] == "U%"c{
//					hd = hd[2:]
//				}
//
//				if column == hd {
//					data += value + "<;;>"
//					isField = true
//				} else {
//					isField = isField || false
//				}
//			}
//			if isField != true {
//				return DBError{"Table struct and data attributes doesn't match."}
//			}
//
//		}
//		if _, err = fd.Write([]byte(data + "\n")); err != nil {
//			return DBError{"Error writting table."}
//		}
//		return nil
//	}
//
//	return DBError{"Attribute length mismatch."}
//}
//
//func (db *engine) Edit(q *query.Query) error {
//	fd, err := os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
//	if err != nil {
//		return DBError{"Table not found."}
//	}
//
//	scanner := bufio.NewScanner(fd)
//	if !scanner.Scan() {
//		return DBError{"Table not found."}
//	}
//
//	header := strings.Split(scanner.Text(), "<;;>")
//	headers := header[:len(header)-1]
//
//	filters := make(map[int]string)
//	for i, hd := range headers {
//		header := hd
//		if header[:2] == "U%" {
//			header = header[2:]
//		}
//
//		for filter, value := range q.f {
//			if header == filter {
//				filters[i] = value
//			}
//		}
//	}
//
//	var content []string
//	for scanner.Scan() {
//		content = strings.Split(scanner.Text(), "<;;>")
//		match := true
//		for pos, value := range filters {
//			match = value == content[pos]
//			if !match {
//				break
//			}
//		}
//		if match {
//			break
//		}
//	}
//
//	fd.Close()
//
//	old := strings.Join(content, "<;;>")
//	for i, hd := range headers {
//		for key, value := range q.d {
//			if hd[:2] == "U%" && hd[2:] == key {
//				return DBError{"Unique column cannot be modificated."}
//			}
//			if hd == key {
//				content[i] = value
//			}
//		}
//	}
//
//	new := []byte(strings.Join(content, "<;;>"))
//
//	fd, err = os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
//	if err != nil {
//		return DBError{"Error opening table."}
//	}
//
//	fileinfo, _ := fd.Stat()
//	data := make([]byte, fileinfo.Size())
//	if _, err = fd.Read(data); err != nil {
//		return DBError{"Error reading table."}
//	}
//
//	data = bytes.Replace(data, []byte(old), new, -1)
//	if err = fd.Truncate(0); err != nil {
//		return DBError{"Error truncating table."}
//	}
//
//	defer fd.Close()
//
//	if _, err = fd.Write(data); err != nil {
//		return DBError{"Error writting table."}
//	}
//	return nil
//}
//
//func (db *engine) Delete(q *query.Query) error {
//	fd, err := os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
//	if err != nil {
//		return DBError{"Table not found."}
//	}
//
//	scanner := bufio.NewScanner(fd)
//	if !scanner.Scan() {
//		return DBError{"Table not found."}
//	}
//
//	header := strings.Split(scanner.Text(), "<;;>")
//	headers := header[:len(header)-1]
//
//	filters := make(map[int]string)
//	for i, hd := range headers {
//		if hd[:2] == "U%" {
//			hd = hd[2:]
//		}
//
//		for filter, value := range q.f {
//			if hd == filter {
//				filters[i] = value
//			}
//		}
//	}
//
//	var content []string
//	for scanner.Scan() {
//		content = strings.Split(scanner.Text(), "<;;>")
//		match := true
//		for pos, value := range filters {
//			match = value == content[pos]
//			if !match {
//				break
//			}
//		}
//		if match {
//			break
//		}
//	}
//
//	fd.Close()
//
//	old := strings.Join(content, "<;;>") + "\n"
//	for i, hd := range headers {
//		for key, value := range q.d {
//			if hd == key {
//				content[i] = value
//			}
//		}
//	}
//
//	fd, err = os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
//	if err != nil {
//		return DBError{"Error opening table."}
//	}
//
//	fileinfo, _ := fd.Stat()
//	data := make([]byte, fileinfo.Size())
//	if _, err = fd.Read(data); err != nil {
//		return DBError{"Error reading table."}
//	}
//
//	data = bytes.Replace(data, []byte(old), []byte(""), -1)
//	if err = fd.Truncate(0); err != nil {
//		return DBError{"Error truncating table."}
//	}
//
//	defer fd.Close()
//
//	if _, err = fd.Write(data); err != nil {
//		return DBError{"Error writting table."}
//	}
//	return nil
//}
//
//func (db *engine) Get(q *query.Query) ([]map[string]string, error) {
//	fd, err := os.OpenFile(db.location+q.t, os.O_RDONLY, os.ModePerm)
//	if err != nil {
//		return nil, DBError{"Table not found."}
//	}
//
//	scanner := bufio.NewScanner(fd)
//	if !scanner.Scan() {
//		return nil, DBError{"Table not found."}
//	}
//
//	header := strings.Split(scanner.Text(), "<;;>")
//	headers := header[:len(header)-1]
//
//	defer fd.Close()
//
//	filters := make(map[int]string)
//	for i, hd := range headers {
//		if hd[:2] == "U%" {
//			hd = hd[2:]
//		}
//
//		for filter, value := range q.f {
//			if hd == filter {
//				filters[i] = value
//			}
//		}
//	}
//
//	var result []map[string]string
//	for scanner.Scan() {
//		content := strings.Split(scanner.Text(), "<;;>")
//		match := true
//		for pos, value := range filters {
//			match = value == content[pos]
//			if !match {
//				break
//			}
//		}
//		if match {
//			item := make(map[string]string)
//			for i, hd := range headers {
//				if hd[:2] == "U%" {
//					hd = hd[2:]
//				}
//				item[hd] = content[i]
//			}
//			result = append(result, item)
//		}
//	}
//
//	err = DBError{"No result."}
//	if len(result) > 0 {
//		err = nil
//	}
//	return result, err
//
//}
//
//func (db *engine) GetOne(q *query.Query) (map[string]string, error) {
//	fd, err := os.OpenFile(db.location+q.t, os.O_RDONLY, os.ModePerm)
//	if err != nil {
//		return nil, DBError{"Table not found."}
//	}
//
//	scanner := bufio.NewScanner(fd)
//	if !scanner.Scan() {
//		return nil, DBError{"Table not found."}
//	}
//
//	header := strings.Split(scanner.Text(), "<;;>")
//	headers := header[:len(header)-1]
//
//	defer fd.Close()
//
//	filters := make(map[int]string)
//	for i, hd := range headers {
//		if hd[:2] == "U%" {
//			hd = hd[2:]
//		}
//
//		for filter, value := range q.f {
//			if hd == filter {
//				filters[i] = value
//			}
//		}
//	}
//
//	for scanner.Scan() {
//		content := strings.Split(scanner.Text(), "<;;>")
//		match := true
//		for pos, value := range filters {
//			match = value == content[pos]
//			if !match {
//				break
//			}
//		}
//		if match {
//			item := make(map[string]string)
//			for i, hd := range headers {
//				if hd[:2] == "U%" {
//					hd = hd[2:]
//				}
//				item[hd] = content[i]
//			}
//			return item, nil
//		}
//	}
//
//	return nil, DBError{"No result."}
//}
