package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func GetArticlesFromDB(db *sql.DB) ([]Article, error) { 
	rows, err := db.Query("SELECT id, title, body, publishedAt FROM articles") 
	if err != nil {
		return nil, err
	}
	defer rows.Close() 

	var articles []Article 
	for rows.Next() {
		var article Article 
		if err := rows.Scan(&article.ID, &article.Title, &article.Body, &article.PublishedAt); err != nil { 
			return nil, err
		}

		mediaRows, err := db.Query("SELECT id, contentUrl, contentType FROM medias WHERE article_id = ?", article.ID)
		if err != nil {
			return nil, err
		}
		defer mediaRows.Close() 

		for mediaRows.Next() {
			var media Media 
			if err := mediaRows.Scan(&media.ID, &media.ContentUrl, &media.ContentType); err != nil { 
				return nil, err
			}
			article.Medias = append(article.Medias, media) 
		}
		articles = append(articles, article) 
	}
	return articles, nil
}

func GetArticles(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        articles, err := GetArticlesFromDB(db)
        if err != nil {
            http.Error(w, "Error retrieving articles", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(articles)
    }
}