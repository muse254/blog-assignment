package pkg

import (
	"time"

	"github.com/google/uuid"
	db "github.com/sonyarouje/simdb"
)

// An Article is a blog post.
type Article struct {
	Id       string
	Title    string
	Author   string
	Content  string
	Comments []Comment
}

// This method is required by the simdb library.
func (a Article) ID() (jsonField string, value interface{}) {
	value = a.Id
	jsonField = "Id"
	return
}

// A Comment is a comment on a blog post.
type Comment struct {
	Id        string
	Name      string
	Comment   string
	CreatedAt time.Time
	DeletedAt time.Time
}

type DeleteCommentData struct {
	Data []DeleteEntry
}

type DeleteEntry struct {
	CreatedAt time.Time
	DeletedAt time.Time
}

func CreateArticle(art Article, driver *db.Driver) (*Article, error) {
	art.Id = uuid.New().String()
	if err := driver.Insert(art); err != nil {
		return nil, err
	}

	return &art, nil
}

func GetArticle(id string, driver *db.Driver) (*Article, error) {
	var art Article
	err := driver.Open(Article{}).Where("Id", "=", id).First().AsEntity(&art)
	if err != nil {
		return nil, err
	}

	return &art, nil
}

func GetArticles(driver *db.Driver) ([]Article, error) {
	var arts []Article
	if err := driver.Open(Article{}).Get().AsEntity(&arts); err != nil {
		return nil, err
	}

	return arts, nil
}

func CreateComment(driver *db.Driver, art *Article, com Comment) (*Comment, error) {
	com.Id = uuid.New().String()
	art.Comments = append(art.Comments, com)
	if err := driver.Update(art); err != nil {
		return nil, err
	}

	return &com, nil
}

func DeleteComment(driver *db.Driver, art *Article, commentID string) error {
	for i, com := range art.Comments {
		if com.Id == commentID {
			art.Comments[i].DeletedAt = time.Now()
			if err := driver.Update(art); err != nil {
				return err
			}
		}
	}

	return nil
}
