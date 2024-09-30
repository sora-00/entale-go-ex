package main

import (
	"database/sql"  //sqlのimport
	"encoding/json" //goのデータ構造↔︎json形式　の変換を行うパッケージのimport
	"io"            //入出力を行うためのパッケージをimport
	"log"           //ログメッセージを出力するためのパッケージをimport
	"net/http"      //外部APIからデータを取得したり、HTTPサーバを起動したりするためのパッケージをimport

	"github.com/go-chi/chi"         //chiのimport
	_ "github.com/mattn/go-sqlite3" //SQLite3のドラーバーをinportしているが、初期化関数が呼び出されるだけで、直接的にパッケージの関数を使わない(database/sqlがどのデータベースに接続する必要があるか知るために必要)
)

//typeでgoプログラム内のメモリ上で使われるデータ型を定義できる(SQLのテーブルとは別物)
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

//データを取得して変換して返す
func fetchArticles() ([]Article, error) { //引数を取らずに呼び出すことができる 記事のスライス(リスト)とエラーを返す　fetch:DBの検索結果から一つ引き出すこと
	res, err := http.Get("https://gist.githubusercontent.com/gotokatsuya/cc78c04d3af15ebe43afe5ad970bc334/raw/dc39bacb834105c81497ba08940be5432ed69848/articles.json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //データが使用されなくなったら、リソースを解放するために閉じる　これを忘れると、メモリリークなどが発生する可能性がある　

	body, err := io.ReadAll(res.Body) //res.Bodyをすべて読み取って、bodyに格納。この時点で、レスポンスのデータ（ここではJSONデータ）がbodyに入る
	if err != nil {
		return nil, err
	}

	var articles []Article //取得したデータを格納するための変数を宣言
	if err := json.Unmarshal(body, &articles); err != nil { //bodyに入っているJSON形式のデータをarticlesというgoの構造体に変換
		return nil, err
	}

	return articles, nil //取得したデータを返す
}

//引数として渡されたArticle structのデータをarticlesテーブルに保存し、Mediaデータをmediasテーブルに挿入
func saveArticleToDB(db *sql.DB, article Article) error { //*sql.DB型はデータベースとの接続を表す Articleのようなstructも型として定義できる dbとarticleが引数でnilかerrorを返す
	stmt, err := db.Prepare("INSERT INTO articles(title, body, publishedAt) VALUES(?, ?, ?)") //stmtとerrをSQLのINSERT文で定義している(テーブルにデータを挿入) ?は仮 あとで入れる
	if err != nil {
		return err
	}
	defer stmt.Close() //上記と同様

	res, err := stmt.Exec(article.Title, article.Body, article.PublishedAt) //53行目の?に具体的な値を挿入
	if err != nil {
		return err
	}
	articleID, err := res.LastInsertId() //データベースに新しく挿入された記事のIDを取得→articleIDに入る
	if err != nil {
		return err
	}

	//メディアデータの挿入
	for _, media := range article.Medias { //ループ回数がいらない時は_ で省略できる 2つの変数を返す必要がなくてもrangeを使うと色々良いことが

		stmt, err := db.Prepare("INSERT INTO medias(article_id, contentUrl, contentType) VALUES(?, ?, ?)") //53行目と同様
		if err != nil {
			return err
		}
		_, err = stmt.Exec(articleID, media.ContentUrl, media.ContentType)//71行目の?に具体的な値を挿入
		if err != nil {
			return err
		}
	}

	return nil
}

//データベースから記事メディア情報を取得し、スライスに追加
func getArticlesFromDB(db *sql.DB) ([]Article, error) { //dbが引数 記事のスライスとerrorを返す
	rows, err := db.Query("SELECT id, title, body, publishedAt FROM articles") //rowsは取得したデータを順次処理するためのオブジェクトdb.Queryを使ってarticlesテーブルからid,title,body,publishedAtの列を取得
	if err != nil {
		return nil, err
	}
	defer rows.Close() //上記と同様

	var articles []Article //記事のスライスを用意
	for rows.Next() { //SQLクエリの結果セットを1行ずつ反復処理
		var article Article //Article型の変数articleを宣言
		if err := rows.Scan(&article.ID, &article.Title, &article.Body, &article.PublishedAt); err != nil { //rows.Scan()を使って、現在の行のデータ（記事情報）をarticleに読み込み
			return nil, err
		}

		// メディア情報を取得
		mediaRows, err := db.Query("SELECT id, contentUrl, contentType FROM medias WHERE article_id = ?", article.ID) //mediasテーブルから、article_idが article.IDに一致するメディアを取得
		if err != nil {
			return nil, err
		}
		defer mediaRows.Close() //上記と同様

		for mediaRows.Next() { //メディアの結果セットを1行ずつ反復処理
			var media Media //94行目と同様
			if err := mediaRows.Scan(&media.ID, &media.ContentUrl, &media.ContentType); err != nil { //95行目と同様
				return nil, err
			}
			article.Medias = append(article.Medias, media) //記事にメディアを追加
		}
		articles = append(articles, article) //完成したarticleをarticlesスライスに追加
	}
	return articles, nil
}

 //データベース内に２つのテーブルを作成
func createTables(db *sql.DB) {
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
	)`) //FOREIGN KEYでmedias テーブルの article_id と articles テーブルの id を紐付け
	if err != nil {
		log.Fatal(err)
	}
}

//上で定義した関数を使ってる
func main() {
	r := chi.NewRouter() //新規ルーターの作成(この特定のURLパスに対する処理の割り当てを行う)

	db, err := sql.Open("sqlite3", "./mydatabase.db") //ドライバとファイルパスを指定してSQLiteに接続し、dbを取得
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTables(db) //最初にターミナルで作ったからcreateTablesはいらないかと思ったけど、ここで呼び出すために必要

	articles, err := fetchArticles() //記事データの取得
	if err != nil {
		log.Fatal(err)
	}
	for _, article := range articles { //記事を一つずつ保存　articlesスライスの中にある記事全てを一つずつ取り出してarticleに保存
		if err := saveArticleToDB(db, article); err != nil { //記事をデータベースに保存
			log.Println("Error saving article:", err)
		}
	}

	r.Get("/articles", func(w http.ResponseWriter, r *http.Request) { ///articles パスに対する GET リクエストを処理するルートを定義
		articles, err := getArticlesFromDB(db) //データベースから記事データを取得
		if err != nil {
			http.Error(w, http.StatusText(500), 500) //サーバー内部エラー(ステータスコード500を返す)
			return
		}
		json.NewEncoder(w).Encode(articles) //取得した記事のデータをJSON形式にエンコードし、HTTPレスポンスとしてクライアントに返す
	})

	log.Fatal(http.ListenAndServe(":8080", r)) //サーバーを起動
}


