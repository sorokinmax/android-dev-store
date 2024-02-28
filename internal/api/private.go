package api

import (
	"android-store/internal/apk"
	"android-store/internal/db"
	"android-store/internal/globals"
	models "android-store/internal/models/apk"
	telegram "android-store/pkg/telegram"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"strings"

	"github.com/casbin/casbin"
	"github.com/gin-gonic/gin"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/fogleman/gg"
)

func IndexHandler(ctx *gin.Context) {
	var apks []models.Apk
	apks, _ = db.SQLiteGetApks()
	for id, app := range apks {
		apks[id].URL = apk.GetAPKURL(app)
	}

	e := casbin.NewEnforcer("./data/model.conf", "./data/policy.csv")
	//log.Println(ctx.Value("user"))

	// Admin rights
	if user := ctx.Value("user"); user != nil {

		if e.Enforce(user, "index", "write") {
			ctx.HTML(http.StatusOK, "index", gin.H{
				"title":       globals.ServiceFriendlyName,
				"version":     globals.Version,
				"apks":        apks,
				"admin":       1,
				"service_url": globals.Config.Url,
			},
			)
			return
		}
	}

	// Guest rights
	ctx.HTML(http.StatusOK, "index", gin.H{
		"title":       globals.ServiceFriendlyName,
		"version":     globals.Version,
		"apks":        apks,
		"admin":       0,
		"service_url": globals.Config.Url,
	},
	)
}

func RemoveHandler(ctx *gin.Context) {
	var apk models.Apk
	var id = ctx.PostForm("id")
	apk, _ = db.SQLiteGetApk(id)
	//RemoveDir(fmt.Sprintf(".\\apk\\%s", apk.SHA256))
	db.SQLiteDelApk(apk)
	ctx.Redirect(http.StatusMovedPermanently, globals.Config.Url)
	log.Println("Apk delete has completed")
}

func VersionHandler(ctx *gin.Context) {
	var apk models.Apk
	var id = ctx.Param("id")

	apk, _ = db.SQLiteGetApk(id)
	apk.URL = fmt.Sprintf("%s/apps/%s/%s", globals.Config.Url, apk.SHA256, apk.FileName)

	ctx.HTML(http.StatusOK, "version/index", gin.H{
		"title":       globals.ServiceFriendlyName,
		"version":     globals.Version,
		"apk":         apk,
		"service_url": globals.Config.Url,
	},
	)
}

func QRHandler(ctx *gin.Context) {
	var app models.Apk
	var id = ctx.Param("id")

	app, _ = db.SQLiteGetApk(id)
	app.URL = apk.GetAPKURL(app)

	data := app.URL
	description := fmt.Sprintf("%s-%s (%s)", app.AppLabel, app.VersionName, app.Package)

	qrCode, _ := qr.Encode(data, qr.L, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 600, 600)

	im := qrCode

	dc := gg.NewContext(600, 626)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	if err := dc.LoadFontFace("./data/arial.ttf", 16); err != nil {
		panic(err)
	}

	dc.DrawRoundedRectangle(0, 0, 600, 626, 0)
	dc.DrawImage(im, 0, 0)
	dc.DrawStringAnchored(description, 300, 615, 0.5, 0)
	dc.Clip()

	png.Encode(ctx.Writer, dc.Image())

	ctx.String(http.StatusOK, "Done")
}

func PostApkHandler(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		log.Fatal(err)
	}
	if !strings.HasSuffix(file.Filename, ".apk") {
		ctx.JSON(http.StatusBadRequest, gin.H{"responce": "Invalid file extension"})
		return
	}

	//log.Println(file.Filename)

	err = ctx.SaveUploadedFile(file, "./apps/"+file.Filename)
	if err != nil {
		log.Fatal(err)
	}

	app := apk.ApkProcessor("./apps", file.Filename)

	msg := fmt.Sprintf("New build %s %s is ready", app.AppLabel, app.VersionName)
	telegram.TgSendMessage(globals.Config.BotToken, msg, globals.Config.ChatID)
	ctx.JSON(http.StatusOK, gin.H{"responce": "File processed"})
}
