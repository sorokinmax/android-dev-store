package apk

import (
	"android-store/internal/db"
	"android-store/internal/globals"
	models "android-store/internal/models/apk"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/avast/apkparser"
)

func ApkProcessor(dirPath string, fileName string) *models.Apk {
	var apks []models.Apk
	var apk models.Apk

	filePath := dirPath + "/" + fileName

	manifest, err := apkParse(filePath)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "Failed to open the APK: %s", zipErr.Error())
		return nil
	}

	apk.Package = manifest.Package
	apk.AppName = manifest.Application.Name
	apk.AppLabel = manifest.Application.Label
	apk.AppIcon = manifest.Application.Icon
	apk.VersionName = manifest.VersionName
	apk.FileName = fileName
	apk.SHA256 = getSHA256(filePath)
	apk.UploadTime = time.Now().Format("2006.01.02 15:04:05")

	apks, _ = db.SQLiteGetApks()
	if !containsApks(apks, apk) {
		db.SQLiteAddApk(&apk)
	} else {
		log.Printf("%s already exist", apk.VersionName)
		return nil
	}

	/*if err == nil {
		//apk.CFBundleIdentifier = fmt.Sprint(apkInfo["CFBundleIdentifier"])
		apk.DateTime = time.Now().Format("2006.01.02 15:04:05")
		apk.SHA256 = getSHA256(filePath)
		apk.FileName = fileName
		apk.URL = fmt.Sprintf("%s/apk/%s/%s", cfg.Service.Url, apk.SHA256, apk.FileName)

		apks, _ = SQLiteGetApks()
		if !containsApks(apks, apk) {
			CopyFile(dirPath, fmt.Sprintf(".\\apk\\%s", apk.SHA256), fileName)
			SQLiteAddApk(apk)
			deleteFile(filePath)
			//log.Printf("IPA %s is added\n", fmt.Sprintf("%s-%s.%s", apk.CFBundleIdentifier, apk.CFBundleShortVersionString, apk.CFBundleVersion))
		} else {
			//log.Printf("IPA %s is already exist\n", fmt.Sprintf("%s-%s.%s", apk.CFBundleIdentifier, apk.CFBundleShortVersionString, apk.CFBundleVersion))
			deleteFile(filePath)
		}
	}*/
	err = os.MkdirAll(fmt.Sprintf("./apps/%s", apk.SHA256), 0777)
	if err != nil {
		panic(err)
	}
	moveFile(filePath, fmt.Sprintf("./apps/%s/%s", apk.SHA256, fileName))
	return &apk
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

func containsApks(ipaArr []models.Apk, apk models.Apk) bool {
	for _, a := range ipaArr {
		if a.SHA256 == apk.SHA256 {
			return true
		}
	}
	return false
}

func moveFile(source string, destination string) {
	err := os.Rename(source, destination)
	if err != nil {
		fmt.Println(err)
	}
}

func GetAPKURL(apk models.Apk) (URL string) {
	URL = fmt.Sprintf("%s/apps/%s/%s", globals.Config.Url, apk.SHA256, apk.FileName)
	return URL
}
