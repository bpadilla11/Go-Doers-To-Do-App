package main

import (
	"net/http"
	"html/template"
)


func index(res http.ResponseWriter, request *http.Request)  {

	tpl, _ := template.ParseFiles("index.html")
	tpl.Execute(res, nil)
}

func main()  {
	http.HandleFunc("/", index)
	http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("css"))))
	http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("js"))))
	http.ListenAndServe(":8080", nil)


}
