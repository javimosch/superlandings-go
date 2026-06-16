package daemon

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/server"
)

func StartDaemon(cfg *config.Config, port int) error {
	// Check if already running
	if running, _, _ := Status(cfg); running {
		return fmt.Errorf("daemon is already running")
	}

	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}

	// Open log file
	logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening log file: %w", err)
	}
	defer logFile.Close()

	// Start daemon process
	cmd := exec.Command(execPath, "backend", "start", fmt.Sprintf("--port=%d", port))
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting daemon: %w", err)
	}

	// Write PID file
	pid := cmd.Process.Pid
	if err := os.WriteFile(cfg.PIDFile, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("error writing PID file: %w", err)
	}

	fmt.Printf("Daemon started with PID %d\n", pid)
	fmt.Printf("Logs: %s\n", cfg.LogFile)
	fmt.Printf("Access at: http://localhost:%d\n", port)
	return nil
}

func StopDaemon(cfg *config.Config) error {
	// Read PID file
	pidData, err := os.ReadFile(cfg.PIDFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("daemon is not running")
		}
		return fmt.Errorf("error reading PID file: %w", err)
	}

	var pid int
	_, err = fmt.Sscanf(string(pidData), "%d", &pid)
	if err != nil {
		return fmt.Errorf("error parsing PID: %w", err)
	}

	// Find and kill process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("error finding process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("error stopping process: %w", err)
	}

	// Remove PID file
	os.Remove(cfg.PIDFile)
	fmt.Printf("Daemon stopped (PID %d)\n", pid)
	return nil
}

func Status(cfg *config.Config) (bool, int, error) {
	// Check if PID file exists
	if _, err := os.Stat(cfg.PIDFile); os.IsNotExist(err) {
		return false, 0, nil
	}

	// Read PID
	pidData, err := os.ReadFile(cfg.PIDFile)
	if err != nil {
		return false, 0, err
	}

	var pid int
	_, err = fmt.Sscanf(string(pidData), "%d", &pid)
	if err != nil {
		return false, 0, err
	}

	// Check if process is running
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, 0, err
	}

	if err := process.Signal(syscall.Signal(0)); err != nil {
		// Process not running, clean up PID file
		os.Remove(cfg.PIDFile)
		return false, 0, nil
	}

	return true, pid, nil
}

func StartServer(cfg *config.Config, port int) error {
	// Create server
	srv := server.NewServer(cfg)
	
	// Start server
	if err := srv.Start(port); err != nil {
		log.Fatalf("Server error: %v", err)
		return err
	}
	
	return nil
}