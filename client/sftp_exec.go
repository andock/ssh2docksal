package client

import (
	"archive/tar"
	"github.com/apex/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (folder *dockerFile) execFileList(fs *root) ([]os.FileInfo, error) {
	folderName := folder.name

	nameString, err := outpuExec(folder.containerID, "ls -al "+folderName)
	names := strings.Split(nameString, "\n")
	validItems := []os.FileInfo{}
	first := true
	for _, fn := range names {
		if first {
			first = false
			continue
		}
		item, _ := createNewDockerFile(fn, folder.containerID)
		if item.name != "" && item.name != "." && item.name != ".." {
			seperator := ""
			if folderName != "/" {
				seperator = "/"
			}
			item.name = folderName + seperator + item.name
			fs.files[item.name] = item
			validItems = append(validItems, item)
		}
	}

	return validItems, err
}

func (file *dockerFile) execFileUpload(tarFile *os.File) error {
	cli, err := client.NewEnvClient()
	err = cli.CopyToContainer(context.Background(), file.containerID, filepath.Dir(file.name), tarFile, types.CopyToContainerOptions{})
	return err
}

func (file *dockerFile) execFileDownload() error {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Errorf("Unable to download %s", file.containerID+":"+file.name)
		log.WithError(err)
		return err
	}
	reader, _, err := cli.CopyFromContainer(context.Background(), file.containerID, file.name)
	tr := tar.NewReader(reader)
	for {

		_, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			log.WithError(err)
			return err
		}

		bytes, err := ioutil.ReadAll(tr)
		if err != nil {
			log.WithError(err)
			return err
		}
		file.content = bytes
	}
	return nil
}

func (file *dockerFile) execFileCreate() error {
	return simpleExec(file.containerID, "mkdir -p '" + filepath.Dir(file.name) + "'; cd '" + filepath.Dir(file.name) + "'; touch '"+file.Name() + "'" )
}

func (fs *root) execFileInfo(fileName string) (*dockerFile, error) {
	output, err := outpuExec(fs.containerID, "if [ -e '"+fileName+"' ]; then ls -ald '"+fileName+"'; fi")
	if err != nil {
		return nil, err
	}
	if output == "" {
		return nil, os.ErrNotExist
	}
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines); i++ {
		return createNewDockerFile(lines[i], fs.containerID)
	}
	return nil, os.ErrNotExist
}

func (file *dockerFile) execFileChmod(perm string) error {
	return simpleExec(file.containerID, "chmod "+string(perm)+" "+file.name)
}

func (file *dockerFile) execRemove() error {
	flag := " "
	if file.IsDir() {
		flag = " -r "
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
	return simpleExec(folder.containerID, "mkdir -p "+ folder.name+"/"+folderName)
}
