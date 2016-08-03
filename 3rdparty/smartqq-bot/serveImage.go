package main

import "net/http"
import "io/ioutil"
import "os"
import "fmt"
var data []byte

func sendData(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type","image/png")
		w.Write(data)
}

func main() {
	if len(os.Args)!=3 {
		panic("Error: Bad arguments")
	}
	path := os.Args[1]
	listenAddr := os.Args[2]

	var err error
	data,err = ioutil.ReadFile(path)
	if err!=nil {
		panic(err)
	}
	http.HandleFunc("/",sendData)
	fmt.Println("Listening on",listenAddr)
	http.ListenAndServe(listenAddr,nil)
}
