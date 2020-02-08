package jobrunner

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

type PolarisConfig struct {
	PolarisURL   string
	PolarisToken string
}

type PolarisScanner struct {
	config PolarisConfig
	wd     string
}

func NewPolarisScanner(config PolarisConfig) (*PolarisScanner, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	os.Setenv("POLARIS_USER_INPUT_TIMEOUT_MINUTES", "1")
	if err := configurePolarisCliWithAccessToken(config.PolarisURL, config.PolarisToken, wd); err != nil {
		return nil, err
	}
	return &PolarisScanner{config: config, wd: wd}, nil
}

func (p *PolarisScanner) Capture(capturePath string) (string, error) {
	fmt.Printf("Running Polaris Capture\n")
	if output, err := execSh("polaris capture", capturePath); err != nil {
		return "", fmt.Errorf("%s - %s", output, err)
	}

	fmt.Printf("Deleting polaris.yml\n")
	if err := os.Remove(path.Join(capturePath, "polaris.yml")); err != nil {
		return "", err
	}

	return path.Join(capturePath, ".synopsys/polaris/data/coverity/2019.06-5/idir"), nil
}

func (p *PolarisScanner) Scan(repoPath, idirPath string) error {
	if output, err := execSh("polaris setup", repoPath); err != nil {
		return fmt.Errorf("%s - %s", output, err)
	}

	polarisYmlPath := path.Join(repoPath, "polaris.yml")
	content, err := ioutil.ReadFile(polarisYmlPath)
	if err != nil {
		return err
	}

	CoverityConfig := make(map[string]interface{})
	if err := yaml.Unmarshal(content, CoverityConfig); err != nil {
		return err
	}
	// TODO - Need to get the CoverityConfig struct from somewhere :/
	CoverityConfig["capture"] = map[string]interface{}{
		"coverity": map[string]interface{}{
			"idirCapture": map[string]interface{}{
				"idirPath": idirPath,
			},
		},
	}
	CoverityConfigYaml, err := yaml.Marshal(CoverityConfig)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(polarisYmlPath, CoverityConfigYaml, os.FileMode(700)); err != nil {
		return err
	}

	if output, err := execSh("polaris analyze", repoPath); err != nil {
		return fmt.Errorf("%s - %s", output, err)
	}

	return nil
}

func execSh(shellCmd, workdir string) (string, error) {
	execCmd := exec.Command("sh", "-c", shellCmd)
	if len(workdir) > 0 {
		execCmd.Dir = path.Clean(workdir)
	}

	output, err := execCmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("Command Failed - '%s' - %s - %s", shellCmd, output, err)
	}
	return string(output), err
}

func configurePolarisCliWithAccessToken(envURL, token, wd string) error {
	os.Setenv("POLARIS_SERVER_URL", envURL)
	os.Setenv("POLARIS_ACCESS_TOKEN", token)

	cmd := "polaris configure"
	if output, err := execSh(cmd, wd); err != nil {
		fmt.Printf("error Output: %s\n", output)
		return err
	}
	return nil
}

//func captureBuildAndPushToBucket(fromBucket, fromBucketPath, polarisCliPath, bucketName, pathToBucketServiceAccount, storagePathInBucket string) error {
//	// Create a temporary directpry
//	// TODO change this. Use emptyDir
//	tmpDir, err := ioutil.TempDir("/tmp", "scan")
//	if err != nil {
//		return err
//	}
//	defer os.RemoveAll(tmpDir)
//
//	fmt.Printf("Downloading from GS Bucket\n")
//
//	// Download the zip archive
//	tmpFile := path.Join(tmpDir, "master.zip")
//	if err := copyFromGSBucket(fromBucket, pathToBucketServiceAccount, fromBucketPath, tmpFile); err != nil {
//		return err
//	}
//
//	// Unzip then remove
//	if err := util.Unzip(tmpFile, tmpDir); err != nil {
//		return err
//	}
//
//	fmt.Printf("Running Polaris Capture\n")
//	if err := polarisCliCapture(tmpDir, polarisCliPath); err != nil {
//		return err
//	}
//
//	pathToIdir := fmt.Sprintf("%s%s", tmpDir, cleanPath(".synopsys/polaris/data/coverity/2019.06-5/idir"))
//	pathToIdirZip := filepath.Join(tmpDir, "idir.zip")
//	fmt.Printf("Path to IDIR: %s\n", pathToIdir)
//
//	if err := util.Zipit(pathToIdir, pathToIdirZip); err != nil {
//		return err
//	}
//
//	fmt.Printf("Copying to GS Bucket\n")
//	if err := copyToGSBucket(bucketName, pathToBucketServiceAccount, storagePathInBucket, pathToIdirZip); err != nil {
//		return err
//	}
//	fmt.Printf("Scan Complete\n")
//	return nil
//}

//func downloadPolarisCli(envURL string, authHeader map[string]string) (string, error) {
//	// TODO change this to linux
//	apiURL := fmt.Sprintf("%s/api/tools/polaris_cli-linux64.zip", envURL)
//	req, err := http.NewRequest("GET", apiURL, nil)
//	if err != nil {
//		return "", err
//	}
//	for key, val := range authHeader {
//		req.Header.Set(key, val)
//	}
//
//	client := http.Client{
//		Timeout: time.Second * 10,
//	}
//
//	resp, err := client.Do(req)
//	if err != nil {
//		return "", err
//	}
//
//	respBody, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return "", err
//	}
//
//	zipReader, err := zip.NewReader(bytes.NewReader(respBody), int64(len(respBody)))
//	if err != nil {
//		return "", err
//	}
//
//	subDir := ""
//
//	for _, f := range zipReader.File {
//		if f.FileInfo().IsDir() {
//			// Make Folder
//			os.MkdirAll(f.Name, os.ModePerm)
//			subDir = f.Name
//			continue
//		}
//		// Make File
//		if err = os.MkdirAll(filepath.Dir(f.Name), os.ModePerm); err != nil {
//			return "", err
//		}
//		outFile, err := os.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
//		if err != nil {
//			return "", err
//		}
//		rc, err := f.Open()
//		if err != nil {
//			return "", err
//		}
//		_, err = io.Copy(outFile, rc)
//		outFile.Close()
//		rc.Close()
//	}
//
//	resp.Body.Close()
//
//	srcDir := cleanPath(subDir)
//
//	cmd := fmt.Sprintf("chmod +x %s/polaris*", srcDir[1:])
//	if output, err := execSh(cmd); err != nil {
//		return "", fmt.Errorf("%s - %s", output, err)
//	}
//
//	absolutePath, err := os.Getwd()
//	if err != nil {
//		return "", err
//	}
//	polarisCliPath := fmt.Sprintf("%s%s", absolutePath, srcDir)
//
//	fmt.Printf("polarisCliPath: %s\n", polarisCliPath)
//
//	return polarisCliPath, nil
//}

/*func authenticateUserAndGetCookie(envURL, emailID, password string) (string, error) {
	apiURL := fmt.Sprintf("%s/api/auth/authenticate", envURL)
	form := url.Values{}
	form.Add("email", emailID)
	form.Add("password", password)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create NewRequest: %s", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{
		Timeout: time.Second * 10,
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
*/

//
//func scanRepoWithCoverityCli(covBuildCliPath, repoPath, buildToolCommand string) error {
//	// example: /Users/hammer/.synopsys/polaris/coverity-tools-macosx-2019.03/bin/cov-configure --java
//	cmd := fmt.Sprintf("cd %s;%s/cov-configure --java", repoPath, covBuildCliPath)
//	if output, err := execSh(cmd); err != nil {
//		fmt.Printf("error out: %s\n", output)
//		return err
//	}
//	// /Users/hammer/.synopsys/polaris/coverity-tools-macosx-2019.03/bin/cov-build --dir myidir /Users/hammer/go/src/github.com/blackducksoftware/hub-fortify-parser/gradlew
//	cmd = fmt.Sprintf("cd %s;%s/cov-build --dir myidirs %s", repoPath, covBuildCliPath, buildToolCommand)
//	if output, err := execSh(cmd); err != nil {
//		fmt.Printf("error out: %s\n", output)
//		return err
//	}
//	return nil
//}
