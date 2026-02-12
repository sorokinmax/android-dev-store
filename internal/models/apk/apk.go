package models

import (
	"encoding/xml"

	"gorm.io/gorm"
)

type Apk struct {
	gorm.Model

	ApkUrl       string `gorm:"-"`
	AabUrl       string `gorm:"-"`
	APKFileName  string
	AABFileName  string
	SBOMFileName string
	UploadTime   string
	APKSHA256    string

	Package     string
	VersionCode string
	VersionName string
	AppLabel    string
	AppIcon     string
	AppName     string
}

type Manifest struct {
	XMLName                   xml.Name `xml:"manifest"`
	Text                      string   `xml:",chardata"`
	Android                   string   `xml:"android,attr"`
	VersionCode               string   `xml:"versionCode,attr"`
	VersionName               string   `xml:"versionName,attr"`
	CompileSdkVersion         string   `xml:"compileSdkVersion,attr"`
	CompileSdkVersionCodename string   `xml:"compileSdkVersionCodename,attr"`
	Package                   string   `xml:"package,attr"`
	PlatformBuildVersionCode  string   `xml:"platformBuildVersionCode,attr"`
	PlatformBuildVersionName  string   `xml:"platformBuildVersionName,attr"`
	UsesSdk                   struct {
		Text             string `xml:",chardata"`
		MinSdkVersion    string `xml:"minSdkVersion,attr"`
		TargetSdkVersion string `xml:"targetSdkVersion,attr"`
	} `xml:"uses-sdk"`
	UsesPermission []struct {
		Text string `xml:",chardata"`
		Name string `xml:"name,attr"`
	} `xml:"uses-permission"`
	Application struct {
		Text                  string `xml:",chardata"`
		Theme                 string `xml:"theme,attr"`
		Label                 string `xml:"label,attr"`
		Icon                  string `xml:"icon,attr"`
		Name                  string `xml:"name,attr"`
		AllowBackup           string `xml:"allowBackup,attr"`
		NetworkSecurityConfig string `xml:"networkSecurityConfig,attr"`
		RoundIcon             string `xml:"roundIcon,attr"`
		AppComponentFactory   string `xml:"appComponentFactory,attr"`
		Activity              []struct {
			Text                string `xml:",chardata"`
			Label               string `xml:"label,attr"`
			Name                string `xml:"name,attr"`
			Exported            string `xml:"exported,attr"`
			LaunchMode          string `xml:"launchMode,attr"`
			ScreenOrientation   string `xml:"screenOrientation,attr"`
			ConfigChanges       string `xml:"configChanges,attr"`
			WindowSoftInputMode string `xml:"windowSoftInputMode,attr"`
			IntentFilter        struct {
				Text   string `xml:",chardata"`
				Action []struct {
					Text string `xml:",chardata"`
					Name string `xml:"name,attr"`
				} `xml:"action"`
				Category struct {
					Text string `xml:",chardata"`
					Name string `xml:"name,attr"`
				} `xml:"category"`
			} `xml:"intent-filter"`
		} `xml:"activity"`
		Provider []struct {
			Text                string `xml:",chardata"`
			Name                string `xml:"name,attr"`
			Exported            string `xml:"exported,attr"`
			Authorities         string `xml:"authorities,attr"`
			GrantUriPermissions string `xml:"grantUriPermissions,attr"`
			MetaData            struct {
				Text     string `xml:",chardata"`
				Name     string `xml:"name,attr"`
				Resource string `xml:"resource,attr"`
			} `xml:"meta-data"`
		} `xml:"provider"`
	} `xml:"application"`
}
