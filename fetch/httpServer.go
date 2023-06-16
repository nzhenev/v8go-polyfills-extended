package fetch

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Response struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func homePage(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Endpoint Hit: homePage")
	json.NewEncoder(w).Encode(&Response{
		Id:   1,
		Name: "home page",
	})
}
func publicPage(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Endpoint Hit: publicPage")
	json.NewEncoder(w).Encode(&Response{
		Id:   2,
		Name: "public page",
	})
}
func authPage(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Endpoint Hit: authPage")
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Un-Authorized request", http.StatusUnauthorized)
		return
	}
	if username == "" || password == "" {
		http.Error(w, "Invalid credentials", http.StatusBadRequest)
		return
	}
	if username == "test" && password == "test" {
		json.NewEncoder(w).Encode(&Response{
			Id:   3,
			Name: "auth sucess page",
		})
	} else {
		http.Error(w, "Un-Authorized request", http.StatusUnauthorized)
	}
}

func StartHttpServer(addr string) {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/public", publicPage)
	http.HandleFunc("/auth", authPage)
	log.Fatal(http.ListenAndServe(addr, nil))
}

/*
//create a dummy HTTP server for testing
	addr := "localhost:10000"
	go StartHttpServer(addr)
	es.WaitForHost(addr, 10)
*/
