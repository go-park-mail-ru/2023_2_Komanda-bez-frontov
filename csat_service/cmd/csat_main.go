package main

import (
 "fmt"
 "log"
 "net/http"
 "go-form-hub/csat_service/internal"
 "go-form-hub/csat_service/db"
 "go-form-hub/csat_service/repository"
)

type Response struct {
 Data string `json:"data"`
}

func main() {

	database, err := db.NewDB()
	if err != nil {
		fmt.Println(err)
	}
	defer database.Close()

	csat_repository := repository.NewCSATService(database)

	csathandler := api.NewCSATHandler(csat_repository)

	// Регистрируем обработчики для путей /api/v1/csat/check и /api/v1/csat/add
	http.HandleFunc("/api/v1/csat/check", csathandler.CheckCSAT)
	http.HandleFunc("/api/v1/csat/add", csathandler.AddCSAT)
	http.HandleFunc("/api/v1/csat/results", csathandler.ResultsCSAT)
   
	// Настройка и запуск сервера
	server := &http.Server{
	 Addr:    ":8090",           // Слушаем порт 8090
	 Handler: http.DefaultServeMux,
	}
   
	fmt.Println("Сервер запущен на порту 8090")
	log.Fatal(server.ListenAndServe())
}