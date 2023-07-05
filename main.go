package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"

	db "github.com/sonyarouje/simdb"

	"blog-assignment/pkg"
)

type handler struct {
	db *db.Driver
	t  *template.Template
}

type TemplateData struct {
	Articles []pkg.Article
	Metrics  pkg.DeleteCommentData
}

func main() {
	driver, err := db.New("blog")
	if err != nil {
		log.Fatal(err)
	}

	t := template.New("t").Funcs(template.FuncMap{
		"jsCallOp2": func(fnName, field1, field2 string) template.JS {
			return template.JS(fmt.Sprintf(`%s("%s","%s")`, fnName, field1, field2))
		},
		"isDeleted": isDeleted,
	})

	t, err = t.ParseGlob("./pkg/*.html")
	if err != nil {
		log.Fatal(err)
	}

	h := handler{driver, t}

	http.HandleFunc("/", h.handleIndex)
	http.HandleFunc("/article", h.handleArticle)
	http.HandleFunc("/comment", h.handleComment)
	http.HandleFunc("/hello", h.handleHello)

	http.ListenAndServe(":8081", nil)
}

func isDeleted(deleted time.Time) bool {
	return !deleted.Equal(time.Time{})
}

func (h *handler) handleHello(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func (h *handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// List all articles
		articles, err := pkg.GetArticles(h.db)
		if err != nil && err != db.ErrRecordNotFound {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Comment deletion metrics
		var metrics pkg.DeleteCommentData
		for _, art := range articles {
			for _, comment := range art.Comments {
				if isDeleted(comment.DeletedAt) {
					metrics.Data = append(metrics.Data, pkg.DeleteEntry{
						CreatedAt: comment.CreatedAt,
						DeletedAt: comment.DeletedAt,
					})
				}
			}
		}

		// Render the template
		if err := h.t.ExecuteTemplate(w, "index.html", TemplateData{Articles: articles, Metrics: metrics}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *handler) handleArticle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		params, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		art, err := pkg.GetArticle(params.Get("id"), h.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// render the article page
		if err := h.t.ExecuteTemplate(w, "article.html", art); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "POST":
		// Create a new article
		r.ParseForm()
		title := r.FormValue("title")
		author := r.FormValue("author")
		content := r.FormValue("content")

		art := pkg.Article{
			Title:   title,
			Author:  author,
			Content: content,
		}

		if _, err := pkg.CreateArticle(art, h.db); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *handler) handleComment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Create a new comment
		params, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r.ParseForm()
		comment := pkg.Comment{
			Name:      r.FormValue("name"),
			Comment:   r.FormValue("comment"),
			CreatedAt: time.Now(),
		}

		articleID := params.Get("article_id")

		art, err := pkg.GetArticle(articleID, h.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err = pkg.CreateComment(h.db, art, comment); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/article?id="+articleID, http.StatusSeeOther)

	case "DELETE":
		// Delete a comment
		params, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println(params.Get("article_id"))
		fmt.Println(params.Get("id"))

		art, err := pkg.GetArticle(params.Get("article_id"), h.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = pkg.DeleteComment(h.db, art, params.Get("id")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
