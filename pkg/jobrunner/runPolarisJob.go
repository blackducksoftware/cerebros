package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
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

	"github.com/blackducksoftware/cerebros/pkg/util"
)

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
	apiURL := fmt.Sprintf("%s/api/tools/polaris_cli-macosx.zip", envURL)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	for key, val := range authHeader {
		req.Header.Set(key, val)
	}
	client := &http.Client{Timeout: time.Second * 10}
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

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("POST request failed: %s", err)
	}
	if resp.StatusCode == 200 {
		defer resp.Body.Close()
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

	client := http.Client{}
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
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		defer resp.Body.Close()
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

	cmd := "polaris_cli-macosx-1.5.5064/bin/polaris configure"
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

	cmd := exec.Command("gsutil", "cp", "-r", localFilePath, fmt.Sprintf("gs://%s%s", bucketName, bucketFilePath))
	fmt.Printf("%s\n", cmd)
	if info, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("error out: %s\n", info)
		return err
	}

	return nil
}

func copyFromGSBucket(bucketName, serviceAccountPath, bucketFilePath, localFilePath string) error {
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", serviceAccountPath)
	if err != nil {
		return err
	}

	cmd := exec.Command("gsutil", "cp", fmt.Sprintf("gs://%s%s", bucketName, bucketFilePath), fmt.Sprintf("%s", localFilePath))
	if info, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("error out: %s\n", info)
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

func captureBuildAndPushToBucket(repoPath, polarisCliPath, bucketName, pathToBucketServiceAccount, storagePathInBucket string) error {
	fmt.Printf("Running Polaris Capture\n")
	err := polarisCliCapture(repoPath, polarisCliPath)
	if err != nil {
		return err
	}

	pathToIdir := fmt.Sprintf("%s%s", repoPath, cleanPath(".synopsys/polaris/data/coverity/2019.06-5/idir"))
	fmt.Printf("Path to IDIR: %s\n", pathToIdir)

	fmt.Printf("Copying to GS Bucket\n")
	err = copyToGSBucket(bucketName, pathToBucketServiceAccount, storagePathInBucket, pathToIdir)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// coverityVersion := "2019.06"
	polarisURL := ""
	polarisEmailID := ""
	poalrisPassword := ""
	polarisCliPath, err := downloadAndSetupPolarisCli(polarisURL, polarisEmailID, poalrisPassword)
	if err != nil {
		fmt.Printf("[ERROR] %s\n", err)
		return
	}

	queue := util.NewPriorityQueue()

	type job struct {
		RepoPath                   string
		BucketName                 string
		PathToBucketServiceAccount string
		StoragePathInBucket        string
	}

	job1 := job{
		RepoPath:                   cleanPath("/Users/hammer/go/src/github.com/blackducksoftware/JavaVulnerableLab"),
		BucketName:                 "indian-terminator",
		PathToBucketServiceAccount: cleanPath("/Users/hammer/Downloads/polaris-dev-233821-b8a3ac17ca0f.json"),
		StoragePathInBucket:        cleanPath("bucket-idir-job1"),
	}
	job2 := job{
		RepoPath:                   cleanPath("/Users/hammer/go/src/github.com/blackducksoftware/hub-fortify-parser"),
		BucketName:                 "indian-terminator",
		PathToBucketServiceAccount: cleanPath("/Users/hammer/Downloads/polaris-dev-233821-b8a3ac17ca0f.json"),
		StoragePathInBucket:        cleanPath("bucket-idir-job2"),
	}
	job3 := job{
		RepoPath:                   cleanPath("/Users/hammer/go/src/github.com/blackducksoftware/rabbitmq"),
		BucketName:                 "indian-terminator",
		PathToBucketServiceAccount: cleanPath("/Users/hammer/Downloads/polaris-dev-233821-b8a3ac17ca0f.json"),
		StoragePathInBucket:        cleanPath("bucket-idir-job3"),
	}
	job4 := job{
		RepoPath:                   cleanPath("/Users/hammer/go/src/github.com/blackducksoftware/polaris-deploy-sanity"),
		BucketName:                 "indian-terminator",
		PathToBucketServiceAccount: cleanPath("/Users/hammer/Downloads/polaris-dev-233821-b8a3ac17ca0f.json"),
		StoragePathInBucket:        cleanPath("bucket-idir-job4"),
	}
	job5 := job{
		RepoPath:                   cleanPath("/Users/hammer/go/src/github.com/blackducksoftware/synopsys-operator"),
		BucketName:                 "indian-terminator",
		PathToBucketServiceAccount: cleanPath("/Users/hammer/Downloads/polaris-dev-233821-b8a3ac17ca0f.json"),
		StoragePathInBucket:        cleanPath("bucket-idir-job5"),
	}

	queue.Add("job1", 2, job1)
	queue.Add("job2", 3, job2)
	queue.Add("job3", 1, job3)
	queue.Add("job4", 7, job4)
	queue.Add("job5", 7, job5)

	for {
		fmt.Printf("Queue has %d jobs...\n", queue.Size())
		if queue.IsEmpty() {
			return
		}
		jInterface, err := queue.Pop()
		if err != nil {
			fmt.Printf("[ERROR] %s", err)
			return
		}
		if jInterface != nil {
			j := jInterface.(job)
			err = captureBuildAndPushToBucket(j.RepoPath, polarisCliPath, j.BucketName, j.PathToBucketServiceAccount, j.StoragePathInBucket)
			if err != nil {
				fmt.Printf("[ERROR] %s\n", err)
			}
		}
	}
}
