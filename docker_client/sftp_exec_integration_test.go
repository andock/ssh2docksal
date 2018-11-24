package docker_client

import (
	"os"
	"testing"
)

func getReadOnlyTestContainerId() string {
	handler := CliDockerHandler{}
	containerID, err := handler.Find("ssh2docksal_source_cli_ro_1")
	if err != nil {
		panic("Unable to find ssh2docksal_source_cli container. Run in 'tests/ssh2docksal_source' fin up")
	}
	return containerID
}

func TestExecFileInfo(t *testing.T) {
	if !*testIntegration {
		t.Skip("skipping integration test")
	}

	tests := []struct {
		file     string
		isDir   bool
		error error
	}{

		{file: "/usr/local/bin/php", isDir: false, error: nil},
		{file: "/usr/local/bin", isDir: true, error: nil},
		{file: "/usr/local/NOTEXIST", isDir: true, error: os.ErrNotExist},
	}

	containerID := getTestContainerId()
	root := getRoot(containerID)

	for _, test := range tests {
		file, err := root.execFileInfo(test.file)
		if err != test.error {
			t.Errorf("Lookup for %s should be result in an error", test.file)
			continue
		}
		if err == nil {
			if file.IsDir() != test.isDir {
				t.Errorf("IsDir %s should be %T", test.file, test.isDir)
			}
			if file.name != test.file {
				t.Errorf("Filename: %s should be %s", test.file, file.name)
			}
		}
	}
}

func TestFetch(t *testing.T) {
	if !*testIntegration {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		file  string
		isDir bool
	}{
		{file: "/usr/local/bin", isDir: true},
		{file: "/usr/local/bin/php", isDir: false},
	}
	containerID := getTestContainerId()
	root := getRoot(containerID)
	for _, test := range tests {
		result, _ := root.fetch(test.file)
		if (result == nil) {
			t.Errorf("file %s should exists", test.file)
		}
		if test.isDir == true && result.IsDir() != true {
			t.Errorf("%s should be a folder ", test.file)
		}
		if test.isDir == true && result.IsDir() != true {
			t.Errorf("%s should be a file ", test.file)
		}
	}
}

func TestExecFileList(t *testing.T) {
	if !*testIntegration {
		t.Skip("skipping integration test")
	}

	tests := []struct {
		file   string
		result int
	}{
		{file: "/usr/local/bin", result: 26},
	}
	containerID := getTestContainerId()
	root := getRoot(containerID)
	for _, test := range tests {
		folder,err := root.fetch(test.file)
		if (err != nil) {
			t.Errorf("Fetch file: %s failed.", test.file)
		}
		result, _ := folder.execFileList()
		if len(result) != test.result {
			t.Errorf("Filelist: %s should be %T", test.file, test.result)
		}
		if len(result) != test.result {
			t.Errorf("Filelist: %s should be %T", test.file, test.result)
		}
	}
}

func TestExecFileUpload(t *testing.T) {
	if !*testIntegration {
		t.Skip("skipping integration test")
	}

	tests := []struct {
		sourceFile string
		targetFile string

	}{
		{sourceFile: "../tests/sftp_test/sftp_test.txt", targetFile: testDir + "/sftp_test.txt"},
	}
	containerID := getTestContainerId()

	initSftpTest()

	for _, test := range tests {
		targetFile := newDockerFile(test.targetFile, false, containerID)
		sourceFile, _ := os.Open(test.sourceFile)
		error := targetFile.execFileUpload(sourceFile)
		if error != nil {
			t.Errorf("Unable to upload file %s to %s failed", test.sourceFile, test.targetFile)
		}
		if _, err := os.Stat(test.targetFile); os.IsNotExist(err) {
			t.Errorf("Upload file %s to %s failed. File does not exists", test.sourceFile, test.targetFile)
		}
	}
}
func TestExecFileDownload(t *testing.T) {
	if !*testIntegration {
		t.Skip("skipping integration test")
	}

	tests := []struct {
		localFile  string
		dockerFile string
		checkFile  string
	}{
		{localFile: testDir + "/download.php", dockerFile: "/var/www/docroot/index.php"},
	}
	containerID := getTestContainerId()

	initSftpTest()

	for _, test := range tests {
		targetFile := newDockerFile(test.dockerFile, false, containerID)

		error := targetFile.execFileDownload(test.localFile)
		if error != nil {
			t.Errorf("Unable to download file %s to %s failed", test.dockerFile, test.localFile)
		}
		if _, err := os.Stat(test.localFile); os.IsNotExist(err) {
			t.Errorf("Download file %s to %s failed. File does not exists", test.dockerFile, test.localFile)
		}
	}
}
func TestExecRename(t *testing.T) {
	if !*testIntegration {
		t.Skip("skipping integration test")
	}

	tests := []struct {
		sourceFile string
		targetFile string

	}{
		{sourceFile: testDir + "/test1.txt", targetFile: testDir + "/test1_rename.txt"},
	}
	containerID := getTestContainerId()
	root := getRoot(containerID)
	initSftpTest()

	for _, test := range tests {
		file, fetErr := root.fetch(test.sourceFile)
		if (fetErr != nil) {
			t.Errorf("Fetch file: %s failed.", test.sourceFile)
		}
		error := file.execFileRename(test.targetFile)
		if error != nil {
			t.Errorf("Unable to rename file %s to %s failed", test.sourceFile, test.targetFile)
		}
		if _, err := os.Stat(test.targetFile); os.IsNotExist(err) {
			t.Errorf("Rename file %s to %s failed. File does not exists", test.sourceFile, test.targetFile)
		}
	}
}
