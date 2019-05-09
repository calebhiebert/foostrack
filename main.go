package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr/v2"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

var dbase *gorm.DB
var files *packr.Box
var assets *packr.Box

func main() {
	godotenv.Load()
	files = packr.New("Box", "./templates")
	assets = packr.New("Assets", "./templates/assets")
	initDB()

	r := gin.Default()
	assetRoute := r.Group("/assets")
	api := r.Group("/api")

	r.HTMLRender = createRenderer()

	// Sessions
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("session", store))

	assetRoute.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "max-age=86400")

		c.Next()
	})
	assetRoute.StaticFS("/", assets)

	r.Use(func(c *gin.Context) {
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
		session.Save()

		c.Next()
	})

	r.GET("/index", GetIndex)

	r.GET("/startgame", GetStartGame)
	r.POST("/startgame", PostStartGame)

	r.GET("/login", Login)
	r.GET("/logout", Logout)
	r.GET("/callback", Callback)

	r.GET("/games", ListGames)
	r.GET("/game/:id", GetGame)
	api.GET("/games/:id/eventcount", GetGameEventCount)
	r.GET("/game/:id/goal", MarkGoal)
	r.POST("/game/:id/goal", MarkGoal)
	r.POST("/game/:id/start", MarkStarted)
	r.POST("/game/:id/end", MarkEnded)
	r.POST("/game/:id/deadball", MarkDeadBall)
	r.POST("/game/:id/oob", MarkOutOfBounds)
	r.POST("/game/:id/swap", MarkSwap)

	r.GET("/user/:id", GetUser)

	// Catch all other routes and redirect to index
	r.Use(func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/index")
	})

	port := 8080

	if os.Getenv("PORT") != "" {
		parsedPort, err := strconv.Atoi(os.Getenv("PORT"))
		if err == nil {
			port = parsedPort
		}
	}

	r.Run(fmt.Sprintf(":%d", port))
}

func createRenderer() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	addTemplate(r, "index", "base.html", "index.html")
	addTemplate(r, "startgame", "base.html", "start-game.html")
	addTemplate(r, "error", "base.html", "error.html")
	addTemplate(r, "game", "base.html", "game.html")
	addTemplate(r, "games", "base.html", "game-list.html")
	addTemplate(r, "notfound", "base.html", "not-found.html")
	addTemplate(r, "blocked", "base.html", "blocked.html")
	addTemplate(r, "user", "base.html", "user.html")
	return r
}

func addTemplate(r multitemplate.Renderer, name string, filename ...string) {
	tmpl := template.New(name)

	for _, file := range filename {
		contents, err := files.FindString(file)
		if err != nil {
			panic(err)
		}

		tmpl, err = tmpl.Parse(contents)
		if err != nil {
			panic(err)
		}
	}

	r.Add(name, tmpl)
}

func initDB() {
	db, err := gorm.Open("postgres", os.Getenv("CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}

	dbase = db

	sql, err := files.FindString("schema.sql")
	if err != nil {
		panic(err)
	}

	if err := dbase.Exec(sql).Error; err != nil {
		panic(err)
	}
}
