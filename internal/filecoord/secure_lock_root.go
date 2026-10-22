package filecoord

// secureLockRoot returns the directory the no-follow open walk starts from.
// The walk opens every component below it with O_NOFOLLOW, so a symlinked
// component anywhere in the path is refused; starting at the filesystem root
// therefore needs no authority directory of its own.
func secureLockRoot(path string) (string, bool) {
	return "/", true
}
