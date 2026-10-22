//go:build unix

package filecoord

import (
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

// OpenPhysicalPath opens a file or directory without traversing a symlink at
// the final component. Every parent component is opened with O_NOFOLLOW too,
// so a symlinked directory in the path is refused rather than resolved.
func OpenPhysicalPath(path string, directory bool) (*os.File, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	parentFD, err := secureOpenLockParent(string(filepath.Separator), filepath.Dir(absPath))
	if err != nil {
		return nil, err
	}
	defer unix.Close(parentFD)

	flags := syscall.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW
	if directory {
		flags |= unix.O_DIRECTORY
	}
	fd, err := unix.Openat(parentFD, filepath.Base(absPath), flags, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: absPath, Err: err}
	}
	return os.NewFile(uintptr(fd), absPath), nil
}
