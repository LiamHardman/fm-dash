//go:build !windows

package main

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
)

func tryLockConfigFile(file *os.File) (bool, error) {
	err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if errors.Is(err, unix.EWOULDBLOCK) {
		return false, nil
	}
	return err == nil, err
}

func lockConfigFile(file *os.File) error { return unix.Flock(int(file.Fd()), unix.LOCK_EX) }
func unlockConfigFile(file *os.File)     { _ = unix.Flock(int(file.Fd()), unix.LOCK_UN) }
