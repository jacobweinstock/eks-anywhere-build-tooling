package upgrade

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/eks-anywhere-build-tooling/tools/version-tracker/pkg/util/logger"
)

const (
	kubectlBinPath                 = "/usr/bin/kubectl"
	kubeletBinPath                 = "/usr/bin/kubelet"
	kubeletConf                    = "/etc/sysconfig/kubelet"
	kubeletCredProviderFeatureGate = " --feature-gates=KubeletCredentialProviders=true"
	extraArgs                      = "extra_args"
)

func (u *Upgrader) KubeletKubectlUpgrade(ctx context.Context) error {
	cmpDir, err := u.upgradeComponentsKubernetesBinDir()
	if err != nil {
		return fmt.Errorf("getting upgrade componenets binary directory: %v", err)
	}

	if err := u.BackUpAndReplace(kubectlBinPath, cmpDir, fmt.Sprintf("%s/kubectl", cmpDir)); err != nil {
		return fmt.Errorf("backing up and replacing kubectl binary: %v", err)
	}
	logger.Info("Backed up and replaced kubectl sucessfully")

	stopKubeletCmd := []string{"systemctl", "stop", "kubelet"}
	out, err := u.ExecCommand(ctx, stopKubeletCmd[0], stopKubeletCmd[1:]...)
	if err != nil {
		return execError(stopKubeletCmd, string(out))
	}

	if err := u.BackUpAndReplace(kubeletBinPath, cmpDir, fmt.Sprintf("%s/kubelet", cmpDir)); err != nil {
		return fmt.Errorf("backing up and replacing kubelet binary: %v", err)
	}
	logger.Info("Backed up and replaced kubelet sucessfully")

	kubeAdmVersionCmd := []string{"kubeadm", "version", "-oshort"}
	kubeVersion, err := u.ExecCommand(ctx, kubeAdmVersionCmd[0], kubeAdmVersionCmd[1:]...)
	kubeVersionStr := string(kubeVersion)
	if err != nil {
		execError(kubeAdmVersionCmd, kubeVersionStr)
	}

	// KubeletCredentialProviders support became GA in k8s v1.26, and the feature gate was removed in k8s v1.28.
	// For in-place upgrades, we should remove this feature gate if it exists on nodes running k8s v1.26 and above.
	if strings.Contains(kubeVersionStr, "v1.25") {
		if err := u.updateKubeletExtraArgs(cmpDir); err != nil {
			return fmt.Errorf("updating kubelet extra args: %v", err)
		}
	}

	daemonReloadCmd := []string{"systemctl", "daemon-reload"}
	out, err = u.ExecCommand(ctx, daemonReloadCmd[0], daemonReloadCmd[1:]...)
	if err != nil {
		execError(daemonReloadCmd, string(out))
	}

	restartkubeletCmd := []string{"systemctl", "restart", "kubelet"}
	out, err = u.ExecCommand(ctx, restartkubeletCmd[0], restartkubeletCmd[1:]...)
	if err != nil {
		execError(restartkubeletCmd, string(out))
	}

	logger.Info("Kubectl and Kubelet upgrade successful!")
	return nil
}

func (u *Upgrader) updateKubeletExtraArgs(cmpDir string) error {
	if _, err := u.Stat(kubeletConf); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Info("kubelet config file not found, skipping updating extra args", "kubelet_config_path", kubeletConf)
			return nil
		}
	}

	conf, err := u.ReadFile(kubeletConf)
	if err != nil {
		return fmt.Errorf("reading kubelet config on the node: %v", err)
	}
	newConf := bytes.ReplaceAll(conf, []byte(kubeletCredProviderFeatureGate), []byte(""))

	extraArgsDir := fmt.Sprintf("%s/%s", cmpDir, extraArgs)
	if err = u.MkdirAll(extraArgsDir, 0o755); err != nil {
		return fmt.Errorf("creating folder: %v", err)
	}
	kubeletConfBackupFile := fmt.Sprintf("%s/kubelet.bk", extraArgsDir)
	if err = u.copy(kubeletConf, kubeletConfBackupFile); err != nil {
		return copyError(kubeletConf, kubeletConfBackupFile, err)
	}
	if err := u.WriteFile(kubeletConf, newConf, 0o644); err != nil {
		return fmt.Errorf("writing updated kubelet config to file: %v", err)
	}

	logger.Info("Removed deprecated feature flags from kubelet config successfully!")
	return nil
}
