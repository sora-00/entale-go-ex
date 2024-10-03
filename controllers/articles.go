package controllers

import (
	"database/sql"
	"log"
)


type Article struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Body        string  `json:"body"`
	Medias      []Media `json:"medias"`
	PublishedAt string  `json:"publishedAt"`
}

type Media struct {
	ID          int    `json:"id"`
	ContentUrl  string `json:"contentUrl"`   
	ContentType string `json:"contentType"`  
}



func CreateTables(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		body TEXT,
		publishedAt TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS medias (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_id INTEGER,
		contentUrl TEXT,
		contentType TEXT,
		FOREIGN KEY(article_id) REFERENCES articles(id) 
	)`) 
	if err != nil {
		log.Fatal(err)
	}
}