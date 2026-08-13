//go:build windows

package platformservice

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"golang.org/x/sys/windows/svc"
)

// Run integrates the component with the Windows Service Control Manager when
// launched as a service, while preserving normal Ctrl+C behavior in a console.
func Run(name string, component func(context.Context) error) error {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return fmt.Errorf("detect Windows service: %w", err)
	}
	if !isService {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		return component(ctx)
	}
	return svc.Run(name, &handler{component: component})
}

type handler struct {
	component func(context.Context) error
}

func (h *handler) Execute(_ []string, requests <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- h.component(ctx) }()
	status := svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	changes <- status

	for {
		select {
		case err := <-done:
			if err != nil {
				log.Printf("Windows service stopped with error: %v", err)
				return true, 1
			}
			return false, 0
		case request := <-requests:
			switch request.Cmd {
			case svc.Interrogate:
				changes <- status
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				cancel()
				if err := <-done; err != nil {
					log.Printf("Windows service shutdown error: %v", err)
					return true, 1
				}
				return false, 0
			}
		}
	}
}
