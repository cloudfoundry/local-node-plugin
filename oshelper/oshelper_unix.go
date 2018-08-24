// +build linux darwin

package oshelper

import (
	"os/exec"
	"syscall"

	"code.cloudfoundry.org/goshims/osshim"
)

type osHelper struct {
}

func NewOsHelper(os osshim.Os) *osHelper {
	return &osHelper{}
}

func (o *osHelper) Umask(mask int) (oldmask int) {
	return syscall.Umask(mask)
}

func (o *osHelper) Mount(srcPath string, targetPath string) error {
	cmd := exec.Command("mount", "--bind", srcPath, targetPath)
	return cmd.Run()
}

func (o *osHelper) Unmount(targetPath string) error {
	cmd := exec.Command("umount", targetPath)
	return cmd.Run()
}

func (o *osHelper) IsMounted(targetPath string) (bool, error) {
	cmd := exec.Command("mountpoint", "-q", targetPath)
	err := cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
