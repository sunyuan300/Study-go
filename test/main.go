package main

import (
	"encoding/json"
	"fmt"
	_ "net/http/pprof"
)

type Order struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	o := Order{
		ID:   "123",
		Name: "abc",
	}
	b, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
