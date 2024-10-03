package functions

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
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


func FetchArticles() ([]Article, error) { 
	res, err := http.Get("https://gist.githubusercontent.com/gotokatsuya/cc78c04d3af15ebe43afe5ad970bc334/raw/dc39bacb834105c81497ba08940be5432ed69848/articles.json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body) 
	if err != nil {
		return nil, err
	}

	var articles []Article 
	if err := json.Unmarshal(body, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

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

		stmt, err := db.Prepare("INSERT INTO medias(article_id, contentUrl, contentType) VALUES(?, ?, ?)") //53行目と同様
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

func GetArticlesHandler(db *sql.DB) http.HandlerFunc {
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