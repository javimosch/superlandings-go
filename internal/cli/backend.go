package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"

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

		cfg.AuthToken, _ = cmd.Flags().GetString("auth-token")
		cfg.SyncTargetHost, _ = cmd.Flags().GetString("sync-host")
		cfg.SyncTargetUser, _ = cmd.Flags().GetString("sync-user")
		cfg.SyncTargetPort, _ = cmd.Flags().GetInt("sync-port")
		cfg.SyncTargetKey, _ = cmd.Flags().GetString("sync-key")

		if daemonMode {
			if !noSystemd && isSystemdAvailable() {
				authToken, _ := cmd.Flags().GetString("auth-token")
				syncHost, _ := cmd.Flags().GetString("sync-host")
				syncUser, _ := cmd.Flags().GetString("sync-user")
				syncPort, _ := cmd.Flags().GetInt("sync-port")
				syncKey, _ := cmd.Flags().GetString("sync-key")

				if err := installAndStartSystemdService(port, authToken, syncHost, syncUser, syncPort, syncKey); err != nil {
					fmt.Fprintf(os.Stderr, "{\"warning\":\"systemd installation failed: %v, falling back to basic daemon\",\"version\":\"1.0\"}\n", err)
					if err := daemon.StartDaemon(cfg, port); err != nil {
						fail(ExitExtFailed, err.Error())
					}
				} else {
					success("Daemon started via systemd", map[string]interface{}{
						"port": port, "pid_file": cfg.PIDFile, "systemd": true, "auto_start": true,
					})
					return
				}
			} else {
				if err := daemon.StartDaemon(cfg, port); err != nil {
					fail(ExitExtFailed, err.Error())
				}
			}
		} else {
			if err := daemon.StartServer(cfg, port); err != nil {
				fail(ExitExtFailed, err.Error())
			}
		}

		success("Server started", map[string]interface{}{
			"port": port, "daemon": daemonMode, "pid_file": cfg.PIDFile,
		})
	},
}

var backendStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop backend daemon",
	Run: func(cmd *cobra.Command, args []string) {
		uninstall, _ := cmd.Flags().GetBool("uninstall")

		if isSystemdServiceInstalled() {
			if uninstall {
				if err := stopAndRemoveSystemdService(); err != nil {
					failf(ExitExtFailed, "stopping systemd: %v", err)
				}
			} else {
				if err := stopSystemdService(); err != nil {
					failf(ExitExtFailed, "stopping systemd: %v", err)
				}
			}
			success("Daemon stopped", map[string]interface{}{"systemd": true, "uninstalled": uninstall})
			return
		}

		if err := daemon.StopDaemon(cfg); err != nil {
			fail(ExitExtFailed, err.Error())
		}
		success("Daemon stopped", nil)
	},
}

var backendStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check backend daemon status",
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteBackendStatus(target)
			return
		}

		running, pid, err := daemon.Status(cfg)
		if err != nil {
			fail(ExitInternal, err.Error())
		}

		writeJSON(map[string]interface{}{
			"version": "1.0", "running": running, "pid": pid,
			"log_file": cfg.LogFile, "pid_file": cfg.PIDFile,
		})
	},
}


var backendInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install systemd service for boot persistence",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		if err := installSystemdService(port); err != nil {
			fail(ExitExtFailed, err.Error())
		}
		success("Systemd service installed", map[string]interface{}{
			"port": port, "service": "sl-cli",
		})
	},
}

var backendUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove systemd service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := uninstallSystemdService(); err != nil {
			fail(ExitExtFailed, err.Error())
		}
		success("Systemd service uninstalled", nil)
	},
}


func init() {
	backendStartCmd.Flags().Int("port", 8080, "Port for HTTP server")
	backendStartCmd.Flags().Bool("daemon", false, "Run as daemon in background")
	backendStartCmd.Flags().Bool("no-systemd", false, "Disable systemd auto-installation")
	backendStartCmd.Flags().String("auth-token", "", "API authentication token")
	backendStartCmd.Flags().String("sync-host", "", "Sync target host")
	backendStartCmd.Flags().String("sync-user", "root", "Sync target SSH user")
	backendStartCmd.Flags().Int("sync-port", 22, "Sync target SSH port")
	backendStartCmd.Flags().String("sync-key", "", "Sync target SSH key path")
	backendStopCmd.Flags().Bool("uninstall", false, "Stop and uninstall systemd service")
	backendStatusCmd.Flags().String("target", "", "Remote target (host:port)")

	backendCmd.AddCommand(backendStartCmd)
	backendCmd.AddCommand(backendStopCmd)
	backendCmd.AddCommand(backendStatusCmd)
	backendInstallCmd.Flags().Int("port", 3099, "Port for HTTP server")
	backendCmd.AddCommand(backendInstallCmd)
	backendCmd.AddCommand(backendUninstallCmd)

}

func isSystemdAvailable() bool {
	if _, err := os.Stat("/usr/bin/systemctl"); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat("/run/systemd/system"); os.IsNotExist(err) {
		return false
	}
	return true
}

func isSystemdServiceInstalled() bool {
	_, err := os.Stat("/etc/systemd/system/sl-cli.service")
	return err == nil
}

func installAndStartSystemdService(port int, authToken, syncHost, syncUser string, syncPort int, syncKey string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working dir: %w", err)
	}
	user := os.Getenv("USER")
	if user == "" {
		user = "root"
	}

	cmd := fmt.Sprintf("%s backend start --port=%d", execPath, port)
	if authToken != "" {
		cmd += fmt.Sprintf(" --auth-token=%s", authToken)
	}
	if syncHost != "" {
		cmd += fmt.Sprintf(" --sync-host=%s", syncHost)
	}
	if syncUser != "" && syncUser != "root" {
		cmd += fmt.Sprintf(" --sync-user=%s", syncUser)
	}
	if syncPort != 22 {
		cmd += fmt.Sprintf(" --sync-port=%d", syncPort)
	}
	if syncKey != "" {
		cmd += fmt.Sprintf(" --sync-key=%s", syncKey)
	}

	svc := fmt.Sprintf(`[Unit]
Description=SuperLandings CLI Daemon
After=network.target
[Service]
Type=simple
User=%s
WorkingDirectory=%s
ExecStart=%s
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
[Install]
WantedBy=multi-user.target
`, user, workingDir, cmd)

	servicePath := "/etc/systemd/system/sl-cli.service"
	if err := os.WriteFile(servicePath, []byte(svc), 0644); err != nil {
		return fmt.Errorf("writing service file: %w", err)
	}
	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "sl-cli").Run()
	return exec.Command("systemctl", "start", "sl-cli").Run()
}

func stopSystemdService() error {
	return exec.Command("systemctl", "stop", "sl-cli").Run()
}

func stopAndRemoveSystemdService() error {
	exec.Command("systemctl", "stop", "sl-cli").Run()
	exec.Command("systemctl", "disable", "sl-cli").Run()
	os.Remove("/etc/systemd/system/sl-cli.service")
	return exec.Command("systemctl", "daemon-reload").Run()
}
const serviceTmpl = `[Unit]
Description=SuperLandings CLI Daemon
After=network.target
[Service]
Type=simple
User={{.User}}
WorkingDirectory={{.WorkingDir}}
ExecStart={{.Executable}} backend start --port={{.Port}}
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
[Install]
WantedBy=multi-user.target
`

func installSystemdService(port int) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working dir: %w", err)
	}
	user := os.Getenv("USER")
	if user == "" {
		user = "root"
	}

	data := struct {
		User, WorkingDir, Executable string
		Port                         int
	}{user, workingDir, execPath, port}

	tmpl, err := template.New("service").Parse(serviceTmpl)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}
	return os.WriteFile("/etc/systemd/system/sl-cli.service", buf.Bytes(), 0644)
}

func uninstallSystemdService() error {
	return os.Remove("/etc/systemd/system/sl-cli.service")
}
