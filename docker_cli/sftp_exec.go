package docker_cli

import (
	"github.com/apex/log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func simpleExec(containerID string, command string ) error {
	args := execGetBaseArgs(containerID)
	args = append(args, command)
	log.Debugf("Execute %s", command)
	cmd := exec.Command("docker", args...)
	err := cmd.Run()
	if err != nil {
		log.Errorf("Unable to execute %s", command)
		log.WithError(err)
	}
	return err
}

func execGetBaseArgs(containerID string) []string {
	args := []string{"exec"}
	args = append(args, "-u")
	args = append(args, "docker")
	args = append(args, containerID)
	args = append(args, "bash")
	args = append(args, "-c")
	return args
}

func (folder *dockerFile) execFileList(fs *root) ([]os.FileInfo, error) {
	folderName := folder.name
	args := execGetBaseArgs(folder.containerID)
	args = append(args, "ls -a " + folderName)
	cmd := exec.Command("docker", args...)
	outputBytes, err := cmd.CombinedOutput()
	names := strings.Split(string(outputBytes), "\n")
	valid_names := []string{}
	for _, fn := range names {
		if fn != "" && fn != "." && fn != ".." {
			valid_names = append(valid_names, fn)
		}
	}
	sort.Strings(valid_names)
	list := make([]os.FileInfo, len(valid_names))
	for i, fn := range valid_names {
		seperator := ""
		if folderName != "/" {
			seperator = "/"
		}
		item, _ := fs.fetch(folderName + seperator + fn)
		list[i] = item

	}
	return list, err
}

func (file *dockerFile) execFileUpload(localFile *os.File) error {
	args := []string{"cp"}
	args = append(args, localFile.Name())
	args = append(args, file.containerID+":"+file.name)
	cmd := exec.Command("docker", args...)
	err := cmd.Run()
	if err != nil {
		log.Errorf("Unable to upload %s", localFile.Name())
		log.WithError(err)
	}
	return err
}

func (file *dockerFile) execFileDownload(localFilePath string) error {
	args := []string{"cp"}
	args = append(args, file.containerID+":"+file.name)
	args = append(args, localFilePath)
	cmd := exec.Command("docker", args...)
	err := cmd.Run()
	if err != nil {
		log.Errorf("Unable to download %s", file.containerID + ":" + file.name)
		log.WithError(err)
	}
	return err
}

func (file *dockerFile) execFileCreate() error {
	return simpleExec(file.containerID, "cd " + filepath.Dir(file.name) +"; touch " + file.Name())
}

func (fs *root) execFileInfo(fileName string, modifier string) (bool, error) {
	args := execGetBaseArgs(fs.containerID)
	args = append(args, "if [ -"+modifier+" "+fileName+" ]; then echo 1; else echo 0; fi")
	cmd := exec.Command("docker", args...)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Unable to get FileInfo for %s", fileName)
		log.WithError(err)
	}
	output := strings.TrimSpace(string(outputBytes))
	if output == "1" {
		return true, err
	} else if output == "0" {
		return false, err
	}
	return false, err
}

func (file *dockerFile) execFileChmod( perm string) error {
	return simpleExec(file.containerID, "chmod " + string(perm) + " " + file.name )
}

func (file *dockerFile) remove() error {
	flag := " "
	if (file.IsDir()) {
		flag =" -r "
	}
	return simpleExec(file.containerID, "rm" + flag + file.name)
}

func (file *dockerFile) execFileRename(targetName string) error {
	return simpleExec(file.containerID, "mv "+file.name+" "+targetName)
}

func (file *dockerFile) execTruncate(size uint64) error {
	return simpleExec(file.containerID, "truncate -s " + strconv.FormatUint(size, 10) + " " + file.name)
}

func (folder *dockerFile) execMkDir(folderName string) error {
	return simpleExec(folder.containerID, "cd "+folder.name+"; mkdir -p "+folderName)
}

