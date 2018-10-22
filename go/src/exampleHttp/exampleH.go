package main

import (
    "fmt"
    "log"
    "net/http"
)
// this function takes an http.Resp.. and an http.Req.. 
// the first one Resp value assembles the http server's response 
// the second one Req is a DS that rappresent the client http req 
func handler(w http.ResponseWriter, r *http.Request) {
	//  r.URL.Path is the path component of the request URL
	// in this case we take all the Path without the first char which will be "/"
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	// give to the "/" a particular handlar which will handle the request on
	// this ruoute 
	http.HandleFunc("/", handler)
	// blocking function that should listen on port 8080 on any net interface 
	// the second parameter is used usually for middleware so now is nil
	// http.ListenAndServe return just an error so we wrap it to handle this possible err 
	log.Fatal(http.ListenAndServe(":8080", nil))
}