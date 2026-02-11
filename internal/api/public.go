package api

import (
	"android-store/internal/db"
	models "android-store/internal/models/apk"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetApksHandler(ctx *gin.Context) {
	var apks []models.Apk
	apks, err := db.SQLiteGetApks()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch APKs"})
		return
	}

	for i := range apks {
		apks[i].ApkUrl = getApkUrl(apks[i])
		apks[i].AabUrl = getAabUrl(apks[i])
	}

	ctx.JSON(http.StatusOK, apks)
}
