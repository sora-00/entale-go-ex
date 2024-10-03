package controllers

import (
	"database/sql"
	"log"
	"net/http"
)

func SaveArticleToDB(db *sql.DB, article Article) error { 
	stmt, err := db.Prepare("INSERT INTO articles(title, body, publishedAt) VALUES(?, ?, ?)") 
	if err != nil {
		return err
	}
	defer stmt.Close() 

	res, err := stmt.Exec(article.Title, article.Body, article.PublishedAt) 
	if err != nil {
		return err
	}
	articleID, err := res.LastInsertId() 
	if err != nil {
		return err
	}


	for _, media := range article.Medias { 

		stmt, err := db.Prepare("INSERT INTO medias(article_id, contentUrl, contentType) VALUES(?, ?, ?)") 
		if err != nil {
			return err
		}
		_, err = stmt.Exec(articleID, media.ContentUrl, media.ContentType)
		if err != nil {
			return err
		}
	}

	return nil
}

func SaveArticles(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
	articles, err := FetchArticles()
	if err != nil {
		http.Error(w, "Failed to fetch articles", http.StatusInternalServerError)
		return
	
	}

	for _, article := range articles {
		if err := SaveArticleToDB(db, article); err != nil {
			log.Println("Error saving article:", err)
			http.Error(w, "Error saving articles", http.StatusInternalServerError)
			return
		}
	}

	w.Write([]byte("Articles saved successfully"))
}
}