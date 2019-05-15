package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// PostEventUndo will undo an event or mark it as deleted
func PostEventUndo(c *gin.Context) {

	id := c.Param("id")

	var event GameEvent

	if err := dbase.First(&event, "id = ?", id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	// Validate event type
	if event.EventType == "ptp" || event.EventType == "start" || event.EventType == "end" {
		SendError(http.StatusBadRequest, c, errors.New(fmt.Sprintf("Cannot delte event of type %v", event.EventType)))
		return
	}

	if err := dbase.Where("id = ?", event.ID).Delete(&GameEvent{}).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/game/%d", event.GameID))
}
