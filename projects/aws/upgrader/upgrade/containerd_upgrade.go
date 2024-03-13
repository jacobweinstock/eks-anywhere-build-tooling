package upgrade

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere-build-tooling/tools/version-tracker/pkg/util/logger"
)

func (u *Upgrader) ContainerdUpgrade(ctx context.Context) error {
	cmpDir, err := u.upgradeComponentsBinDir()
	if err != nil {
		return fmt.Errorf("getting upgrade componenets binary directory: %v", err)
	}

	containerdVersionCmd := []string{"containerd", "--version"}
	out, err := u.ExecCommand(ctx, containerdVersionCmd[0], containerdVersionCmd[1:]...)
	if err != nil {
		return execError(containerdVersionCmd, string(out))
	}

	cpCmd := []string{"cp", "-rf", fmt.Sprintf("%s/containerd/.", cmpDir), "/"}
	out, err = u.ExecCommand(ctx, cpCmd[0], cpCmd[1:]...)
	if err != nil {
		return execError(cpCmd, string(out))
	}

	version, err := u.ExecCommand(ctx, containerdVersionCmd[0], containerdVersionCmd[1:]...)
	if err != nil {
		return execError(containerdVersionCmd, string(version))
	}

	daemonReloadCmd := []string{"systemctl", "daemon-reload"}
	out, err = u.ExecCommand(ctx, daemonReloadCmd[0], daemonReloadCmd[1:]...)
	if err != nil {
		return execError(daemonReloadCmd, string(out))
	}

	containerdRestartCmd := []string{"systemctl", "restart", "containerd"}
	out, err = u.ExecCommand(ctx, containerdRestartCmd[0], containerdRestartCmd[1:]...)
	if err != nil {
		return execError(containerdRestartCmd, string(out))
	}

	logger.Info("Containerd Version on the Node", "Version", string(version))
	logger.Info("Containerd upgrade successful!")
	return nil
}
