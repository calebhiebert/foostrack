package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

var dbase *gorm.DB

func main() {
	godotenv.Load()
	initDB()

	r := gin.Default()
	r.HTMLRender = createRenderer()

	// Sessions
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("session", store))
	r.Use(static.Serve("/assets", static.LocalFile("templates/assets", false)))

	r.Use(func(c *gin.Context) {
		session := sessions.Default(c)

		general := make(map[string]string)

		id := session.Get("id")

		if id != nil && id.(string) != "" {
			var user User

			if err := dbase.First(&user, "id = ?", id).Error; err != nil {
				c.HTML(http.StatusInternalServerError, "error", gin.H{
					"error":   err,
					"general": c.GetStringMapString("general"),
				})
				return
			}

			general["isloggedin"] = "true"
			general["username"] = user.Username
			general["picture_url"] = user.PictureURL
			general["user_id"] = user.ID
		} else {
			general["isloggedin"] = "false"
		}

		c.Set("general", general)
		session.Save()

		c.Next()
	})

	r.GET("/index", func(c *gin.Context) {
		session := sessions.Default(c)

		fmt.Println(session)

		c.HTML(http.StatusOK, "index", gin.H{
			"title":       "this is a title",
			"username":    session.Get("username"),
			"picture_url": session.Get("picture_url"),
			"general":     c.GetStringMapString("general"),
		})
	})

	r.GET("/startgame", GetStartGame)
	r.POST("/startgame", PostStartGame)

	r.GET("/login", Login)
	r.GET("/logout", Logout)
	r.GET("/callback", Callback)

	r.GET("/game/:id", GetGame)

	r.Run(":8080")
}

func createRenderer() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("index", "templates/base.html", "templates/index.html")
	r.AddFromFiles("startgame", "templates/base.html", "templates/start-game.html")
	r.AddFromFiles("error", "templates/base.html", "templates/error.html")
	r.AddFromFiles("game", "templates/base.html", "templates/game.html")
	r.AddFromFiles("notfound", "templates/base.html", "templates/not-found.html")
	return r
}

func initDB() {
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=foostrack dbname=foostrack password=foostrack sslmode=disable")
	if err != nil {
		panic(err)
	}

	dbase = db

	dbase.AutoMigrate(&User{})
	dbase.AutoMigrate(&Game{})
	dbase.AutoMigrate(&GameEvent{})
}
