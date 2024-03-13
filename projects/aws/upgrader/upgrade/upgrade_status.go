package upgrade

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere-build-tooling/tools/version-tracker/pkg/util/logger"
)

func (u *Upgrader) LogStatusAndCleanup(ctx context.Context) error {
	if err := u.logStatus(ctx); err != nil {
		return fmt.Errorf("logging current status of the node: %v", err)
	}

	upgCmpDir, err := u.upgradeComponentsDir()
	if err != nil {
		return fmt.Errorf("getting upgrade componenets kubernetes binary directory: %v", err)
	}

	cleanUpCmd := []string{"rm", "-rf", upgCmpDir}
	out, err := u.ExecCommand(ctx, cleanUpCmd[0], cleanUpCmd[1:]...)
	if err != nil {
		execError(cleanUpCmd, string(out))
	}

	logger.Info("Cleaning up leftover upgrade components", "in-place components directory", upgCmpDir)
	return nil
}

func (u *Upgrader) logStatus(ctx context.Context) error {
	containerdStatusCmd := []string{"systemctl", "status", "containerd"}
	out, err := u.ExecCommand(ctx, containerdStatusCmd[0], containerdStatusCmd[1:]...)
	if err != nil {
		return execError(containerdStatusCmd, string(out))
	}
	logger.Info("Containerd status on the Node", "status", string(out))

	kubeletStatusCmd := []string{"systemctl", "status", "kubelet"}
	out, err = u.ExecCommand(ctx, kubeletStatusCmd[0], kubeletStatusCmd[1:]...)
	if err != nil {
		return execError(kubeletStatusCmd, string(out))
	}
	logger.Info("Kubelet status on the Node", "status", string(out))

	kubeAdmVersionCmd := []string{"kubeadm", "version"}
	out, err = u.ExecCommand(ctx, kubeAdmVersionCmd[0], kubeAdmVersionCmd[1:]...)
	if err != nil {
		execError(kubeAdmVersionCmd, string(out))
	}
	logger.Info("Kubeadm Version on the Node", "Version", string(out))

	return nil
}
