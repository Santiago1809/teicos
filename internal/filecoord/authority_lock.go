package filecoord

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AuthorityFileLock exposes the hardened, platform-specific no-follow lock
// primitive to the cooperative locks in this package. It preserves kernel
// advisory ownership as truth and never treats PID metadata as authorization.
type AuthorityFileLock struct{ lock *authorityLock }

// authorityLock is one locked file plus the owner record written into it.
type authorityLock struct {
	file  *os.File
	owner lockOwner
}

// lockOwner is the record written into the lock file. It names the holder so
// a reader can report it, but it never grants authorization on its own.
type lockOwner struct {
	Schema     string    `json:"schema"`
	OwnerID    string    `json:"owner_id"`
	PID        int       `json:"pid"`
	AcquiredAt time.Time `json:"acquired_at"`
}

// lockSchema identifies the lock file format.
const lockSchema = "teicos.filecoord-lock/v1"

// lockCoordinationBytes is the byte range the Windows lock primitive
// coordinates. POSIX flock remains whole-file advisory.
const lockCoordinationBytes int64 = 1

// ErrLockContended reports the one refusal that proves, from where it
// happened, that another holder owns the lock.
var ErrLockContended = errors.New("advisory file lock is held")

// lockBusyError is the internal form of ErrLockContended.
type lockBusyError struct{}

func (err lockBusyError) Error() string { return ErrLockContended.Error() }

func (err lockBusyError) Unwrap() []error { return []error{ErrLockContended} }

// ErrLockPreAcquisitionError reports a failure that happened before the lock
// was taken, so the caller knows no ownership was established.
type ErrLockPreAcquisitionError struct {
	Err error
}

func (err *ErrLockPreAcquisitionError) Error() string {
	return fmt.Sprintf("acquire advisory file lock: %v", err.Err)
}

func (err *ErrLockPreAcquisitionError) Unwrap() error { return err.Err }

// AcquireAuthorityFileLock opens and exclusively locks one private lock path.
// The open never follows a symlink, and the lock is the kernel's: a stale
// lock file is residue, not ownership.
func AcquireAuthorityFileLock(path string) (*AuthorityFileLock, error) {
	lock, err := acquireLocalLock(path)
	if err != nil {
		return nil, err
	}
	return &AuthorityFileLock{lock: lock}, nil
}

// Release relinquishes an AuthorityFileLock.
func (lock *AuthorityFileLock) Release() error {
	if lock == nil || lock.lock == nil {
		return nil
	}
	err := lock.lock.release()
	lock.lock = nil
	return err
}

func acquireLocalLock(path string) (*authorityLock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, &ErrLockPreAcquisitionError{Err: err}
	}
	file, err := secureOpenLocalStoreLock(path)
	if err != nil {
		return nil, &ErrLockPreAcquisitionError{Err: err}
	}
	locked, err := tryLockFile(file)
	if err != nil {
		_ = file.Close()
		return nil, &ErrLockPreAcquisitionError{Err: fmt.Errorf("acquire bounded file lock: %w", err)}
	}
	if !locked {
		_ = file.Close()
		return nil, lockBusyError{}
	}

	owner, err := newLockOwner()
	if err != nil {
		_ = unlockFile(file)
		_ = file.Close()
		return nil, err
	}
	payload, err := json.Marshal(owner)
	if err != nil {
		_ = unlockFile(file)
		_ = file.Close()
		return nil, err
	}
	payload = append(payload, '\n')
	if err := file.Truncate(0); err == nil {
		_, err = file.Seek(0, 0)
	}
	if err == nil {
		_, err = file.Write(payload)
	}
	if err == nil {
		err = file.Sync()
	}
	if err != nil {
		_ = unlockFile(file)
		_ = file.Close()
		return nil, fmt.Errorf("write file lock owner: %w", err)
	}
	return &authorityLock{file: file, owner: owner}, nil
}

func (lock *authorityLock) release() error {
	if lock == nil || lock.file == nil {
		return nil
	}
	// Clear the owner record while the advisory lock is still held, so no
	// observer ever reads an exited process as the recorded holder. Kernel
	// ownership stays the truth; the record is only ever a hint.
	truncateErr := lock.file.Truncate(0)
	unlockErr := unlockFile(lock.file)
	closeErr := lock.file.Close()
	lock.file = nil
	return errors.Join(truncateErr, unlockErr, closeErr)
}

func newLockOwner() (lockOwner, error) {
	pid := os.Getpid()
	return lockOwner{
		Schema:     lockSchema,
		OwnerID:    fmt.Sprintf("%d-%d", pid, time.Now().UnixNano()),
		PID:        pid,
		AcquiredAt: time.Now().UTC(),
	}, nil
}
