package main

import (
	"./handler"
	"log"
	"net/http"
	"os"
)

// 因为和 ListError 有相同的接口类型
type appHandler func(writer http.ResponseWriter, request *http.Request) error

//type UploadContent interface {
//	Post(url string)string
//}

func errorWrapper(handler2 appHandler) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := handler2(writer, request)
		if err != nil {
			log.Printf("ERROR Occured %s", err)
			code := http.StatusOK
			switch {
			case os.IsNotExist(err):
				code = http.StatusNotFound
			case os.IsPermission(err):
				code = http.StatusUnauthorized

			default:
				code = http.StatusInternalServerError
			}
			http.Error(writer, http.StatusText(code), code)

		}
	}
}

func main() {
	http.HandleFunc("/list/", errorWrapper(handler.ListError))
	if err := http.ListenAndServe(":8888", nil); err != nil {
		//panic(err)
	}
}
