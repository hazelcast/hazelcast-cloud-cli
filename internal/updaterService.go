package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var GithubRepository = "hazelcast/hazelcast-cloud-cli"
var Version string
var Distribution string
var binary = "hzcloud"

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
		log.Panic("An error occurred while getting latest release.", latestReleaseErr)
	}

	if !hasNewRelease {
		fmt.Println("Hazelcast Cloud CLI is up to date.")
		return
	}
	fmt.Println("CLI update started.")
	browserDownloadUrl := ""
	for _, asset := range latestRelease.Assets {
		if asset.Name == fmt.Sprintf("%s-%s-%s", binary, runtime.GOOS, runtime.GOARCH) {
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

	currentPath, currentPathErr := v.getCurrentPath()
	if currentPathErr != nil {
		panic("An error occurred while updating.")
	}

	binaryPath := fmt.Sprintf("%s/%s", currentPath, binary)
	moveBinaryErr := os.Rename(binaryPath, fmt.Sprintf("%s.tmp", binaryPath))
	if moveBinaryErr != nil {
		panic("An error occurred while updating.")
	}

	out, createErr := os.Create(binaryPath)
	if createErr != nil {
		_ = os.Rename(fmt.Sprintf("%s.tmp", binaryPath), binaryPath)
		panic("An error occurred while updating.")
	}
	defer out.Close()

	_, copyErr := io.Copy(out, newFile.Body)
	if copyErr != nil {
		_ = os.Rename(fmt.Sprintf("%s.tmp", binaryPath), binaryPath)
		panic("An error occurred while updating.")
	}
	_ = os.Chmod(binary, 0755)
	fmt.Println("CLI update finished.")
	os.Exit(1)
}

func (v UpdateService) Run() {
	v.Clean()
	v.Update()
}

func (v UpdateService) Clean() {
	currentPath, _ := v.getCurrentPath()
	os.Remove(fmt.Sprintf("%s/%s.tmp", currentPath, binary))
}

func (v UpdateService) Check(force bool) {
	if !v.isVersionCheckNeeded() && !force {
		return
	}
	latestRelease, hasNewerVersion, compareErr := v.getLatestRelease()
	if compareErr != nil && force {
		fmt.Println("An error occurred while checking updates")
	}
	v.updateLastVersionCheck()
	if hasNewerVersion {
		bold := color.New(color.Bold)
		cyan := color.New(color.Bold, color.FgHiBlue)
		fmt.Printf("%s has new version -> %s!\nYou can update with ",
			bold.Sprintf("Hazelcast Cloud CLI "), bold.Sprintf(latestRelease.TagName))
		if strings.ToLower(Distribution) == "brew" {
			_, _ = cyan.Println("brew update hzcloud")
		} else {
			_, _ = cyan.Println("hazelcast version update")
		}
	} else if force {
		fmt.Println("Hazelcast Cloud CLI is up to date")
	}
}

func (v UpdateService) getCurrentPath() (string, error) {
	currentDir, currentDirErr := filepath.Abs(filepath.Dir(os.Args[0]))
	if currentDirErr != nil {
		return "", currentDirErr
	}
	return currentDir, nil
}

func (v UpdateService) updateLastVersionCheck() {

}

func (v UpdateService) isVersionCheckNeeded() bool {
	lastVersionCheckTS := v.ConfigService.GetInt64(LastVersionCheckTime)

	if lastVersionCheckTS == 0 {
		v.ConfigService.Set(LastVersionCheckTime, time.Now().Unix())
		return true
	}

	lastVersionCheckTime := time.Unix(lastVersionCheckTS, 0)

	return time.Since(lastVersionCheckTime).Hours() > 24
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
