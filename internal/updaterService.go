package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

var GithubRepository = "hazelcast/hazelcast-cloud-cli"
var Version string
var Distribution string

type UpdateService struct {
	ConfigService ConfigService
}

func NewUpdaterService() UpdateService {
	return UpdateService{
		ConfigService: NewConfigService(),
	}
}
func (v UpdateService) Update() {
	if strings.ToLower(Distribution) == "brew" {
		println("You can not update brew package with this command, please use `brew upgrade hzcloud`")
		return
	}
	latestRelease, hasNewRelease, latestReleaseErr := v.getLatestRelease()
	if latestReleaseErr != nil {
		fmt.Println("An error occurred while getting latest release.")
		return
	}
	if !hasNewRelease {
		fmt.Println("Hazelcast Cloud CLI is up to date.")
		return
	}
	fmt.Println("CLI update started.")
	browserDownloadUrl := ""
	for _, asset := range latestRelease.Assets {
		if asset.Name == fmt.Sprintf("hzcloud-%s-%s", runtime.GOOS, runtime.GOARCH) {
			browserDownloadUrl = asset.BrowserDownloadUrl
		}
	}
	if browserDownloadUrl == "" {
		panic("An error occurred while getting latest release.")
	}
	newFile, err := http.Get(browserDownloadUrl)
	if err != nil {
		panic("An error occurred while getting latest release.")
	}
	defer newFile.Body.Close()
	executablePath, executablePathErr := v.getExecutablePath()
	if executablePathErr != nil {
		panic("An error occurred while updating.")
	}
	moveBinaryErr := os.Rename(executablePath, fmt.Sprintf("%s.tmp", executablePath))
	if moveBinaryErr != nil {
		panic(fmt.Sprintf("An error occurred while updating. %s", moveBinaryErr))
	}
	out, createErr := os.Create(executablePath)
	if createErr != nil {
		_ = os.Rename(fmt.Sprintf("%s.tmp", executablePath), executablePath)
		panic("An error occurred while updating.")
	}
	defer out.Close()
	_, copyErr := io.Copy(out, newFile.Body)
	if copyErr != nil {
		_ = os.Rename(fmt.Sprintf("%s.tmp", executablePath), executablePath)
		panic("An error occurred while updating.")
	}
	_ = os.Chmod(executablePath, 0755)
	fmt.Println("CLI update finished.")
	os.Exit(1)
}
func (v UpdateService) Run() {
	v.Clean()
	v.Update()
}
func (v UpdateService) Clean() {
	currentPath, _ := v.getExecutablePath()
	_ = os.Remove(fmt.Sprintf("%s.tmp", currentPath))
}
func (v UpdateService) Check(force bool) {
	if !v.isVersionCheckNeeded() && !force {
		return
	}
	latestRelease, hasNewerVersion, compareErr := v.getLatestRelease()
	bold := color.New(color.Bold)
	cyan := color.New(color.Bold, color.FgHiBlue)
	if compareErr != nil && force {
		fmt.Println("An error occurred while checking updates.")
		return
	}
	if hasNewerVersion {
		fmt.Printf("%s\nCurrent Version: %s\nNew version: %s\nYou can update with ",
			cyan.Sprintf("Hazelcast Cloud CLI"), bold.Sprintf(Version), bold.Sprintf(latestRelease.TagName))
		if strings.ToLower(Distribution) == "brew" {
			_, _ = cyan.Println("brew upgrade hzcloud")
		} else {
			_, _ = bold.Println("hzcloud version update")
		}
	} else if force {
		fmt.Printf("%s %s is up to date.\n", bold.Sprintf("Hazelcast Cloud CLI"), bold.Sprintf(Version))
	}
}
func (v UpdateService) getExecutablePath() (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return executablePath, nil
}
func (v UpdateService) isVersionCheckNeeded() bool {
	lastVersionCheckTS := v.ConfigService.GetInt64(LastVersionCheckTime)
	lastVersionCheckTime := time.Unix(lastVersionCheckTS, 0)
	isNeeded := time.Since(lastVersionCheckTime).Hours() > 24
	if isNeeded {
		v.ConfigService.Set(LastVersionCheckTime, time.Now().Unix())
	}
	return isNeeded
}
func (v UpdateService) getLatestRelease() (Release, bool, error) {
	resp, respErr := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GithubRepository))
	if respErr != nil {
		return Release{}, false, respErr
	}
	respBody, respBodyErr := ioutil.ReadAll(resp.Body)
	if respBodyErr != nil {
		return Release{}, false, respBodyErr
	}
	var latestRelease Release
	unmarshalErr := json.Unmarshal(respBody, &latestRelease)
	if unmarshalErr != nil {
		return Release{}, false, unmarshalErr
	}
	if latestRelease.TagName == "" {
		return Release{}, false, errors.New("latest release not found")
	}
	currentVersion, semVerErr := semver.ParseTolerant(Version)
	if semVerErr != nil {
		return latestRelease, false, semVerErr
	}
	releaseVersion, semVerErr := semver.ParseTolerant(latestRelease.TagName)
	if semVerErr != nil {
		return latestRelease, false, semVerErr
	}
	return latestRelease, releaseVersion.GT(currentVersion), nil
}

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}
