package docker_cli


// This serves as an example of how to implement the request server handler as
// well as a dummy backend for testing. It implements an in-memory backend that
// works as a very simple filesystem with simple flat key-value lookup system.

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

func getRoot(containerID string) (*root){
	root := &root{
		files: make(map[string]*dockerFile),
		containerID: containerID,
	}
	root.dockerFile = newDockerFile("/", true, root.containerID)
	return root
}
// DocCliHandler returns a Hanlders object for docker cli.
func DockerCliSftpHandler(containerID string) sftp.Handlers {
	root := getRoot(containerID)
	return sftp.Handlers{root, root, root, root}
}

func (fs *dockerFile) execFileUpload(localFile *os.File) (error) {
	args := []string{"cp"}
	args = append(args, localFile.Name())
	args = append(args, fs.containerID + ":" + fs.name)
	cmd := exec.Command("docker", args...)
	_, err := cmd.CombinedOutput()
	return err;
}


func (fs *dockerFile) execFileDownload(localFilePath string) (error) {
	args := []string{"cp"}
	args = append(args, fs.containerID + ":" + fs.name)
	args = append(args, localFilePath)
	cmd := exec.Command("docker", args...)
	_, err := cmd.CombinedOutput()
	return err;
}

func (fs *root) execFileInfo(fileName string, modifier string) (bool, error) {
	args := getBaseArgs(fs.containerID)
	args = append(args, "if [ -"+ modifier + " " + fileName + " ]; then echo 1; else echo 0; fi")
	cmd := exec.Command("docker", args...)
	outputBytes, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(outputBytes))
	if (output == "1") {
		return true, err;
	} else if (output == "0") {
		return false, err;
	}
	return false, err;
}

func (fs *root) execFileRename(sourceName string, targetName string) (error) {
	args := getBaseArgs(fs.containerID)
	args = append(args, "mv " + sourceName + " " + targetName)
	cmd := exec.Command("docker", args...)
	_, err := cmd.CombinedOutput()
	return err;
}

func (fs *root) execMkDir(folderName string, rootFolder string) (error) {
	args := getBaseArgs(fs.containerID)
	args = append(args, "cd " + rootFolder + "; mkdir " + folderName)
	cmd := exec.Command("docker", args...)
	_, err := cmd.CombinedOutput()
	return err;
}


func getBaseArgs(containerID string) ([]string) {
	args := []string{"exec"}
	args = append(args, "-u")
	args = append(args, "docker")
	args = append(args, containerID)
	args = append(args, "bash")
	args = append(args, "-c")
	return args
}

func (fs *root) execFileList(folderName string) ([]os.FileInfo, error) {
	args := getBaseArgs(fs.containerID)
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
		if (folderName != "/") {
			seperator = "/"
		}
		item, _ := fs.fetch(folderName + seperator + fn)
		list[i] = item

	}
	return list, err;
}

func (fs *root) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	if fs.mockErr != nil {
		return nil, fs.mockErr
	}
	fs.filesLock.Lock()
	defer fs.filesLock.Unlock()
	file, err := fs.fetch(r.Filepath)
	uuid, err := uuid.NewRandom()
	tmpFileFile := os.TempDir() + "/" + uuid.String()
	file.execFileDownload(tmpFileFile)
	f, err := os.Open(tmpFileFile)
	b1 := make([]byte, 5)
	f.Read(b1)
	if err != nil {
		return nil, err
	}
	file.content = b1
	os.Remove(tmpFileFile)

	if file.symlink != "" {
		file, err = fs.fetch(file.symlink)
		if err != nil {
			return nil, err
		}
	}
	return file.ReaderAt()
}

func (fs *root) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	if fs.mockErr != nil {
		return nil, fs.mockErr
	}
	fs.filesLock.Lock()
	defer fs.filesLock.Unlock()
	file, err := fs.fetch(r.Filepath)
	if err == os.ErrNotExist {
		dir, err := fs.fetch(filepath.Dir(r.Filepath))
		if err != nil {
			return nil, err
		}
		if !dir.isdir {
			return nil, os.ErrInvalid
		}
		file = newDockerFile(r.Filepath, false, fs.containerID)
		fs.files[r.Filepath] = file
	}
	return file.WriterAt()
}

func (fs *root) Filecmd(r *sftp.Request) error {
	if fs.mockErr != nil {
		return fs.mockErr
	}
	fs.filesLock.Lock()
	defer fs.filesLock.Unlock()
	switch r.Method {
	case "Setstat":
		return nil
	case "Rename":
		err := fs.execFileRename(r.Filepath, r.Target)
		if err != nil {
			return err
		}
	case "Rmdir", "Remove":
		_, err := fs.fetch(filepath.Dir(r.Filepath))
		if err != nil {
			return err
		}
		delete(fs.files, r.Filepath)
	case "Mkdir":
		folder, err := fs.fetch(filepath.Dir(r.Filepath))
		if err != nil {
			return err
		}
		err = fs.execMkDir(filepath.Base(r.Filepath), folder.name)
		if err != nil {
			return err
		}
	case "Symlink":
		// Not implemented
	}
	return nil
}

type listerat []os.FileInfo

// Modeled after strings.Reader's ReadAt() implementation
func (f listerat) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	var n int
	if offset >= int64(len(f)) {
		return 0, io.EOF
	}
	n = copy(ls, f[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}


func (fs *root) Filelist(r *sftp.Request) (sftp.ListerAt, error) {

	if fs.mockErr != nil {
		return nil, fs.mockErr
	}
	fs.filesLock.Lock()
	defer fs.filesLock.Unlock()
	path := r.Filepath
	if r.Filepath == "/" {
		path = fs.dockerFile.name
	}

	switch r.Method {
	case "List":
		list, err := fs.execFileList(path)
		return listerat(list), err
	case "Stat":
		file, err := fs.fetch(path)
		if err != nil {
			return nil, err
		}
		return listerat([]os.FileInfo{file}), nil

		//
	case "Readlink":
		file, err := fs.fetch(r.Filepath)
		if err != nil {
			return nil, err
		}
		if file.symlink != "" {
			file, err = fs.fetch(file.symlink)
			if err != nil {
				return nil, err
			}
		}
		return listerat([]os.FileInfo{file}), nil
	}
	return nil, nil
}

// In memory file-system-y thing that the Hanlders live on
type root struct {
	*dockerFile
	files     map[string]*dockerFile
	filesLock sync.Mutex
	mockErr   error
	containerID string
}

// Set a mocked error that the next handler call will return.
// Set to nil to reset for no error.
func (fs *root) returnErr(err error) {
	fs.mockErr = err
}

func (fs *root) fetch(path string) (*dockerFile, error) {
	if path == "/" {
		return fs.dockerFile, nil
	}
	hasPos := strings.Index(path, ".")
	check1 := "d"
	check2 := "f"
	if hasPos != -1 {
		check1 = "f"
		check2 = "d"
	}
	check1_result, err := fs.execFileInfo(path, check1)
	if err != nil {
		return nil, err
	}
	if (check1_result == true && check1 == "d") {
		file := newDockerFile(path, true, fs.containerID);
		return file, nil
	}
	if (check1_result == true && check1 == "f") {
		file := newDockerFile(path, false, fs.containerID);
		return file, nil
	}

	check2_result, err := fs.execFileInfo(path, check2)
	if (err != nil) {
		return nil, err
	}
	if (check2_result == true && check2 == "d") {
		file := newDockerFile(path, true, fs.containerID);
		return file, nil
	}
	if (check2_result == true && check2 == "f") {
		file := newDockerFile(path, false, fs.containerID);
		return file, nil
	}
	return nil, os.ErrNotExist
}

// Implements os.FileInfo, Reader and Writer interfaces.
// These are the 3 interfaces necessary for the Handlers.
type dockerFile struct {
	name        string
	modtime     time.Time
	symlink     string
	isdir       bool
	content     []byte
	contentLock sync.RWMutex
	containerID string
}

// factory to make sure modtime is set
func newDockerFile(name string, isdir bool, containerID string) *dockerFile {
	return &dockerFile{
		name:    name,
		modtime: time.Now(),
		isdir:   isdir,
		containerID: containerID,
	}
}

// Have memFile fulfill os.FileInfo interface
func (f *dockerFile) Name() string { return filepath.Base(f.name) }
func (f *dockerFile) Size() int64  { return int64(len(f.content)) }
func (f *dockerFile) Mode() os.FileMode {
	ret := os.FileMode(0644)
	if f.isdir {
		ret = os.FileMode(0755) | os.ModeDir
	}
	if f.symlink != "" {
		ret = os.FileMode(0777) | os.ModeSymlink
	}
	return ret
}
func (f *dockerFile) ModTime() time.Time { return f.modtime }
func (f *dockerFile) IsDir() bool        { return f.isdir }
func (f *dockerFile) Sys() interface{} {
	return nil
}

// Read/Write
func (f *dockerFile) ReaderAt() (io.ReaderAt, error) {

	if f.isdir {
		return nil, os.ErrInvalid
	}
	return bytes.NewReader(f.content), nil
}

func (f *dockerFile) WriterAt() (io.WriterAt, error) {
	if f.isdir {
		return nil, os.ErrInvalid
	}
	return f, nil
}
func (f *dockerFile) WriteAt(p []byte, off int64) (int, error) {
	uuid, err := uuid.NewRandom()
	tmpFile := os.TempDir() + "/" + uuid.String()
	createdFile, err := os.Create(tmpFile)
	if (err != nil) {
		return 0, err
	}
	createdFile.Write(p)
	f.execFileUpload(createdFile)
	os.Remove(tmpFile)
	return len(p), nil
}

