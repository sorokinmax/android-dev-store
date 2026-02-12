package api

import (
	"android-store/internal/db"
	"android-store/internal/globals"
	models "android-store/internal/models/apk"
	telegram "android-store/pkg/telegram"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/casbin/casbin"
	"github.com/gin-gonic/gin"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/fogleman/gg"

	"github.com/avast/apkparser"
)

var (
	enforcer     *casbin.Enforcer
	enforcerOnce sync.Once
)

func getEnforcer() *casbin.Enforcer {
	enforcerOnce.Do(func() {
		enforcer = casbin.NewEnforcer("./data/model.conf", "./data/policy.csv")
	})
	return enforcer
}

func IndexHandler(ctx *gin.Context) {
	var apks []models.Apk
	apks, _ = db.SQLiteGetApks()
	for id, apk := range apks {
		apks[id].ApkUrl = getApkUrl(apk)
		apks[id].AabUrl = getAabUrl(apk)
	}

	// Casbin enforcer
	e := getEnforcer()
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

func VersionHandler(ctx *gin.Context) {
	var apk models.Apk
	var id = ctx.Param("id")

	apk, _ = db.SQLiteGetApk(id)
	apk.ApkUrl = getApkUrl(apk)
	apk.AabUrl = getAabUrl(apk)

	ctx.HTML(http.StatusOK, "version/index", gin.H{
		"title":       globals.ServiceFriendlyName,
		"version":     globals.Version,
		"apk":         apk,
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

func QRHandler(ctx *gin.Context) {
	var app models.Apk
	var id = ctx.Param("id")

	app, _ = db.SQLiteGetApk(id)
	app.ApkUrl = getApkUrl(app)

	data := app.ApkUrl
	description := fmt.Sprintf("%s-%s (%s)", app.AppLabel, app.VersionName, app.APKFileName)

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
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"responce": fmt.Sprintf("Get form error: %s", err.Error())})
		return
	}

	// Extract the apk and aab files
	apkFound := false
	aabFound := false
	sbomFound := false
	var apkFile, aabFile *multipart.FileHeader
	var sbomFile *multipart.FileHeader
	files := form.File["files"]
	for _, file := range files {
		if strings.HasSuffix(file.Filename, ".apk") {
			apkFound = true
			apkFile = file
		}
		if strings.HasSuffix(file.Filename, ".aab") {
			aabFound = true
			aabFile = file
		}
		if strings.HasSuffix(file.Filename, "sbom.json") {
			sbomFound = true
			sbomFile = file
		}
	}
	defer os.Remove(globals.TMPDIR + filepath.Base(apkFile.Filename))
	if aabFound {
		defer os.Remove(globals.TMPDIR + filepath.Base(aabFile.Filename))
	}
	if sbomFound {
		defer os.Remove(globals.TMPDIR + filepath.Base(sbomFile.Filename))
	}

	// One of the files must be an APK
	if !apkFound {
		ctx.JSON(http.StatusBadRequest, gin.H{"responce": "APK not found"})
		return
	}

	// Saving the APK
	if err := ctx.SaveUploadedFile(apkFile, globals.TMPDIR+filepath.Base(apkFile.Filename)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"responce": fmt.Sprintf("Upload file error: %s", err.Error())})
		return
	}

	// Saving the AAB
	var aabFileName string
	if aabFound {
		if err := ctx.SaveUploadedFile(aabFile, globals.TMPDIR+filepath.Base(aabFile.Filename)); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"responce": fmt.Sprintf("Upload file error: %s", err.Error())})
			return
		}
		aabFileName = filepath.Base(aabFile.Filename)
	}

	// Saving SBOM
	var sbomFileName string
	if sbomFound {
		if err := ctx.SaveUploadedFile(sbomFile, globals.TMPDIR+filepath.Base(sbomFile.Filename)); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"responce": fmt.Sprintf("Upload file error: %s", err.Error())})
			return
		}
		sbomFileName = filepath.Base(sbomFile.Filename)
	}

	// Checking for re-placement
	var apks []models.Apk
	apks, _ = db.SQLiteGetApks()
	if containsApks(apks, getSHA256(globals.TMPDIR+filepath.Base(apkFile.Filename))) {
		log.Println("This APK already exist")
		ctx.JSON(http.StatusBadRequest, gin.H{"responce": "This APK already exist"})
		return
	}

	// Processing
	app, err := apkProcessor(filepath.Base(apkFile.Filename), aabFileName, sbomFileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"responce": fmt.Sprintf("APK processing error: %s", err.Error())})
		return
	}

	// Notification
	if globals.Config.BotToken != "" && globals.Config.ChatID != 0 {
		msg := fmt.Sprintf("New build <a href='%s/app/%d'>%s %s</a> is ready", globals.Config.Url, app.ID, app.AppLabel, app.VersionName)
		telegram.TgSendMessage(globals.Config.BotToken, msg, globals.Config.ChatID)
	}
	ctx.JSON(http.StatusOK, gin.H{"responce": "File processed"})
}

// PutApkHandler updates the metadata of an existing APK and renames the physical files
// on disk if APKFileName or AABFileName are changed. It supports updating any field
// present in the Apk model via form data in the request body.
func PutApkHandler(ctx *gin.Context) {
	// Require the APK's database ID to identify which record to update
	id := ctx.PostForm("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"responce": "ID not provided"})
		return
	}

	apk, _ := db.SQLiteGetApk(id)
	oldAPKName := apk.APKFileName
	oldAABName := apk.AABFileName

	// Ensure form data is parsed so we can read all fields
	_ = ctx.Request.ParseForm()

	var newAPKName, newAABName string

	// Apply all provided fields to the Apk struct
	for key, vals := range ctx.Request.PostForm {
		if len(vals) == 0 {
			continue
		}
		val := vals[0]
		switch key {
		case "APKFileName":
			if val != apk.APKFileName {
				newAPKName = val
				apk.APKFileName = val
			}
		case "AABFileName":
			if val != apk.AABFileName {
				newAABName = val
				apk.AABFileName = val
			}
		case "UploadTime":
			apk.UploadTime = val
		case "APKSHA256":
			apk.APKSHA256 = val
		case "Package":
			apk.Package = val
		case "VersionCode":
			apk.VersionCode = val
		case "VersionName":
			apk.VersionName = val
		case "AppLabel":
			apk.AppLabel = val
		case "AppIcon":
			apk.AppIcon = val
		case "AppName":
			apk.AppName = val
			// ignore ApkUrl, AabUrl as they are not persisted
		}
	}

	// Rename APK file if needed
	if newAPKName != "" && newAPKName != oldAPKName {
		oldPath := fmt.Sprintf("./data/apps/%s/%s", apk.APKSHA256, oldAPKName)
		newPath := fmt.Sprintf("./data/apps/%s/%s", apk.APKSHA256, newAPKName)
		if err := os.Rename(oldPath, newPath); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"responce": fmt.Sprintf("Rename APK failed: %s", err.Error())})
			return
		}
		apk.APKFileName = newAPKName
	}

	// Rename AAB file if needed
	if newAABName != "" && newAABName != oldAABName {
		oldPath := fmt.Sprintf("./data/apps/%s/%s", apk.APKSHA256, oldAABName)
		newPath := fmt.Sprintf("./data/apps/%s/%s", apk.APKSHA256, newAABName)
		if err := os.Rename(oldPath, newPath); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"responce": fmt.Sprintf("Rename AAB failed: %s", err.Error())})
			return
		}
		apk.AABFileName = newAABName
	}

	// Persist updates to the DB
	db.SQLiteSaveApk(apk)
	ctx.JSON(http.StatusOK, gin.H{"responce": "APK updated"})
}

//////////
// ADDITIONAL FUNCTIONS
//////////

func apkProcessor(apkFileName string, aabFileName string, sbomFileName string) (*models.Apk, error) {
	var apk models.Apk
	var aabFilePath string
	var sbomFilePath string

	apkFilePath := globals.TMPDIR + apkFileName

	if aabFileName != "" {
		aabFilePath = globals.TMPDIR + aabFileName
		apk.AABFileName = aabFileName
	} else {
		apk.AABFileName = ""
	}

	if sbomFileName != "" {
		sbomFilePath = globals.TMPDIR + sbomFileName
		apk.SBOMFileName = aabFileName
	} else {
		apk.SBOMFileName = aabFileName
	}

	manifest, err := apkParse(apkFilePath)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "Failed to open the APK: %s", zipErr.Error())
		return nil, err
	}

	apk.Package = manifest.Package
	apk.AppName = manifest.Application.Name
	apk.AppLabel = manifest.Application.Label
	apk.AppIcon = manifest.Application.Icon
	apk.VersionCode = manifest.VersionCode
	apk.VersionName = manifest.VersionName
	apk.APKFileName = apkFileName
	apk.APKSHA256 = getSHA256(apkFilePath)
	apk.UploadTime = time.Now().Format("2006.01.02 15:04:05")

	err = os.MkdirAll(fmt.Sprintf("./data/apps/%s", apk.APKSHA256), 0777)
	if err != nil {
		panic(err)
	}

	// Moving APK
	err = moveFile(apkFilePath, fmt.Sprintf("./data/apps/%s/%s", apk.APKSHA256, apkFileName))
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	// Moving AAB
	if aabFileName != "" {
		err = moveFile(aabFilePath, fmt.Sprintf("./data/apps/%s/%s", apk.APKSHA256, aabFileName))
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}
	}
	// Moving SBOM
	if sbomFileName != "" {
		err = moveFile(sbomFilePath, fmt.Sprintf("./data/apps/%s/%s", apk.APKSHA256, sbomFileName))
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}
	}

	db.SQLiteAddApk(&apk)
	return &apk, nil
}

func apkParse(name string) (*models.Manifest, error) {

	w := &bytes.Buffer{}
	enc := xml.NewEncoder(w)
	enc.Indent("", "\t")

	//parser := apkparser.NewParser()

	zipErr, resErr, manErr := apkparser.ParseApk(name, enc)

	if zipErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to open the APK: %s", zipErr.Error())
		return nil, zipErr
	}
	if resErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse resources: %s", resErr.Error())
	}
	if manErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse AndroidManifest.xml: %s", manErr.Error())
		return nil, resErr
	}
	var manifest models.Manifest
	err := xml.Unmarshal(w.Bytes(), &manifest)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}

	//fmt.Println(apk)
	return &manifest, nil
}

func getSHA256(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(h.Sum(nil))
}

func containsApks(ipaArr []models.Apk, apkSha256 string) bool {
	for _, a := range ipaArr {
		if a.APKSHA256 == apkSha256 {
			return true
		}
	}
	return false
}

func getApkUrl(apk models.Apk) (URL string) {
	URL = fmt.Sprintf("%s/apps/%s/%s", globals.Config.Url, apk.APKSHA256, apk.APKFileName)
	return URL
}
func getAabUrl(apk models.Apk) (URL string) {
	if apk.AABFileName == "" {
		return ""
	}
	URL = fmt.Sprintf("%s/apps/%s/%s", globals.Config.Url, apk.APKSHA256, apk.AABFileName)
	return URL
}

func moveFile(source, destination string) error {
	err := os.Rename(source, destination)
	if err != nil && strings.Contains(err.Error(), "invalid cross-device link") {
		return moveCrossDevice(source, destination)
	}
	return err
}

func moveCrossDevice(source, destination string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	dst, err := os.Create(destination)
	if err != nil {
		src.Close()
		return err
	}
	_, err = io.Copy(dst, src)
	src.Close()
	dst.Close()
	if err != nil {
		return err
	}
	fi, err := os.Stat(source)
	if err != nil {
		os.Remove(destination)
		return err
	}
	err = os.Chmod(destination, fi.Mode())
	if err != nil {
		os.Remove(destination)
		return err
	}
	os.Remove(source)
	return nil
}
