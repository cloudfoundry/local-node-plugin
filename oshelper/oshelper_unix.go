// +build linux darwin

package oshelper

import "syscall"

type osHelper struct {
}

type OsHelper interface {
	Umask(mask int) (oldmask int)
}

func NewOsHelper() OsHelper {
	return &osHelper{}
}

func (o *osHelper) Umask(mask int) (oldmask int) {
	return syscall.Umask(mask)
}
