package upgrade

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere-build-tooling/tools/version-tracker/pkg/util/logger"
)

func (u *Upgrader) CniPluginsUpgrade(ctx context.Context) error {
	cmpDir, err := u.upgradeComponentsBinDir()
	if err != nil {
		return fmt.Errorf("getting upgrade componenets binary directory: %v", err)
	}

	cniVersionCmd := []string{"/opt/cni/bin/loopback", "--version"}
	out, err := u.ExecCommand(ctx, cniVersionCmd[0], cniVersionCmd[1:]...)
	if err != nil {
		return execError(cniVersionCmd, string(out))
	}

	cpCmd := []string{"cp", "-rf", fmt.Sprintf("%s/cni-plugins/.", cmpDir), "/"}
	out, err = u.ExecCommand(ctx, cpCmd[0], cpCmd[1:]...)
	if err != nil {
		return execError(cpCmd, string(out))
	}

	out, err = u.ExecCommand(ctx, cniVersionCmd[0], cniVersionCmd[1:]...)
	if err != nil {
		return execError(cniVersionCmd, string(out))
	}

	logger.Info("Cni-Plugins Version on the Node", "Version", string(out))
	logger.Info("Cni-Plugins upgrade succesful!")
	return nil
}
