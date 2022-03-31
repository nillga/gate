package dto

import "time"

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

type CommentInput struct {
	UserId  int64 `json:"userId"`
	Id      int64  `json:"id"`
	Comment string `json:"comment"`
	Admin   bool   `json:"isAdmin"`
}

type MehmInput struct {
	UserId      int64  `json:"userId"`
	Description string `json:"description"`
	Title       string `json:"title"`
	Admin       bool   `json:"isAdmin"`
}
