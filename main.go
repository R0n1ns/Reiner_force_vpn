package main

import (
	"Project/UX"
	"log"
	"net/http"
	"strings"
)

func notf(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
	return
}
func secureFileServer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		referer := r.Header.Get("Referer")
		// Если запрос не с вашего домена — блокируем
		if referer == "" || !strings.HasPrefix(referer, "http://127.0.0.1:4000") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	//подключаем файлы статики
	fileServer := http.FileServer(http.Dir("./UI/media/"))
	mux.Handle("/media/", secureFileServer(http.StripPrefix("/media", fileServer)))
	//заглушка,надо сделать норм
	mux.HandleFunc("/", notf)
	//главная страница
	mux.HandleFunc("/home", UX.Home)

	log.Println("Запуск веб-сервера на http://127.0.0.1:4000")
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
