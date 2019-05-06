package main

type User struct {
	ID         string `gorm:"primary_key;unique_index"`
	Username   string
	PictureURL string
	Events     []GameEvent `gorm:"foreignkey:UserID"`
}
