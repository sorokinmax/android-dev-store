package main

import (
	"android-store/internal/api"
	"android-store/internal/db"
	"android-store/internal/globals"
	models "android-store/internal/models/apk"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	_ "time/tzdata"

	sloggin "android-store/pkg/gin-slog"

	"github.com/caarlos0/env/v10"
	"github.com/fatih/color"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		slog.Group("pinfo",
			slog.String("name", globals.ServiceName),
			slog.String("version", globals.Version),
		),
	)
	slog.SetDefault(logger)

	err := env.Parse(&globals.Config)
	if err != nil {
		color.Set(color.FgHiRed)
		slog.Error(fmt.Sprintf("Cannot unmarshal environment variables: %v", err.Error()))
		color.Unset()
	}

	gin.SetMode(gin.ReleaseMode)
}

func main() {
	runtime.GOMAXPROCS(1)

	db.SQLiteCreateDB(models.Apk{})

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(sloggin.New(slog.Default()))

	router.HTMLRender = ginview.Default()
	//config := websspi.NewConfig()
	//auth, _ := websspi.New(config)

	//router.Use(MidAuth(auth))
	//router.Use(AddUserToCtx())
	router.StaticFile("favicon.ico", "./views/favicon.ico")
	router.Use(static.Serve("/apps", static.LocalFile("./apps", false)))
	router.Use(static.Serve("/icons", static.LocalFile("./icons", false)))
	router.GET("/", api.IndexHandler)
	//router.GET("/admin", middleware.MidAuth(auth), middleware.AddUserToCtx(), api.IndexHandler)
	router.GET("/app/:id", api.VersionHandler)
	router.GET("/qr/:id", api.QRHandler)
	router.POST("/remove", api.RemoveHandler)
	router.POST("/apk", api.PostApkHandler)

	slog.Info(fmt.Sprintf("Web is available at localhost:%d", globals.Config.HttpPort))
	router.Run(fmt.Sprintf(":%d", globals.Config.HttpPort))
}
