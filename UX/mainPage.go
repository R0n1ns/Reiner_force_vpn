package UX

import (
	"Project/db"
	"html/template"
	"log"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	//получаем страницу
	ts, err := template.ParseFiles("./UI/mainPage.gohtml")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
	products := db.Getproducts()
	//отправляем страницу
	err = ts.Execute(w, *products)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}
