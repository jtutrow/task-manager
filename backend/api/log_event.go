package api

import (
	"log"

	"github.com/GeneralTask/task-manager/backend/database"
	"github.com/gin-gonic/gin"
)

type LogEventParams struct {
	EventType string `json:"event_type" binding:"required"`
}

func (api *API) LogEventAdd(c *gin.Context) {
	var params LogEventParams
	err := c.BindJSON(&params)
	if err != nil {
		log.Printf("error: %v", err)
		c.JSON(400, gin.H{"detail": "Invalid or missing 'event_type' parameter."})
		return
	}

	db, dbCleanup, err := database.GetDBConnection()
	if err != nil {
		Handle500(c)
		return
	}
	defer dbCleanup()

	err = database.InsertLogEvent(db, params.EventType)
	if err != nil {
		log.Printf("failed to insert waitlist entry: %v", err)
		Handle500(c)
		return
	}
	c.JSON(201, gin.H{})
}
