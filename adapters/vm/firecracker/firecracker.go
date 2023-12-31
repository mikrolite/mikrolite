package firecracker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"

	"github.com/spf13/afero"

	"github.com/mikrolite/mikrolite/adapters/vm/shared"
	"github.com/mikrolite/mikrolite/core/ports"
	"github.com/mikrolite/mikrolite/defaults"
)

const (
	ProviderName = "firecracker"
)

func New(binaryPath string, stateService ports.StateService, ds ports.DiskService, fs afero.Fs) ports.VMProvider {
	return &Provider{
		ss:         stateService,
		fs:         fs,
		ds:         ds,
		binaryPath: binaryPath,
	}
}

type Provider struct {
	ss         ports.StateService
	ds         ports.DiskService
	fs         afero.Fs
	binaryPath string
}

// Stop will stop a running vm.
func (f *Provider) Stop(ctx context.Context, name string) error {
	slog.Debug("stopping firecracker vm", "name", name)

	pid, err := f.ss.GetPID()
	if err != nil {
		return fmt.Errorf("getting vm pid: %w", err)
	}

	if pid == 0 {
		slog.Debug("pid not set for vm, skipping stop", "name", name)

		return nil
	}

	p, _ := os.FindProcess(pid)
	p.Signal(syscall.SIGHUP)

	//TODO: wait for the process to exit

	return nil
}

// Delete will delete a running vm.
func (f *Provider) Delete(ctx context.Context, name string) error {
	pid, err := f.ss.GetPID()
	if err != nil {
		return fmt.Errorf("getting vm pid: %w", err)
	}

	if pid == 0 {
		slog.Debug("pid not set for vm, skipping stop", "name", name)

		return nil
	}

	if err := shared.StopProcess(ctx, pid); err != nil {
		return fmt.Errorf("stopping firecracker process: %w", err)
	}

	return nil
}

func (f *Provider) HasMetadataService() bool {
	return true
}

func (f *Provider) ensureLogPath() error {
	logFile, err := f.fs.OpenFile(f.ss.LogPath(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, defaults.DataFilePerm)
	if err != nil {
		return fmt.Errorf("creating log file %s: %w", f.ss.LogPath(), err)
	}
	logFile.Close()

	return nil
}

func (f *Provider) writeNetworkConfig(path, networkName string) error {
	return os.WriteFile(path, []byte(fmt.Sprintf(`{
		"cniVersion": "0.3.1",
		"name": "%s",
		"plugins": [
		  {
			"type": "ptp",
			"ipam": {
			  "type": "host-local",
			  "subnet": "192.168.127.0/24"
			}
		  },
		  {
			"type": "tc-redirect-tap"
		  }
		]
	  }`, networkName)), 0644)
}

func (f *Provider) socketPath() string {
	return filepath.Join(f.ss.Root(), "firecracker.sock")
}

func defaultKernelCmdLine() map[string]string {
	return map[string]string{
		"console":       "ttyS0",
		"reboot":        "k",
		"panic":         "1",
		"pci":           "off",
		"i8042.noaux":   "",
		"i8042.nomux":   "",
		"i8042.nopnp":   "",
		"i8042.dumbkbd": "",
	}
}
