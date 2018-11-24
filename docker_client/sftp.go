package docker_client

// This serves as an example of how to implement the request server handler as
// well as a dummy backend for testing. It implements an in-memory backend that
// works as a very simple filesystem with simple flat key-value lookup system.

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)


func getRoot(containerID string) *root {
	root := &root{
		files:       make(map[string]*dockerFile),
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

func (fs *root) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	fs.filesLock.Lock()
	defer fs.filesLock.Unlock()
	file, err := fs.fetch(r.Filepath)
	uuid, err := uuid.NewRandom()
	tmpFileFile := os.TempDir() + "/" + uuid.String()
	file.execFileDownload(tmpFileFile)
	bytes, err := ioutil.ReadFile(tmpFileFile)
	if err != nil {
		return nil, err
	}
	file.content = bytes
	os.Remove(tmpFileFile)

	return file.ReaderAt()
}

func (fs *root) Filewrite(r *sftp.Request) (io.WriterAt, error) {
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
	} else {
		file.cleanTmpFile()
		file.content = nil
	}

	uuid, err := uuid.NewRandom()
	tmpFilePath := os.TempDir() + "/" + uuid.String()
	tmpFile, err := os.Create(tmpFilePath) //os.Create(tmpFilePath)
	if err != nil {
		return nil, err
	}
	file.tempFile = tmpFile
	return file.WriterAt()
}
func oct(i int, prefix bool) string {
	i64 := int64(i)

	if prefix {
		return "0o" + strconv.FormatInt(i64, 8) // base 8 for octal
	} else {
		return strconv.FormatInt(i64, 8) // base 8 for octal
	}
}
func (fs *root) Filecmd(r *sftp.Request) error {
	switch r.Method {
	case "Setstat":
		// CHMOD
		attrFlags := r.AttrFlags()
		if attrFlags.Permissions {
			file, err := fs.fetch(r.Filepath)
			if err != nil {
				return err
			}
			fileStat := r.Attributes()
			fileMode := fileStat.FileMode()
			permissions := oct(int(fileMode), false)
			return file.execFileChmod(string(permissions))
		}
		if (attrFlags.Size) {
			file, err := fs.fetch(r.Filepath)
			if (err != nil) {
				return err
			}
			fileStat := r.Attributes()
			size := fileStat.Size
			return file.execTruncate(size)
		}
		return nil
	case "Rename":
		file, err := fs.fetch(r.Filepath)
		if (err != nil) {
			return err
		}
		err = file.execFileRename(r.Target)
		if err != nil {
			return err
		}
	case "Rmdir", "Remove":
		file, err := fs.fetch(r.Filepath)
		if err != nil {
			return err
		}
		err = file.remove()
		if err != nil {
			return err
		}
		delete(fs.files, r.Filepath)
	case "Mkdir":
		folder, err := fs.fetch(filepath.Dir(r.Filepath))
		if err != nil {
			return err
		}
		err = folder.execMkDir(filepath.Base(r.Filepath))
		if err != nil {
			return err
		}
		newDockerFile(r.Filepath, true, folder.containerID)
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

	path := r.Filepath
	if r.Filepath == "/" {
		path = fs.dockerFile.name
	}

	switch r.Method {
	case "List":
		parent, err:= fs.fetch(path)
		if (err != nil) {
			return nil, err
		}
		list, err := parent.execFileList()
		return listerat(list), err
	case "Stat":
		file, err := fs.fetch(path)
		if err != nil {
			return nil, err
		}
		return listerat([]os.FileInfo{file}), nil
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
	files       map[string]*dockerFile
	filesLock   sync.Mutex
	containerID string
}



func (fs *root) fetch(path string) (*dockerFile, error) {
	if path == "/" {
		return fs.dockerFile, nil
	}
	if file, ok := fs.files[path]; ok {
		return file, nil
	}

	file, err := fs.execFileInfo(path)
	if err != nil {
		return nil, err
	}
	fs.files[path] = file
	return file, nil
}

// Implements os.FileInfo, Reader and Writer interfaces.
// These are the 3 interfaces necessary for the Handlers.
type dockerFile struct {
	name        string
	modtime     time.Time
	symlink     string
	size	    string
	isdir       bool
	content     []byte
	contentLock sync.RWMutex
	containerID string
	tempFile    *os.File
}

func createNewDockerFile(lsString string, containerID string) (*dockerFile, error) {
	//lsString = strings.Replace(lsString, "  ", " ", -1)
	parts := strings.Fields(lsString)
	isDirString := parts[0][0:1]
	isDir:=false
	nameIdentifier := 8
	if isDirString == "d" {
		isDir = true
		nameIdentifier = 8
	}
	size := parts[4]
	name := strings.Join(parts[nameIdentifier:len(parts)], " ")
	if strings.Index(name, " ") != -1 {
		name = name + ""
	}
	file := newDockerFile(name, isDir, containerID)
	file.size = size
	return file, nil
}
// factory to make sure modtime is set
func newDockerFile(name string, isdir bool, containerID string) *dockerFile {
	return &dockerFile{
		name:        name,
		modtime:     time.Now(),
		isdir:       isdir,
		containerID: containerID,
	}
}

func (f *dockerFile) cleanTmpFile() error {
	if f.tempFile != nil {
		err := os.Remove(f.tempFile.Name())

		return err
	}
	return nil
}

// Have dockerFile fulfill os.FileInfo interface
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
	n, err := f.tempFile.WriteAt(p, off)
	if err != nil {
		return 0, err
	}
	if len(p) != 0 {
		err = f.execFileUpload(f.tempFile)
	} else {
		err = f.execFileCreate()
	}
	return n, err
}
