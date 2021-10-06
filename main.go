package main

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
)

type dogdogdog int

func (m dogdogdog) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatalln(err)
	}

	data := struct {
		Method      string
		Submissions url.Values
	}{
		req.Method,
		req.Form,
	}
	tpl.ExecuteTemplate(w, "index.gohtml", data)
}

var tpl *template.Template = template.Must(template.ParseFiles("app/views/templates/index.gohtml"))

func main() {
	var d dogdogdog
	http.ListenAndServe("localhost:8080", d)
}
