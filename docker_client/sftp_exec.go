package docker_client

import (
	"github.com/apex/log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)


func (folder *dockerFile) execFileList() ([]os.FileInfo, error) {
	folderName := folder.name


	nameString, err := outpuExec(folder.containerID, "ls -al " + folderName)
	names := strings.Split(nameString, "\n")
	valid_names := []string{}
	first := true
	for _, fn := range names {
		if (first) {
			first = false
			continue
		}
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
		item, _ := createNewDockerFile(fn, folder.containerID)
		item.name = folderName + seperator + item.name
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
	args = append(args, file.containerID + ":" + file.name)
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

func (fs *root) execFileInfo(fileName string) (*dockerFile, error) {

	output, err := outpuExec(fs.containerID, "ls -ald " + fileName)
	if (err != nil) {
		return nil, os.ErrNotExist
	}
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines); i++ {
		return createNewDockerFile(lines[i], fs.containerID)
	}
	return nil, os.ErrNotExist
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
	return simpleExec(folder.containerID, "mkdir -p "+folder.name + "/" +folderName)
}

