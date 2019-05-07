package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	err := session.Save()
	if err != nil {
		SendError(http.StatusBadRequest, c, err)
		return
	}

	audience := oauth2.SetAuthURLParam("audience", aud)
	url := conf.AuthCodeURL(state, audience)

	c.Redirect(http.StatusFound, url)
}

func Logout(c *gin.Context) {

	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		SendError(http.StatusBadRequest, c, err)
		return
	}

	c.Redirect(http.StatusFound, "/index")
}

func Callback(c *gin.Context) {
	conf := getOauth2Config()

	session := sessions.Default(c)

	queryState := c.Query("state")
	sessionState := session.Get("state")

	if sessionState == nil {
		SendError(http.StatusBadRequest, c, errors.New("Missing state"))
		return
	}

	if queryState != sessionState.(string) {
		SendError(http.StatusBadRequest, c, errors.New("Invalid state parameter"))
		return
	}

	code := c.Query("code")

	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	client := conf.Client(context.Background(), token)
	resp, err := client.Get("https://" + os.Getenv("DOMAIN") + "/userinfo")
	if err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	defer resp.Body.Close()

	var profile map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	session.Set("id", profile["sub"])
	err = session.Save()

	var existingUser User

	if err = dbase.First(&existingUser, "id = ?", profile["sub"]).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			user := User{
				ID:         profile["sub"].(string),
				Username:   profile["name"].(string),
				PictureURL: profile["picture"].(string),
			}

			if err = dbase.Create(&user).Error; err != nil {
				SendError(http.StatusInternalServerError, c, err)
				return
			}
		} else {
			SendError(http.StatusInternalServerError, c, err)
			return
		}
	}

	if err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Redirect(http.StatusFound, "/index")
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
