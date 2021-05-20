// Copyright 2021 Matus Mrekaj. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	r := mux.NewRouter()

	r.HandleFunc("/", parseHtml()).Methods(http.MethodPost)

	log.Printf("listening on port: 8080")
	return http.ListenAndServe(":8080", r)
}
