package client

// sftp request server connects to docker container.

import (
	"bytes"
	"github.com/apex/log"
	"github.com/mholt/archiver"
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
	if err != nil {
		return nil, err
	}
	err = file.execFileDownload()

	if err != nil {
		return nil, err
	}
	return file.ReaderAt()
}
func (fs *root) createDockerFile(path string, isdir bool, containerID string) *dockerFile {
	fs.files[path] = newDockerFile(path, isdir, fs.containerID)
	return fs.files[path]
}
func (fs *root) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	fs.filesLock.Lock()
	defer fs.filesLock.Unlock()
	file, err := fs.fetch(r.Filepath)

	if err == os.ErrNotExist {
		dir, err := fs.fetch(filepath.Dir(r.Filepath))
		if err != nil {
			parentFolderName := filepath.Base(filepath.Dir(r.Filepath))
			parentFolderPath := filepath.Dir(filepath.Dir(r.Filepath))
			parentDir := fs.createDockerFile(parentFolderPath, true, fs.containerID)
			err := parentDir.execMkDir(parentFolderName)
			if err != nil {
				return nil, os.ErrInvalid
			}
			dir = fs.createDockerFile(filepath.Dir(r.Filepath), true, fs.containerID)
		}
		if !dir.isdir {
			return nil, os.ErrInvalid
		}
		file = fs.createDockerFile(r.Filepath, false, fs.containerID)
	} else {
		file.content = nil
	}
	return file, nil
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
	fs.filesLock.Lock()
	defer fs.filesLock.Unlock()
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
		if attrFlags.Size {
			file, err := fs.fetch(r.Filepath)
			if err != nil {
				return err
			}
			fileStat := r.Attributes()
			size := fileStat.Size
			return file.execTruncate(size)
		}
		return nil
	case "Rename":
		file, err := fs.fetch(r.Filepath)
		if err != nil {
			return err
		}
		err = file.execFileRename(r.Target)
		if err != nil {
			return err
		}
		fs.files[r.Target] = fs.files[r.Filepath]
		delete(fs.files, r.Filepath)
	case "Rmdir", "Remove":
		file, err := fs.fetch(r.Filepath)
		if err != nil {
			return err
		}
		go func() {
			fs.contentLock.Lock()
			file.execRemove()
			defer fs.contentLock.Unlock()
		}()

		file.deleted = true

		// If it is a directory we need to clean up all subfiles/folders.
		if file.IsDir() {
			for path, descendantFile := range fs.files {
				if strings.Contains(path+"/", file.name) {
					descendantFile.deleted = true
				}
			}
		}
	case "Mkdir":
		folder := fs.createDockerFile(filepath.Dir(r.Filepath), true, fs.containerID)
		err := folder.execMkDir(filepath.Base(r.Filepath))
		if err != nil {
			return err
		}
		fs.createDockerFile(r.Filepath, true, folder.containerID)
	case "Symlink":
		// Not implemented
	}
	return nil
}

type listerat []os.FileInfo

// ListAt modeled after strings.Reader's ReadAt() implementation
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
		parent, err := fs.fetch(path)
		if err != nil {
			return nil, err
		}
		list, err := parent.execFileList(fs)

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
	containerID string
	filesLock   sync.Mutex
}

func (fs *root) fetch(path string) (*dockerFile, error) {
	if path == "/" {
		return fs.dockerFile, nil
	}
	if file, ok := fs.files[path]; ok {
		if file.deleted == true {
			return nil, os.ErrNotExist
		}
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
	size        string
	isdir       bool
	content     []byte
	contentLock sync.RWMutex
	containerID string
	deleted     bool
}

func createNewDockerFile(lsString string, containerID string) (*dockerFile, error) {
	//lsString = strings.Replace(lsString, "  ", " ", -1)
	parts := strings.Fields(lsString)
	isDirString := parts[0][0:1]
	isDir := false
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

func (f *dockerFile) WriteAt(p []byte, off int64) (int, error) {
	log.Debugf("Upload file: %s", f.name)
	f.contentLock.Lock()
	defer f.contentLock.Unlock()

	plen := len(p) + int(off)
	if plen >= len(f.content) {
		nc := make([]byte, plen)
		copy(nc, f.content)
		f.content = nc
	}
	copy(f.content[off:], p)

	// Check if bytes where transfered.
	// Otherwise a simple touch is more performant.
	if len(p) != 0 {
		// Here starts the tricky part.
		// The docker api only supports tar upload.
		// Right now there seems no other way as uploading the complete file each time.
		// First create tmp folder.
		tmpDirPath, err := ioutil.TempDir("", "ssh2docksal")
		tmpFilePath := tmpDirPath + "/" + f.Name()
		tmpTarPath := tmpFilePath + ".tar"
		// Create tmp file with the correct file name.
		tmpFile, err := os.Create(tmpFilePath)
		_, err = tmpFile.Write(f.content)
		if err != nil {
			return 0, err
		}
		// than create archive file
		err = archiver.Archive([]string{tmpFilePath}, tmpTarPath)
		if err != nil {
			return 0, err
		}
		// Last .. upload the file.
		tarFile, err := os.Open(tmpTarPath)
		if err != nil {
			return 0, err
		}
		err = f.execFileUpload(tarFile)

		// Cleanup the tmp folder.
		os.Remove(tmpDirPath)

	} else {
		err := f.execFileCreate()
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}
