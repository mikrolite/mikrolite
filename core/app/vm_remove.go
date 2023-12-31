package app

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
)

func (a *app) RemoveVM(ctx context.Context, name string, owner string) error {
	pterm.DefaultSpinner.Info(fmt.Sprintf("ℹ️  Removing VM: %s\n", name))

	if err := a.vmService.Stop(ctx, name); err != nil {
		return fmt.Errorf("stopping vm: %w", err)
	}

	if err := a.vmService.Delete(ctx, name); err != nil {
		return fmt.Errorf("deleting vm: %w", err)
	}

	if err := a.imageService.Cleanup(ctx, owner); err != nil {
		return fmt.Errorf("cleaning up vm images: %w", err)
	}

	vm, err := a.stateService.GetVM()
	if err != nil {
		return fmt.Errorf("getting vm config: %w", err)
	}

	for _, netStatus := range vm.Status.NetworkStatus {
		if err := a.networkService.InterfaceDelete(netStatus.HostDeviveName); err != nil {
			return fmt.Errorf("deleting vm network interface: %w", err)
		}
	}

	if err := a.fs.RemoveAll(a.stateService.Root()); err != nil {
		return fmt.Errorf("removing vm state directory %s: %w", a.stateService.Root(), err)
	}

	pterm.DefaultSpinner.Info(fmt.Sprintf("ℹ️  Removed VM: %s\n", name))
	return nil
}
