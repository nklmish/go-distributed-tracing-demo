package main

import (
	"github.com/gorilla/mux"
	"flag"
	"net/http"
	"log"
	"encoding/json"
	"fmt"
	"strconv"
)

var (
	port = flag.Int("port", 8090, "configure http port")
)

type Product struct {
	Name string `json:"name"`
}

func main() {
	flag.Parse()
	router := mux.NewRouter()
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/{id}", ProductHandler)
	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(*port), router))
}

func indexHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "Demo service, pass product id to get products")
}

func ProductHandler(res http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]

	res.Header().Set(
		"Content-Type",
		"application/json; charset=UTF-8",
	)

	products := []Product{
		{Name: fmt.Sprintf("Product A-%s", id)},
		{Name: fmt.Sprintf("Product B-%s", id)},
		{Name: fmt.Sprintf("Product C-%s", id)},
	}

	if err := json.NewEncoder(res).Encode(products); err != nil {
		panic(err)
	}
}
