package docker_cli

import (
	"os"
	"testing"
)

func getTestContainerId()(string) {
	handler := CliDockerHandler{}
	containerID, err := handler.Find("ssh2docksal_source_cli")
	if (err != nil) {
		panic("Unable to find ssh2docksal_source_cli container. Run in 'tests/ssh2docksal_source' fin up")
	}
	return containerID
}

func initSftpTest () {
	if _, err := os.Stat("../tests/ssh2docksal_source/docroot/sftp_test"); !os.IsNotExist(err) {
		os.RemoveAll("../tests/ssh2docksal_source/docroot/sftp_test")
	}
	os.Mkdir("../tests/ssh2docksal_source/docroot/sftp_test", os.ModePerm)
	os.Mkdir("../tests/ssh2docksal_source/docroot/sftp_test/dummy", os.ModePerm)
}

func TestFetch(t *testing.T) {
	tests := []struct{
		file      string
		isDir      bool
	}{
		{ file: "/usr/local/bin", isDir: true},
		{ file: "/usr/local/bin/php", isDir: false},
	}
	containerID := getTestContainerId()
	root := getRoot(containerID)
	for _, test := range tests {
		result, _ := root.fetch(test.file)
		if (test.isDir == true && result.IsDir() != true) {
			t.Errorf("%s should be a folder ", test.file)
		}
		if (test.isDir == true && result.IsDir() != true) {
			t.Errorf("%s should be a file ", test.file)
		}
	}
}

func TestExecFileList(t *testing.T) {

	tests := []struct{
		file      string
		result      int
	}{
		{ file: "/usr/local/bin", result: 26},
	}
	containerID := getTestContainerId()
	root := getRoot(containerID)
	for _, test := range tests {
		result, _ := root.execFileList(test.file)
		if (len(result) != test.result) {
			t.Errorf("Filelist: %s should be %b", test.file, test.result)
		}
	}
}

func TestExecFileUpload(t *testing.T) {
	tests := []struct{
		sourceFile      string
		targetFile      string
		checkFile      string
	}{
		{ sourceFile: "../tests/sftp_test/sftp_test.txt", targetFile: "/var/www/docroot/sftp_test/sftp_test.txt", checkFile: "../tests/ssh2docksal_source/docroot/sftp_test/sftp_test.txt"},
	}
	containerID := getTestContainerId()

	initSftpTest()

	for _, test := range tests {
		targetFile := newDockerFile(test.targetFile, false, containerID)
		sourceFile, _ := os.Open(test.sourceFile)
		error := targetFile.execFileUpload(sourceFile)
		if (error != nil) {
			t.Errorf("Unable to upload file %s to %s failed", test.sourceFile, test.targetFile)
		}
		if _, err := os.Stat(test.checkFile); os.IsNotExist(err) {
			t.Errorf("Upload file %s to %s failed. File does not exists", test.sourceFile, test.targetFile)
		}
	}
}
func TestExecFileDownload(t *testing.T) {
	tests := []struct{
		localFile      string
		dockerFile      string
		checkFile      string
	}{
		{ localFile: "../tests/sftp_test/index.php", dockerFile: "/var/www/docroot/index.php"},
	}
	containerID := getTestContainerId()

	initSftpTest()

	for _, test := range tests {
		targetFile := newDockerFile(test.dockerFile, false, containerID)

		error := targetFile.execFileDownload(test.localFile)
		if (error != nil) {
			t.Errorf("Unable to download file %s to %s failed", test.dockerFile, test.localFile)
		}
		if _, err := os.Stat(test.localFile); os.IsNotExist(err) {
			t.Errorf("Download file %s to %s failed. File does not exists", test.dockerFile, test.localFile)
		}
	}
}
func TestExecRename(t *testing.T) {
	tests := []struct{
		sourceFile      string
		targetFile      string
		checkFile      string
	}{
		{ checkFile: "../tests/ssh2docksal_source/docroot/sftp_test/dummy_rename", sourceFile: "/var/www/docroot/sftp_test/dummy", targetFile: "/var/www/docroot/sftp_test/dummy_rename"},
	}
	containerID := getTestContainerId()
	root := getRoot(containerID)
	initSftpTest()

	for _, test := range tests {
		error := root.execFileRename(test.sourceFile, test.targetFile)
		if (error != nil) {
			t.Errorf("Unable to rename file %s to %s failed", test.sourceFile, test.targetFile)
		}
		if _, err := os.Stat(test.checkFile); os.IsNotExist(err) {
			t.Errorf("Rename file %s to %s failed. File does not exists", test.sourceFile, test.targetFile)
		}
	}
}
func TestExecFileInfo(t *testing.T) {

	tests := []struct{
		file      string
		result      bool
		modifier      string

	}{
		{ file: "/usr/local/bin/php", modifier: "d", result: false},
		{ file: "/usr/local/bin/php", modifier: "f", result: true},
		{ file: "/usr/local/bin", modifier: "d", result: true},
		{ file: "/usr/local/bin", modifier: "f", result: false},
	}
	containerID := getTestContainerId()
	root := getRoot(containerID)

	for _, test := range tests {
		result, _ := root.execFileInfo(test.file, test.modifier)
		if (result != test.result) {
			t.Errorf("File: %s with modifier %s should be %T", test.file, test.modifier, test.result)
		}
	}
}