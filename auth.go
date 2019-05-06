package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"
)

func Login(c *gin.Context) {
	aud := "https://foostrack.panchem.io"

	conf := getOauth2Config()

	// Generate a random state
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.StdEncoding.EncodeToString(b)

	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()

	audience := oauth2.SetAuthURLParam("audience", aud)
	url := conf.AuthCodeURL(state, audience)

	c.Redirect(http.StatusMovedPermanently, url)
}

func Callback(c *gin.Context) {
	conf := getOauth2Config()

	session := sessions.Default(c)

	queryState := c.Query("state")
	state := session.Get("state").(string)

	if queryState != state {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"error": "Invalid state parameter",
		})
		return
	}

	code := c.Query("code")

	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"error": err,
		})
		return
	}

	client := conf.Client(context.Background(), token)
	resp, err := client.Get("https://" + os.Getenv("DOMAIN") + "/userinfo")
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"error": err,
		})
		return
	}

	defer resp.Body.Close()

	var profile map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"error": err,
		})
		return
	}

	session.Set("access_token", token.AccessToken)
	session.Set("username", profile["name"])

	var existingUser User

	if err = dbase.First(&existingUser, "id = ?", profile["sub"]).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			user := User{
				ID:         profile["sub"].(string),
				Username:   profile["name"].(string),
				PictureURL: profile["picture"].(string),
			}

			if err = dbase.Create(&user).Error; err != nil {
				c.HTML(http.StatusInternalServerError, "error", gin.H{
					"error": err,
				})
				return
			}
		} else {
			c.HTML(http.StatusInternalServerError, "error", gin.H{
				"error": err,
			})
			return
		}
	}

	err = session.Save()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"error": err,
		})
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/index")
}

func getOauth2Config() *oauth2.Config {
	domain := os.Getenv("DOMAIN")

	return &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		Scopes:       []string{"openid", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://" + domain + "/authorize",
			TokenURL: "https://" + domain + "/oauth/token",
		},
	}
}
