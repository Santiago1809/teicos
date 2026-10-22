package filecoord

import (
	"errors"
	"os"
)

// acquireCooperativeLock backs the cooperative contract with the hardened
// authority-lock primitive instead of a second platform implementation:
// AcquireAuthorityFileLock owns the secure no-follow open walk (O_NOFOLLOW
// per component on POSIX, reparse-safe NtCreateFile on Windows) and kernel
// advisory locking (flock / LockFileEx), and never treats lock-file residue
// as ownership. Symlinked roots and lock paths are rejected by that walk, so
// callers must pass canonical lock roots.
func acquireCooperativeLock(path string) (*Lease, error) {
	authority, err := AcquireAuthorityFileLock(path)
	if err != nil {
		if errors.Is(err, ErrLockContended) {
			return nil, &BusyError{Cause: err}
		}
		return nil, &OperationalError{Cause: err}
	}
	return &Lease{unlock: func(*os.File) error { return authority.Release() }}, nil
}
