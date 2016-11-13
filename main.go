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
	if err := json.NewEncoder(res).Encode(id); err != nil {
		panic(err)
	}
}
