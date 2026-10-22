//go:build windows

package filecoord

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// OpenPhysicalPath opens a file or directory without traversing a reparse
// point. The object is opened with OBJ_DONT_REPARSE, so a link or junction
// at the final component is refused rather than resolved.
func OpenPhysicalPath(path string, directory bool) (*os.File, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	handle, err := openWindowsStoreLockObject(ntPath(absPath), windows.NtCreateFile, queryWindowsDosDevice)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: absPath, Err: err}
	}
	return os.NewFile(uintptr(handle), absPath), nil
}
