package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SendHTML is a wrapper around gin context HTML function, it includes the "general"
// object to be sent to the template. This object contains info such as whether or not
// the user is logged in, their username, etc...
func SendHTML(statusCode int, c *gin.Context, page string, data gin.H) {

	if data == nil {
		data = gin.H{}
	}

	data["general"] = c.GetStringMapString("general")

	c.HTML(statusCode, page, data)
}

// SendError is a wrapper around SendHTML that sends the error.html template
func SendError(code int, c *gin.Context, err error) {
	SendHTML(code, c, "error", gin.H{
		"error": err,
	})
}

// SendNotFound is a wrapper around SendHTML that sends a 404 page
func SendNotFound(c *gin.Context) {
	SendHTML(http.StatusNotFound, c, "notfound", nil)
}

// EnsureLoggedIn will send a not logged in page/message and return false if the
// current user is not logged in, does nothing and returns true otherwise
func EnsureLoggedIn(c *gin.Context) bool {
	general := c.GetStringMapString("general")

	if general["isloggedin"] != "true" {
		SendHTML(http.StatusForbidden, c, "blocked", nil)
		return false
	}

	return true
}

// PrettyDuration formats a time.Duration into a string, (only goes up to minutes)
func PrettyDuration(duration time.Duration) string {
	seconds := duration.Seconds()

	remainingSeconds := int64(seconds) % 60
	remainingMinutes := (int64(seconds) - remainingSeconds) / 60

	if remainingMinutes > 0 {
		return fmt.Sprintf("%d min %d sec", remainingMinutes, remainingSeconds)
	}

	return fmt.Sprintf("%d sec", remainingSeconds)
}

// ExtractFirstName returns the first name from a full name
func ExtractFirstName(name string) string {
	nameParts := strings.Split(name, " ")

	if len(nameParts) > 0 {
		return nameParts[0]
	}

	return ""
}
