package functions

import (
	"encoding/json"
	"io"
	"net/http"
)

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