package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr/v2"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

// Database object, use this to make queries
var dbase *gorm.DB

// Files box, this contains templates/sql files
var files *packr.Box

// Assets box, this contains images, css, js to be served statically
var assets *packr.Box

var templates map[string]*template.Template

func main() {

	// Load environment variables from a .env file
	// https://github.com/joho/godotenv
	godotenv.Load()

	// Create file boxes. This makes sure static assets are included in the compiled binary
	// https://github.com/gobuffalo/packr (this program uses v2 of this library)
	files = packr.New("Box", "./templates")
	assets = packr.New("Assets", "./templates/assets")

	templates = make(map[string]*template.Template)
	initTemplates()

	// Initialize the database. Using https://gorm.io/
	initDB()

	// Create the default gin router
	// https://github.com/gin-gonic/gin
	r := gin.Default()

	// Create additional routes
	assetRoute := r.Group("/assets")
	api := r.Group("/api")

	// Sessions
	store := cookie.NewStore([]byte("secret"))

	// Add the session middleware
	r.Use(sessions.Sessions("session", store))

	// Add a Cache-Control header to all static assets
	assetRoute.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Next()
	})

	// Serve static assets from the templates/assets folder
	assetRoute.StaticFS("/", assets)

	r.Use(AuthMiddleware)

	// *****************************
	// * Define Application Routes *
	// *****************************

	r.GET("/index", GetIndex)

	r.GET("/startgame", GetStartGame)
	r.POST("/startgame", PostStartGame)

	r.GET("/login", Login)
	r.GET("/logout", Logout)
	r.GET("/callback", Callback)

	r.GET("/games", ListGames)
	r.GET("/game/:id", GetGame)
	api.GET("/games/:id/eventcount", GetGameEventCount)
	r.POST("/game/:id/goal", MarkGoal)
	r.POST("/game/:id/antigoal", MarkAntiGoal)
	r.POST("/game/:id/start", MarkStarted)
	r.POST("/game/:id/end", MarkEnded)
	r.POST("/game/:id/deadball", MarkDeadBall)
	r.POST("/game/:id/oob", MarkOutOfBounds)
	r.POST("/game/:id/swap", MarkSwap)

	r.POST("/events/:id/undo", PostEventUndo)

	r.GET("/teams/:id/edit", GetTeamEditForm)
	r.GET("/teams", GetTeamList)
	r.GET("/team/create", GetTeamForm)
	r.POST("/team/create", PostCreateTeam)

	r.GET("/user/:id", GetUser)

	// Fallback route, if the request does not match any of the above routes
	// the user will be redirected to the index page
	r.Use(func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/index")
	})

	port := 8080

	// Check the PORT environment variable and use it if present
	if os.Getenv("PORT") != "" {
		parsedPort, err := strconv.Atoi(os.Getenv("PORT"))
		if err == nil {
			port = parsedPort
		}
	}

	// Start the gin server
	r.Run(fmt.Sprintf(":%d", port))
}

// Defines all possible template pages, and the files that make them up
func initTemplates() {
	addTemplate("index", "base.html", "index.html")
	addTemplate("startgame", "base.html", "start-game.html")
	addTemplate("error", "base.html", "error.html")
	addTemplate("game", "base.html", "game.html")
	addTemplate("games", "base.html", "game-list.html")
	addTemplate("notfound", "base.html", "not-found.html")
	addTemplate("blocked", "base.html", "blocked.html")
	addTemplate("user", "base.html", "user.html")
	addTemplate("teamform", "base.html", "team-form.html")
	addTemplate("teamlist", "base.html", "team-list.html")
}

// Compiles multiple files into a single template
// https://github.com/gin-contrib/multitemplate
func addTemplate(name string, filename ...string) {
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

	templates[name] = tmpl
}

// Starts a connection to the database and executes the schema.sql file
// Any migrations, etc... should go in that file
// TODO: move migrations into their own sql script
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

	sql, err = files.FindString("migrate.sql")
	if err != nil {
		panic(err)
	}

	if err = dbase.Exec(sql).Error; err != nil {
		panic(err)
	}

	sql, err = files.FindString("views.sql")
	if err != nil {
		panic(err)
	}

	if err = dbase.Exec(sql).Error; err != nil {
		panic(err)
	}
}
