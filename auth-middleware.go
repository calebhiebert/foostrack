package main

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// AuthMiddleware checks the user's session and grabs the user from the database
// Sets the "general" object with some information, and continues the request on it's merry way
func AuthMiddleware(c *gin.Context) {
	session := sessions.Default(c)

	general := make(map[string]string)

	id := session.Get("id")

	if id != nil && id.(string) != "" {
		var user User

		if err := dbase.First(&user, "id = ?", id).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				general["isloggedin"] = "false"
			} else {
				SendError(http.StatusInternalServerError, c, err)
				return
			}
		} else {
			general["isloggedin"] = "true"
			general["username"] = user.Username
			general["picture_url"] = user.PictureURL
			general["user_id"] = user.ID
		}
	} else {
		general["isloggedin"] = "false"
	}

	c.Set("general", general)

	if err := session.Save(); err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Next()
}
