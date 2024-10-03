package main

import (
	"database/sql"
	"log"

	"net/http"

	"entale-go-ex/functions"

	"github.com/go-chi/chi"
	_ "github.com/mattn/go-sqlite3"
)


func main() {
	db, err:= sql.Open("sqlite3", "./mydatabase.db") 
	if err != nil {
    log.Fatal("Failed to connect to database", err)
}
	defer db.Close()

	functions.CreateTables(db) 

	r := chi.NewRouter() 

	r.Get("/save", functions.SaveArticles(db))
	r.Get("/articles", functions.GetArticlesHandler(db))

	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Failed to startup server", err)
}
}


