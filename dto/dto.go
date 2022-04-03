package dto

import (
	"net/http"
	"time"
)

type Genre uint8

const (
	PROGRAMMING Genre = iota
	DHBW
	OTHER
)

type MehmDTO struct {
	Id          int       `json:"id"`
	AuthorName  string    `json:"authorName"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageSource string    `json:"imageSource"`
	CreatedDate time.Time `json:"createdDate"`
	Genre       Genre     `json:"genre"`
	Likes       int       `json:"likes"`
}

type CommentDTO struct {
	Comment  string    `json:"id"`
	Author   string    `json:"author"`
	DateTime time.Time `json:"dateTime"`
}

type LoggedIn struct {
	http.Cookie `json:"jwt"`
	Id          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Admin       bool   `json:"Admin"`
}

type CommentInput struct {
	MehmID  int64  `json:"id"`
	Comment string `json:"text"`
}

type MehmInput struct {
	Description string `json:"description"`
	Title       string `json:"title"`
}

type Comment struct {
	MehmId  int64  `json:"mehmId"`
	Comment string `json:"comment"`
}
