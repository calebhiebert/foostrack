package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	r := gin.Default()
	r.HTMLRender = createRenderer()

	// Sessions
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("session", store))

	r.GET("/index", func(c *gin.Context) {
		session := sessions.Default(c)

		fmt.Println(session)

		c.HTML(http.StatusOK, "index", gin.H{
			"title":       "this is a title",
			"username":    session.Get("username"),
			"picture_url": session.Get("picture_url"),
		})
	})

	r.GET("/startgame", func(c *gin.Context) {
		c.HTML(http.StatusOK, "startgame", gin.H{
			"test": "testing testing",
		})
	})

	r.POST("/startgame", func(c *gin.Context) {
		fmt.Println("Stuff", c.PostForm("blue_goalie"))

		c.HTML(http.StatusOK, "startgame", gin.H{})
	})

	r.GET("/login", Login)
	r.GET("/callback", Callback)

	r.Run(":8080")
}

func createRenderer() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("index", "templates/base.html", "templates/index.html")
	r.AddFromFiles("startgame", "templates/base.html", "templates/start-game.html")
	r.AddFromFiles("error", "templates/base.html", "templates/error.html")
	return r
}
