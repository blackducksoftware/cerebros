package jobrunner

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/util"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

type Job struct {
	FromBucket     string `json:"fromBucket"`
	FromBucketPath string `json:"fromBucketPath"`

	ToBucket     string `json:"toBucket"`
	ToBucketPath string `json:"toBucketPath"`
}

type PolarisConfig struct {
	PolarisURL      string
	PolarisEmailID  string
	PolarisPassword string
}

func execSh(shellCmd string) (string, error) {
	execCmd := exec.Command("sh", "-c", shellCmd)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("Command Failed - '%s' - %s - %s", shellCmd, output, err)
	}
	return string(output), err
}

// cleanPath ensures Paths start with '/' and do not end with '/'
func cleanPath(rawPath string) string {
	p := path.Clean(rawPath)
	if p[0:1] != "/" {
		return fmt.Sprintf("/%s", p)
	}
	return p
}

func downloadPolarisCli(envURL string, authHeader map[string]string) (string, error) {
	// TODO change this to linux
	apiURL := fmt.Sprintf("%s/api/tools/polaris_cli-linux64.zip", envURL)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	for key, val := range authHeader {
		req.Header.Set(key, val)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(respBody), int64(len(respBody)))
	if err != nil {
		return "", err
	}

	subDir := ""

	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(f.Name, os.ModePerm)
			subDir = f.Name
			continue
		}
		// Make File
		if err = os.MkdirAll(filepath.Dir(f.Name), os.ModePerm); err != nil {
			return "", err
		}
		outFile, err := os.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
	}

	resp.Body.Close()

	srcDir := cleanPath(subDir)

	cmd := fmt.Sprintf("chmod +x %s/polaris*", srcDir[1:])
	if output, err := execSh(cmd); err != nil {
		return "", fmt.Errorf("%s - %s", output, err)
	}

	absolutePath, err := os.Getwd()
	if err != nil {
		return "", err
	}
	polarisCliPath := fmt.Sprintf("%s%s", absolutePath, srcDir)

	fmt.Printf("polarisCliPath: %s\n", polarisCliPath)

	return polarisCliPath, nil
}

func authenticateUserAndGetCookie(envURL, emailID, password string) (string, error) {
	apiURL := fmt.Sprintf("%s/api/auth/authenticate", envURL)
	form := url.Values{}
	form.Add("email", emailID)
	form.Add("password", password)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create NewRequest: %s", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("POST request failed: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return resp.Header["Set-Cookie"][0], nil
	}
	return "", fmt.Errorf("bad status code: %+v - %+v - %+v", resp.StatusCode, resp.Status, resp.Body)
}

func getPolarisAccessToken(envURL string, authHeader map[string]string) (*string, error) {
	apiURL := fmt.Sprintf("%s/api/auth/apitokens", envURL)
	payload, err := json.Marshal(map[string]map[string]string{
		"data": map[string]string{
			"attributes": "{\"access-token\": None,\"date-created\": None,\"name\": \"test-token-by-api-automation\",\"revoked\": False}",
			"type":       "apitokens",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %s", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create NewRequest: %s", err)
	}
	req.Header.Set("Content-Type", authHeader["Content-Type"])
	req.Header.Set("Cookie", authHeader["Cookie"])

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST request failed: %s", err)
	}
	type RespFormat struct {
		Data struct {
			Attributes struct {
				AccessToken string `json:"access-token"`
			} `json:"attributes"`
		} `json:"data"`
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		bs := RespFormat{}
		err := json.Unmarshal(body, &bs)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal: %s", err)
		}
		finalStr := bs.Data.Attributes.AccessToken
		return &finalStr, nil
	}
	return nil, fmt.Errorf("bad status code: %+v - %+v - %+v", resp.StatusCode, resp.Status, resp.Body)
}

func configurePolarisCliWithAccessToken(envURL, token string) error {
	os.Setenv("POLARIS_SERVER_URL", envURL)
	os.Setenv("POLARIS_ACCESS_TOKEN", token)

	// TODO use linux
	cmd := "polaris_cli-linux64-1.5.5064/bin/polaris configure"
	if output, err := execSh(cmd); err != nil {
		fmt.Printf("error Output: %s\n", output)
		return err
	}
	return nil
}

func scanRepoWithPolarisCli(envURL, repoURL, coverityVersion string) error {
	cmd := fmt.Sprintf("cd /Users/hammer/go/src/github.com/blackducksoftware/JavaVulnerableLab;echo %s | /Users/hammer/go/src/github.com/blackducksoftware/cerebros/pkg/polaris_cli-macosx-1.5.5064/bin/polaris setup", envURL)
	if output, err := execSh(cmd); err != nil {
		fmt.Printf("error Output: %s\n", output)
		return err
	}

	cmd = fmt.Sprintf("sed -i '' s/default/%s/g polaris.yml", coverityVersion)
	if output, err := execSh(cmd); err != nil {
		fmt.Printf("error Output: %s\n", output)
		return err
	}

	cmd = "cd /Users/hammer/go/src/github.com/blackducksoftware/JavaVulnerableLab;/Users/hammer/go/src/github.com/blackducksoftware/cerebros/pkg/polaris_cli-macosx-1.5.5064/bin/polaris analyze"
	if output, err := execSh(cmd); err != nil {
		fmt.Printf("error Output: %s\n", output)
		return err
	}

	cmd = fmt.Sprintf("sed -i '' s/%s/default/g polaris.yml", coverityVersion)
	if output, err := execSh(cmd); err != nil {
		fmt.Printf("error Output: %s\n", output)
		return err
	}

	return nil
}

func polarisCliCapture(repoPath, polarisCliPath string) error {
	cmd := fmt.Sprintf("cd %s;%s/polaris setup", repoPath, polarisCliPath)
	fmt.Printf("Running Polaris Setup\n")
	fmt.Printf("%s\n", cmd)
	if output, err := execSh(cmd); err != nil {
		return fmt.Errorf("%s - %s", output, err)
	}

	cmd = fmt.Sprintf("cd %s;%s/polaris capture", repoPath, polarisCliPath)
	fmt.Printf("Running Polaris Capture\n")
	fmt.Printf("%s\n", cmd)
	if output, err := execSh(cmd); err != nil {
		return fmt.Errorf("%s - %s", output, err)
	}

	cmd = fmt.Sprintf("cd %s;rm polaris.yml", repoPath)
	fmt.Printf("Deleting polaris.yml\n")
	fmt.Printf("%s\n", cmd)
	if output, err := execSh(cmd); err != nil {
		return fmt.Errorf("%s - %s", output, err)
	}
	return nil
}

func scanRepoWithCoverityCli(covBuildCliPath, repoPath, buildToolCommand string) error {
	// example: /Users/hammer/.synopsys/polaris/coverity-tools-macosx-2019.03/bin/cov-configure --java
	cmd := fmt.Sprintf("cd %s;%s/cov-configure --java", repoPath, covBuildCliPath)
	if output, err := execSh(cmd); err != nil {
		fmt.Printf("error out: %s\n", output)
		return err
	}
	// /Users/hammer/.synopsys/polaris/coverity-tools-macosx-2019.03/bin/cov-build --dir myidir /Users/hammer/go/src/github.com/blackducksoftware/hub-fortify-parser/gradlew
	cmd = fmt.Sprintf("cd %s;%s/cov-build --dir myidirs %s", repoPath, covBuildCliPath, buildToolCommand)
	if output, err := execSh(cmd); err != nil {
		fmt.Printf("error out: %s\n", output)
		return err
	}
	return nil
}

func copyToGSBucket(bucketName, serviceAccountPath, bucketFilePath, localFilePath string) error {
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", serviceAccountPath)
	if err != nil {
		return err
	}

	ctx := context.Background()

	f, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	wc := client.Bucket(bucketName).Object(bucketFilePath).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

func copyFromGSBucket(bucketName, serviceAccountPath, bucketFilePath, localFilePath string) error {
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", serviceAccountPath)
	if err != nil {
		return err
	}

	ctx := context.Background()

	f, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	wc, err := client.Bucket(bucketName).Object(bucketFilePath).NewReader(ctx)
	if err != nil {
		return err
	}

	if _, err = io.Copy(f, wc); err != nil {
		return err
	}

	return nil
}

func downloadAndSetupPolarisCli(envURL, emailID, password string) (string, error) {
	fmt.Printf("Authenticating User and Getting Auth Header for requests\n")
	polarisCookie, err := authenticateUserAndGetCookie(envURL, emailID, password)
	if err != nil {
		return "", err
	}
	fmt.Printf("polarisCookie: %+v\n", polarisCookie)

	polarisAuthHeader := &map[string]string{
		"Content-Type": "application/vnd.api+json",
		"Cookie":       polarisCookie,
	}
	fmt.Printf("polarisAuthHeader: %+v\n", polarisAuthHeader)

	fmt.Printf("Downloading Polaris CLI\n")
	polarisCliPath, err := downloadPolarisCli(envURL, *polarisAuthHeader)
	if err != nil {
		return "", err
	}

	accessToken, err := getPolarisAccessToken(envURL, *polarisAuthHeader)
	if err != nil {
		return "", err
	}
	fmt.Printf("accessToken: %+v\n", *accessToken)

	fmt.Printf("Configuring Cli With Access Token\n")
	err = configurePolarisCliWithAccessToken(envURL, *accessToken)
	if err != nil {
		return "", err
	}

	return polarisCliPath, nil
}

func captureBuildAndPushToBucket(fromBucket, fromBucketPath, polarisCliPath, bucketName, pathToBucketServiceAccount, storagePathInBucket string) error {
	// Create a temporary directpry
	// TODO change this. Use emptyDir
	tmpDir, err := ioutil.TempDir("/tmp", "scan")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Downloading from GS Bucket\n")

	// Download the zip archive
	tmpFile := path.Join(tmpDir, "master.zip")
	if err := copyFromGSBucket(fromBucket, pathToBucketServiceAccount, fromBucketPath, tmpFile); err != nil {
		return err
	}

	// Unzip then remove
	if err := util.Unzip(tmpFile, tmpDir); err != nil {
		return err
	}

	fmt.Printf("Running Polaris Capture\n")
	if err := polarisCliCapture(tmpDir, polarisCliPath); err != nil {
		return err
	}

	//pathToIdir := fmt.Sprintf("%s%s", tmpDir, cleanPath(".synopsys/polaris/data/coverity/2019.06-5/idir"))
	pathToIdir := fmt.Sprintf("%s%s", tmpDir, cleanPath(".synopsys/polaris"))
	pathToIdirZip := filepath.Join(tmpDir, "polaris.zip")
	fmt.Printf("Path to IDIR: %s\n", pathToIdir)

	if err := util.Zipit(pathToIdir, pathToIdirZip); err != nil {
		return err
	}

	fmt.Printf("Copying to GS Bucket\n")
	if err := copyToGSBucket(bucketName, pathToBucketServiceAccount, storagePathInBucket, pathToIdirZip); err != nil {
		return err
	}
	fmt.Printf("Scan Complete\n")
	return nil
}

func Start(job Job, polarisConfig PolarisConfig, pathToBucketServiceAccount string) error {
	// coverityVersion := "2019.06"

	polarisCliPath, err := downloadAndSetupPolarisCli(polarisConfig.PolarisURL, polarisConfig.PolarisEmailID, polarisConfig.PolarisPassword)
	if err != nil {
		return err
	}

	err = captureBuildAndPushToBucket(job.FromBucket, job.FromBucketPath, polarisCliPath, job.ToBucket, pathToBucketServiceAccount, job.ToBucketPath)
	if err != nil {
		return err
	}

	return nil
}
