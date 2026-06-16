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
			fmt.Fprintf(os.Stderr, "Error installing systemd service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Systemd service installed successfully!")
		fmt.Println("Enable and start with:")
		fmt.Println("  sudo systemctl enable sl-cli")
		fmt.Println("  sudo systemctl start sl-cli")
		fmt.Println("Check status with:")
		fmt.Println("  sudo systemctl status sl-cli")
	},
}

var systemdUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall systemd service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := uninstallSystemdService(); err != nil {
			fmt.Fprintf(os.Stderr, "Error uninstalling systemd service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Systemd service uninstalled successfully!")
	},
}

func init() {
	systemdInstallCmd.Flags().Int("port", 3099, "Port for HTTP server")
	systemdCmd.AddCommand(systemdInstallCmd)
	systemdCmd.AddCommand(systemdUninstallCmd)
}

const systemdServiceTemplate = `[Unit]
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
	// In production, you might want to make this configurable
	user := os.Getenv("USER")
	if user == "" {
		user = "root"
	}

	// Create service file content
	data := struct {
		User        string
		WorkingDir  string
		Executable  string
		Port        int
	}{
		User:       user,
		WorkingDir: workingDir,
		Executable: execPath,
		Port:       port,
	}

	tmpl, err := template.New("service").Parse(systemdServiceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var serviceContent bytes.Buffer
	if err := tmpl.Execute(&serviceContent, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Write service file
	servicePath := "/etc/systemd/system/sl-cli.service"
	if err := os.WriteFile(servicePath, serviceContent.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd
	// Note: This requires sudo, will fail if not run as root
	// exec.Command("systemctl", "daemon-reload").Run()

	return nil
}

func uninstallSystemdService() error {
	servicePath := "/etc/systemd/system/sl-cli.service"
	if err := os.Remove(servicePath); err != nil {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd
	// exec.Command("systemctl", "daemon-reload").Run()

	return nil
}