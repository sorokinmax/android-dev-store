package main

import (
	"log/slog"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

func HandlerError(ctx *gin.Context, statusCode int, err error, msg string) {
	if err != nil {
		color.Set(color.FgHiRed)
		slog.Error(msg, slog.Group("details", "rawError", err.Error()))
		color.Unset()
		ctx.AbortWithStatusJSON(statusCode, gin.H{"error": msg, "raw": err.Error()})
	} else {
		ctx.AbortWithStatusJSON(statusCode, gin.H{"error": msg, "raw": ""})
	}
}
