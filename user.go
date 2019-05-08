package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type User struct {
	ID         string `gorm:"primary_key;unique_index"`
	Username   string
	PictureURL string
	Events     []GameEvent `gorm:"foreignkey:UserID"`
}

type UserWithPosition struct {
	ID         string
	Username   string
	PictureURL string
	Team       string
	Position   string
}

// GetUser will render a user page
func GetUser(c *gin.Context) {
	id := c.Param("id")

	var user User

	if err := dbase.First(&user, "id = ?", id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	SendHTML(http.StatusOK, c, "user", gin.H{
		"user": user,
	})
}
