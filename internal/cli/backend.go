package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/javimosch/superlandings-go/internal/daemon"
	"github.com/spf13/cobra"
)

var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Manage backend daemon",
	Long:  `Start, stop, and check status of the backend daemon for the web UI.`,
}

var backendStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start backend daemon",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		daemonMode, _ := cmd.Flags().GetBool("daemon")
		noSystemd, _ := cmd.Flags().GetBool("no-systemd")

		if daemonMode {
			// Try systemd first if available and not disabled
			if !noSystemd && isSystemdAvailable() {
				fmt.Println("Systemd detected, installing service...")
				if err := installAndStartSystemdService(port); err != nil {
					fmt.Fprintf(os.Stderr, "Systemd installation failed: %v\n", err)
					fmt.Println("Falling back to basic daemon mode...")
					if err := daemon.StartDaemon(cfg, port); err != nil {
						fmt.Fprintf(os.Stderr, "Error starting daemon: %v\n", err)
						os.Exit(1)
					}
				} else {
					fmt.Println("Systemd service installed and started successfully!")
					fmt.Println("The service will auto-start on boot.")
					fmt.Printf("Access at: http://localhost:%d\n", port)
					fmt.Println("Manage with: sudo systemctl {start|stop|restart|status} sl-cli")
					return
				}
			} else {
				// Use basic daemon mode
				if err := daemon.StartDaemon(cfg, port); err != nil {
					fmt.Fprintf(os.Stderr, "Error starting daemon: %v\n", err)
					os.Exit(1)
				}
			}
		} else {
			if err := daemon.StartServer(cfg, port); err != nil {
				fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

var backendStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop backend daemon",
	Run: func(cmd *cobra.Command, args []string) {
		uninstall, _ := cmd.Flags().GetBool("uninstall")

		// Check if systemd service is running
		if isSystemdServiceInstalled() {
			if uninstall {
				fmt.Println("Stopping and uninstalling systemd service...")
				if err := stopAndRemoveSystemdService(); err != nil {
					fmt.Fprintf(os.Stderr, "Error removing systemd service: %v\n", err)
					fmt.Println("Falling back to basic daemon stop...")
					if err := daemon.StopDaemon(cfg); err != nil {
						fmt.Fprintf(os.Stderr, "Error stopping daemon: %v\n", err)
						os.Exit(1)
					}
				} else {
					fmt.Println("Systemd service stopped and uninstalled successfully!")
					return
				}
			} else {
				fmt.Println("Stopping systemd service...")
				if err := stopSystemdService(); err != nil {
					fmt.Fprintf(os.Stderr, "Error stopping systemd service: %v\n", err)
					fmt.Println("Falling back to basic daemon stop...")
					if err := daemon.StopDaemon(cfg); err != nil {
						fmt.Fprintf(os.Stderr, "Error stopping daemon: %v\n", err)
						os.Exit(1)
					}
				} else {
					fmt.Println("Systemd service stopped successfully!")
					fmt.Println("Service remains installed and will auto-start on boot.")
					fmt.Println("To completely remove, use: sl-cli backend stop --uninstall")
					return
				}
			}
		} else {
			// Use basic daemon mode
			if err := daemon.StopDaemon(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Error stopping daemon: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

var backendStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check backend daemon status",
	Run: func(cmd *cobra.Command, args []string) {
		running, pid, err := daemon.Status(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking status: %v\n", err)
			os.Exit(1)
		}

		if running {
			fmt.Printf("Daemon running (PID: %d)\n", pid)
			fmt.Printf("Logs: %s\n", cfg.LogFile)
		} else {
			fmt.Println("Daemon not running")
		}
	},
}

func init() {
	backendStartCmd.Flags().Int("port", 8080, "Port for HTTP server")
	backendStartCmd.Flags().Bool("daemon", false, "Run as daemon in background")
	backendStartCmd.Flags().Bool("no-systemd", false, "Disable systemd auto-installation")
	backendStopCmd.Flags().Bool("uninstall", false, "Stop and uninstall systemd service")
	
	backendCmd.AddCommand(backendStartCmd)
	backendCmd.AddCommand(backendStopCmd)
	backendCmd.AddCommand(backendStatusCmd)
}

// systemd helper functions

func isSystemdAvailable() bool {
	// Check if systemctl exists and we're running under systemd
	if _, err := os.Stat("/usr/bin/systemctl"); os.IsNotExist(err) {
		return false
	}
	// Check if we're in a systemd environment
	if _, err := os.Stat("/run/systemd/system"); os.IsNotExist(err) {
		return false
	}
	return true
}

func isSystemdServiceInstalled() bool {
	if _, err := os.Stat("/etc/systemd/system/sl-cli.service"); os.IsNotExist(err) {
		return false
	}
	return true
}

func installAndStartSystemdService(port int) error {
	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Get current user
	user := os.Getenv("USER")
	if user == "" {
		user = "root"
	}

	// Create service file content
	serviceContent := fmt.Sprintf(`[Unit]
Description=SuperLandings CLI Daemon
After=network.target

[Service]
Type=simple
User=%s
WorkingDirectory=%s
ExecStart=%s backend start --port=%d
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`, user, workingDir, execPath, port)

	// Write service file
	servicePath := "/etc/systemd/system/sl-cli.service"
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	// Enable service
	if err := exec.Command("systemctl", "enable", "sl-cli").Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	// Start service
	if err := exec.Command("systemctl", "start", "sl-cli").Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

func stopSystemdService() error {
	if err := exec.Command("systemctl", "stop", "sl-cli").Run(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}
	return nil
}

func stopAndRemoveSystemdService() error {
	// Stop service
	if err := exec.Command("systemctl", "stop", "sl-cli").Run(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	// Disable service
	if err := exec.Command("systemctl", "disable", "sl-cli").Run(); err != nil {
		return fmt.Errorf("failed to disable service: %w", err)
	}

	// Remove service file
	servicePath := "/etc/systemd/system/sl-cli.service"
	if err := os.Remove(servicePath); err != nil {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	return nil
}