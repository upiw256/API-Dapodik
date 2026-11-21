//go:build windows

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows/svc"
)

const windowsServiceName = "DapodikAPI"

func runServiceOrConsole() error {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return fmt.Errorf("detect Windows service mode: %w", err)
	}
	if !isService {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		return runAPI(ctx)
	}

	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate service executable: %w", err)
	}
	logFile, err := os.OpenFile(filepath.Join(filepath.Dir(executablePath), "dapodik-api.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open service log: %w", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	return svc.Run(windowsServiceName, apiWindowsService{})
}

type apiWindowsService struct{}

func (apiWindowsService) Execute(_ []string, requests <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- runAPI(ctx) }()
	status := svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	changes <- status

	for {
		select {
		case err := <-done:
			if err != nil {
				log.Printf("API service stopped: %v", err)
				return false, 1
			}
			return false, 0
		case change := <-requests:
			switch change.Cmd {
			case svc.Interrogate:
				changes <- status
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				cancel()
				if err := <-done; err != nil {
					log.Printf("API service stopped with error: %v", err)
					return false, 1
				}
				return false, 0
			}
		}
	}
}
