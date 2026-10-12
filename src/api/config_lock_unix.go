//go:build !windows

package main

import (
	"golang.org/x/sys/unix"
	"os"
)

func lockConfigFile(file *os.File) error { return unix.Flock(int(file.Fd()), unix.LOCK_EX) }
func unlockConfigFile(file *os.File)     { _ = unix.Flock(int(file.Fd()), unix.LOCK_UN) }
