package cli

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

var systemdCmd = &cobra.Command{
	Use:   "systemd",
	Short: "Manage systemd service",
	Long:  `Install, uninstall, and manage systemd service for boot-persistent daemon.`,
}

var systemdInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install systemd service",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		if err := installSystemdService(port); err != nil {
			fail(ExitExtFailed, err.Error())
		}
		success("Systemd service installed successfully", map[string]interface{}{
			"port": port, "service": "sl-cli", "path": "/etc/systemd/system/sl-cli.service",
		})
	},
}

var systemdUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall systemd service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := uninstallSystemdService(); err != nil {
			fail(ExitExtFailed, err.Error())
		}
		success("Systemd service uninstalled successfully", nil)
	},
}

func init() {
	systemdInstallCmd.Flags().Int("port", 3099, "Port for HTTP server")
	systemdCmd.AddCommand(systemdInstallCmd)
	systemdCmd.AddCommand(systemdUninstallCmd)
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
