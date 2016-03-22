package main

import "fmt"

import (
	"stupiddb/engine"
	"stupiddb/query"
)

func main() {
	//engine.CreateInstance("goomotic")
	db, err := engine.Instance("goomotic")
	if err != nil {
		fmt.Println(err)
	}

	/*fields := []string{
		engine.Unique("username"),
		"password",
	}

	if err := db.CreateTable("users", fields); err != nil {
		fmt.Println(err)
	}

	data := map[string]string{
		"username": "lucas",
		"password": "0301195",
	}

	q := query.New().Table("users").Data(data)
	if err := db.Add(q); err != nil {
		fmt.Println(err)
	}

	data = map[string]string{
		"username": "kazzzweb",
		"password": "2802049978",
	}

	q = query.New().Table("users").Data(data)
	if err := db.Add(q); err != nil {
		fmt.Println(err)
	}*/

	filter := map[string]string{
		"password": "0301195",
	}

	q := query.New().Table("users").Filters(filter)

	var result map[string]string
	result, err = db.GetOne(q)
	fmt.Println(result)
	fmt.Println(err)
}
