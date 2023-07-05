package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	db "github.com/sonyarouje/simdb"
)

var _handler *handler

const (
	articleID = "078b1ddc-15b2-49cd-b8c7-e1bfc972bb05"
	commentID = "60f91dca-7fee-4309-80ae-b244ada4cf6e"
)

func init() {
	driver, err := db.New("test")
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

	_handler = &handler{driver, t}
}

func TestHandler_Hello(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()

	_handler.handleHello(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if string(data) != "Hello, World!" {
		t.Errorf(`expected "Hello, World!" got %v`, string(data))
	}
}

func TestHandler_Index(t *testing.T) {
	tests := []struct {
		name           string
		req            *http.Request
		expectedStatus int
	}{
		{
			name:           "testing status 200 for GET",
			req:            httptest.NewRequest(http.MethodGet, "/", nil),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "testing status 405 for DELETE",
			req:            httptest.NewRequest(http.MethodDelete, "/", nil),
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			_handler.handleIndex(w, test.req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != test.expectedStatus {
				t.Errorf("expected %v got %v", test.expectedStatus, res.StatusCode)
			}
		})
	}
}

func TestHandler_Article(t *testing.T) {
	tests := []struct {
		name           string
		req            *http.Request
		expectedStatus int
	}{
		{
			name:           "testing 500; article not found",
			req:            httptest.NewRequest(http.MethodGet, "/article?id=345", nil),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "testing 200; article found",
			req:            httptest.NewRequest(http.MethodGet, "/article?id="+articleID, nil),
			expectedStatus: http.StatusOK,
		},
		{
			name: "post article success",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/article", nil)

				form := url.Values{}
				form.Add("title", "test")
				form.Add("author", "test")
				form.Add("content", "test")
				r.PostForm = form
				r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

				return r
			}(),
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "testing 405; DELETE",
			req:            httptest.NewRequest(http.MethodDelete, "/article?id="+articleID, nil),
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			_handler.handleArticle(w, test.req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != test.expectedStatus {
				t.Errorf("expected %v got %v", test.expectedStatus, res.StatusCode)
			}
		})
	}
}

func TestHandler_Comment(t *testing.T) {
	tests := []struct {
		name           string
		req            *http.Request
		expectedStatus int
	}{
		{
			name:           "GET method not allowed",
			req:            httptest.NewRequest(http.MethodGet, "/comment", nil),
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name: "POST; get article not found",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/comment?article_id=123", nil)

				form := url.Values{}
				form.Add("name", "test")
				form.Add("comment", "test")
				r.PostForm = form
				r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

				return r
			}(),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "POST success",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/comment?article_id="+articleID, nil)

				form := url.Values{}
				form.Add("name", "test")
				form.Add("comment", "test")
				r.PostForm = form
				r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

				return r
			}(),
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "DELETE success",
			req:            httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/comment?id=%s&article_id=%s", commentID, articleID), nil),
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			_handler.handleComment(w, test.req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != test.expectedStatus {
				t.Errorf("expected %v got %v", test.expectedStatus, res.StatusCode)
			}
		})
	}
}

func Test_isDeleted(t *testing.T) {
	// has been deleted
	if !isDeleted(time.Now()) {
		t.Errorf("expected true got false")
	}

	// has not been deleted
	if isDeleted(time.Time{}) {
		t.Errorf("expected false got true")
	}
}
