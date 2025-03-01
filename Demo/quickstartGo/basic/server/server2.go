package main

import (
	"log"
	"net/http"
	"net/url"
)

var userCount = make(map[string]int64, 0)

func main() {

	//http.HandleFunc("/", HttpHandle)
	// 启动服务器
	err := http.ListenAndServe(":20520", test{})
	if err != nil {
		panic(err)
	}
}

type test struct {
}

func (t test) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if URL, err := url.ParseRequestURI(request.RequestURI); err == nil {
		if query, err1 := url.ParseQuery(URL.RawQuery); err1 == nil {
			username := query.Get("username")
			if count, ok := userCount[username]; ok {
				count++
				userCount[username] = count
				log.Print(count)
			} else {
				userCount[username] = 1
			}
		}
	}
}

func (test) HttpHandle(w http.ResponseWriter, r *http.Request) {

}
