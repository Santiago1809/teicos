//go:build linux

package filecoord

import "golang.org/x/sys/unix"

func secureDirectoryOpenFlags() int {
	return unix.O_PATH | unix.O_DIRECTORY | unix.O_CLOEXEC
}
