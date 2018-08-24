// +build windows

package oshelper

import (
	"os"

	"code.cloudfoundry.org/goshims/osshim"
)

type osHelper struct {
	os osshim.Os
}

func NewOsHelper(os osshim.Os) *osHelper {
	return &osHelper{
		os: os,
	}
}

func (o *osHelper) Umask(mask int) (oldmask int) {
	return 0
}

func (o *osHelper) Mount(srcPath string, targetPath string) error {
	return o.os.Symlink(srcPath, targetPath)
}

func (o *osHelper) Unmount(targetPath string) error {
	return o.os.Remove(targetPath)
}

func (o *osHelper) IsMounted(targetPath string) (bool, error) {
	_, err := o.os.Stat(targetPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
