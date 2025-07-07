//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"intel.com/aog/config"
	"intel.com/aog/console"
	"intel.com/aog/internal/api"
	"intel.com/aog/internal/api/dto"
	"intel.com/aog/internal/constants"
	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/datastore/jsonds"
	jsondsTemplate "intel.com/aog/internal/datastore/jsonds/data"
	"intel.com/aog/internal/datastore/sqlite"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/provider"
	"intel.com/aog/internal/schedule"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
	"intel.com/aog/internal/utils/bcode"
	"intel.com/aog/internal/utils/progress"
	"intel.com/aog/version"
)

// NewCommand will contain all commands
func NewCommand() *cobra.Command {
	cmds := &cobra.Command{
		Use: constants.AppName,
	}

	cmds.AddCommand(
		// Common
		NewApiserverCommand(),
		NewVersionCommand(),

		NewGetCommand(),
		NewInstallServiceCommand(),
		NewEditCommand(),
		NewDeleteCommand(),

		// Models
		NewInstallModelCommand(),

		// Export/Import
		NewExportServiceCommand(),
		NewImportServiceCommand(),
	)

	return cmds
}

func NewEditCommand() *cobra.Command {
	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit resources",
	}
	editCmd.AddCommand(NewEditServiceCommand())
	editCmd.AddCommand(NewEditProviderCommand())

	return editCmd
}

func NewEditProviderCommand() *cobra.Command {
	var filePath string

	editProviderCmd := &cobra.Command{
		Use:    "provider <provider_name>",
		Short:  "Edit service data",
		Long:   "Edit service status and scheduler policy",
		Args:   cobra.ExactArgs(1),
		PreRun: CheckAOGServer,
		Run: func(cmd *cobra.Command, args []string) {
			filePath, err := cmd.Flags().GetString("file")
			if err != nil {
				fmt.Println("Error: failed to get file path")
				return
			}
			if filePath == "" {
				fmt.Println("Error: file path is required for service_provider")
				return
			}
			err = updateServiceProviderHandler(args[0], filePath)
			if err != nil {
				fmt.Println("Error: service provider install failed ", err.Error())
				return
			}
		},
	}

	editProviderCmd.Flags().StringVarP(&filePath, "file", "f", "", "service provider config file path")

	return editProviderCmd
}

func NewEditServiceCommand() *cobra.Command {
	var hybridPolicy string
	var remoteProvider string
	var localProvider string

	updateServiceCmd := &cobra.Command{
		Use:    "service <service_name>",
		Short:  "Edit service data",
		Long:   "Edit service status and scheduler policy",
		Args:   cobra.ExactArgs(1),
		PreRun: CheckAOGServer,
		Run: func(cmd *cobra.Command, args []string) {
			serviceName := args[0]
			hybridPolicy, err := cmd.Flags().GetString("hybrid_policy")
			remoteProvider, err := cmd.Flags().GetString("remote_provider")
			localProvider, err := cmd.Flags().GetString("local_provider")
			if err != nil {
				fmt.Println("An error occurred while obtaining the hybrid_policy parameter:", err)
				os.Exit(1)
			}

			if !utils.Contains(types.SupportHybridPolicy, hybridPolicy) {
				fmt.Printf("\rInvalid hybrid_policy value: %s，Allowed values are.: %v\n", hybridPolicy, types.SupportHybridPolicy)
				os.Exit(1)
			}

			req := dto.UpdateAIGCServiceRequest{
				ServiceName:  serviceName,
				HybridPolicy: hybridPolicy,
			}
			resp := bcode.Bcode{}

			if remoteProvider != "" {
				req.RemoteProvider = remoteProvider
			}
			if localProvider != "" {
				req.LocalProvider = localProvider
			}

			c := config.NewAOGClient()
			routerPath := fmt.Sprintf("/%s/%s/service", constants.AppName, version.AOGVersion)

			err = c.Client.Do(context.Background(), http.MethodPut, routerPath, req, &resp)
			if err != nil {
				return
			}
			if resp.HTTPCode > 200 {
				fmt.Println(resp.Message)
			}
			fmt.Printf("Service edit success!")
		},
	}

	updateServiceCmd.Flags().StringVar(&hybridPolicy, "hybrid_policy", "default", "only support default/always_local/always_remote.")
	updateServiceCmd.Flags().StringVarP(&remoteProvider, "remote_provider", "", "", "remote ai service provider")
	updateServiceCmd.Flags().StringVarP(&localProvider, "local_provider", "", "", "local ai service provider")

	return updateServiceCmd
}

func Run(ctx context.Context) error {
	// Initialize the datastore
	ds, err := sqlite.New(config.GlobalAOGEnvironment.Datastore)
	if err != nil {
		slog.Error("[Init] Failed to load datastore", "error", err)
		return err
	}

	err = ds.Init()
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return err
	}

	datastore.SetDefaultDatastore(ds)

	// 初始化 control panel
	jds := jsonds.NewJSONDatastore(jsondsTemplate.JsonDataStoreFS)

	err = jds.Init()
	if err != nil {
		fmt.Printf("Failed to initialize json file store: %v\n", err)
	}
	datastore.SetDefaultJsonDatastore(jds)

	logger.InitLogger(
		logger.LogConfig{
			LogLevel: config.GlobalAOGEnvironment.LogLevel,
			LogPath:  config.GlobalAOGEnvironment.LogDir,
		})
	// Initialize core core app server
	aogServer := api.NewAOGCoreServer()
	aogServer.Register()

	logger.LogicLogger.Info("start_app")

	// load all flavors
	// this loads all config based API Flavors. You need to manually
	// create and RegisterAPIFlavor for costimized API Flavors
	err = schedule.InitAPIFlavors()
	if err != nil {
		slog.Error("[Init] Failed to load API Flavors", "error", err)
		return nil
	}

	// start
	schedule.StartScheduler("basic")

	// Inject the router
	api.InjectRouter(aogServer)

	// Inject all flavors to the router
	// Setup flavors
	for _, flavor := range schedule.AllAPIFlavors() {
		flavor.InstallRoutes(aogServer.Router)
		schedule.InitProviderDefaultModelTemplate(flavor)
	}

	pidFile := filepath.Join(config.GlobalAOGEnvironment.RootDir, constants.AppName+".pid")
	err = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0o644)
	if err != nil {
		slog.Error("[Run] Failed to write pid file", "error", err)
		return err
	}

	_ = console.RegisterConsoleRoutes(aogServer.Router)

	go ListenModelEngineHealth()

	// Run the server
	err = aogServer.Run(ctx, config.GlobalAOGEnvironment.ApiHost)
	if err != nil {
		slog.Error("[Run] Failed to run server", "error", err)
		return err
	}

	_, _ = color.New(color.FgHiGreen).Println("AOG Gateway starting on port", config.GlobalAOGEnvironment.ApiHost)

	return nil
}

func updateServiceProviderHandler(providerName, configFile string) error {
	if configFile == "" {
		return fmt.Errorf("configuration file is required")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %w", err)
	}

	var spConf dto.UpdateServiceProviderRequest
	err = json.Unmarshal(data, &spConf)
	if err != nil {
		return fmt.Errorf("failed to parse configuration file: %w", err)
	}

	if spConf.ServiceName == "" || spConf.ServiceSource == "" || spConf.ApiFlavor == "" {
		return fmt.Errorf("service_name, service_source, flavor_name are required")
	}

	if spConf.AuthType != "none" && spConf.AuthKey == "" {
		return fmt.Errorf("auth_key is required when auth_type is not none")
	}

	if spConf.ProviderName != providerName {
		return fmt.Errorf("please check whether the provider name is the same as the provider name in the file")
	}

	resp := dto.UpdateServiceProviderResponse{}

	c := config.NewAOGClient()
	routerPath := fmt.Sprintf("/%s/%s/service_provider", constants.AppName, version.AOGVersion)

	err = c.Client.Do(context.Background(), http.MethodPut, routerPath, spConf, &resp)
	if err != nil {
		fmt.Printf("\rService provider edit failed: %s", err.Error())
		return err
	}

	if resp.HTTPCode > 200 {
		fmt.Printf("\rService provider edit failed: %s", resp.Message)
		return err
	}

	fmt.Println("Service provider edit success!")

	return nil
}

func NewGetCommand() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get resources",
	}
	getCmd.AddCommand(NewListServicesCommand())
	getCmd.AddCommand(NewListModelsCommand())
	getCmd.AddCommand(NewListProvidersCommand())

	return getCmd
}

func NewDeleteCommand() *cobra.Command {
	editCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources",
	}
	editCmd.AddCommand(NewDeleteModelCommand())
	editCmd.AddCommand(NewDeleteProviderCommand())

	return editCmd
}

func NewApiserverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Manage " + constants.AppName + " server",
		Long:  "Manage " + constants.AppName + " server (start, stop, etc.)",
	}

	cmd.AddCommand(
		NewStartApiServerCommand(),
		NewStopApiServerCommand(),
	)

	return cmd
}

func NewStopApiServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop daemon server.",
		Long:  "Stop daemon server.",
		Args:  cobra.ExactArgs(0), // 不需要参数
		RunE:  stopAogServer,
	}
}

func stopAogServer(cmd *cobra.Command, args []string) error {
	files, err := filepath.Glob(filepath.Join(config.GlobalAOGEnvironment.RootDir, "*.pid"))
	if err != nil {
		return fmt.Errorf("failed to list pid files: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No running processes found")
		return nil
	}

	// Traverse all pid files.
	for _, pidFile := range files {
		pidData, err := os.ReadFile(pidFile)
		if err != nil {
			fmt.Printf("Failed to read PID file %s: %v\n", pidFile, err)
			continue
		}

		pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
		if err != nil {
			fmt.Printf("Invalid PID in file %s: %v\n", pidFile, err)
			continue
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Printf("Failed to find process with PID %d: %v\n", pid, err)
			continue
		}

		if err := process.Kill(); err != nil {
			if strings.Contains(err.Error(), "process already finished") {
				fmt.Printf("Process with PID %d is already stopped\n", pid)
			} else {
				fmt.Printf("Failed to kill process with PID %d: %v\n", pid, err)
				continue
			}
		} else {
			fmt.Printf("Successfully stopped process with PID %d\n", pid)
		}

		// remove pid file
		if err := os.Remove(pidFile); err != nil {
			fmt.Printf("Failed to remove PID file %s: %v\n", pidFile, err)
		}
	}
	if runtime.GOOS == "windows" {
		extraProcessName := "ollama-lib.exe"
		extraCmd := exec.Command("taskkill", "/IM", extraProcessName, "/F")
		_, err := extraCmd.CombinedOutput()
		if err != nil {
			// fmt.Printf("failed to kill process: %s", extraProcessName)
			return nil
		}

		ovmsProcessName := "ovms.exe"
		ovmsCmd := exec.Command("taskkill", "/IM", ovmsProcessName, "/F")
		_, err = ovmsCmd.CombinedOutput()
		if err != nil {
			// fmt.Printf("failed to kill process: %s", ovmsProcessName)
			return nil
		}

		fmt.Printf("Successfully killed process: %s\n", extraProcessName)
	}

	return nil
}

// NewInstallServiceCommand will install a service
func NewInstallServiceCommand() *cobra.Command {
	var (
		providerName  string
		remoteFlag    bool
		remoteURL     string
		authType      string
		method        string
		authKey       string
		flavor        string
		skipModelFlag bool
		model         string
	)

	installServiceCmd := &cobra.Command{
		Use:    "install <service>",
		Short:  "Install a service or service provider",
		Long:   `Install a service by name or a service provider from a file.`,
		Args:   cobra.ExactArgs(1),
		PreRun: StartAOGServer,
		Run:    InstallServiceHandler,
	}

	installServiceCmd.Flags().BoolVarP(&remoteFlag, "remote", "r", false, "Enable remote connect")
	installServiceCmd.Flags().StringVar(&providerName, "name", "", "Give the service an alias")
	installServiceCmd.Flags().StringVar(&remoteURL, "remote_url", "", "Remote URL for connect")
	installServiceCmd.Flags().StringVar(&authType, "auth_type", "none", "Authentication type (apikey/token/none)")
	installServiceCmd.Flags().StringVar(&method, "method", "POST", "HTTP method (default POST)")
	installServiceCmd.Flags().StringVar(&authKey, "auth_key", "", "Authentication key json format")
	installServiceCmd.Flags().StringVar(&flavor, "flavor", "", "Flavor (tencent/deepseek)")
	installServiceCmd.Flags().StringP("file", "f", "", "Path to the service provider file (required for service_provider)")
	installServiceCmd.Flags().BoolVarP(&skipModelFlag, "skip_model", "", false, "Skip the model download")
	installServiceCmd.Flags().StringVarP(&model, "model_name", "m", "", "Pull model name")

	return installServiceCmd
}

// NewVersionCommand print client version
func NewVersionCommand() *cobra.Command {
	ver := &cobra.Command{
		Use:   "version",
		Short: "Prints build version information.",
		Long:  "Prints build version information.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`AOG Version: %s`,
				version.AOGVersion)
		},
	}

	return ver
}

// NewStartApiServerCommand  Create a new cobra.Command Object with default values.
func NewStartApiServerCommand() *cobra.Command {
	config.GlobalAOGEnvironment = config.NewAOGEnvironment()
	logger.InitLogger(logger.LogConfig{LogLevel: config.GlobalAOGEnvironment.LogLevel, LogPath: config.GlobalAOGEnvironment.LogDir})
	cmd := &cobra.Command{
		Use:   "start",
		Short: "apiserver is a aipc open gateway",
		Long:  "apiserver is a aipc open gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			isDaemon, err := cmd.Flags().GetBool("daemon")
			if err != nil {
				return err
			}
			if isDaemon {
				StartAOGServer(cmd, args)
				return nil
			}

			isDebug, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			startMode := types.EngineStartModeDaemon
			if isDebug {
				startMode = types.EngineStartModeStandard
			}

			err = StartModelEngine("openvino", startMode)
			if err != nil {
				return err
			}

			err = StartModelEngine("ollama", startMode)
			if err != nil {
				return err
			}

			return Run(context.Background())
		},
	}

	cmd.Flags().BoolP("daemon", "d", false, "Start the server in daemon mode")
	cmd.Flags().BoolP("verbose", "v", false, "Enable debug mode")
	return cmd
}

func NewInstallModelCommand() *cobra.Command {
	var (
		serviceName  string
		providerName string
		remote       bool
	)

	pullModelCmd := &cobra.Command{
		Use:    "pull <model_name>",
		Short:  "Pull a model for a specific service",
		Long:   `Pull a model for a specific service with optional remote flag.`,
		Args:   cobra.ExactArgs(1),
		PreRun: CheckAOGServer,
		Run:    PullHandler,
	}

	pullModelCmd.Flags().StringVarP(&serviceName, "for", "f", "", "Name of the service to pull the model for, e.g: chat/embed (required)")
	pullModelCmd.Flags().StringVarP(&providerName, "provider", "p", "", "Name of the service provider to pull the model for, e.g: local_ollama_chat (required)")
	pullModelCmd.Flags().BoolVarP(&remote, "remote", "r", false, "Pull the model from a remote source (default: false)")

	if err := pullModelCmd.MarkFlagRequired("for"); err != nil {
		slog.Error("Error: --for is required")
	}

	return pullModelCmd
}

func NewDeleteModelCommand() *cobra.Command {
	var (
		serviceName  string
		providerName string
		remote       bool
	)

	deleteModelCmd := &cobra.Command{
		Use:    "model <model_name>",
		Short:  "Remove a model for a specific service",
		Long:   `Remove a model for a specific service with optional remote flag.`,
		Args:   cobra.ExactArgs(1),
		PreRun: CheckAOGServer,
		Run:    DeleteModelHandler,
	}

	deleteModelCmd.Flags().StringVarP(&serviceName, "for", "f", "", "Name of the service to delete the model for, e.g: chat/embed (required)")
	deleteModelCmd.Flags().StringVarP(&providerName, "provider", "p", "", "Name of the service provider to remove the model for, e.g: local_ollama_chat (required)") // -p 更常见
	deleteModelCmd.Flags().BoolVarP(&remote, "remote", "r", false, "delete the model from a remote source (default: false)")

	if err := deleteModelCmd.MarkFlagRequired("provider"); err != nil {
		slog.Error("Error: --provider is required")
	}

	return deleteModelCmd
}

func NewDeleteProviderCommand() *cobra.Command {
	deleteProviderCmd := &cobra.Command{
		Use:    "service_provider <provider_name>",
		Short:  "Remove a provider for a specific service",
		Long:   `Remove a provider for a specific service with optional remote flag.`,
		Args:   cobra.ExactArgs(1),
		PreRun: CheckAOGServer,
		Run:    DeleteProviderHandler,
	}

	return deleteProviderCmd
}

func NewListServicesCommand() *cobra.Command {
	listModelCmd := &cobra.Command{
		Use:    "services <service_name>",
		Short:  "Display all available service information.",
		Long:   `Display all available service information.`,
		Args:   cobra.MaximumNArgs(1),
		PreRun: CheckAOGServer,
		Run: func(cmd *cobra.Command, args []string) {
			req := dto.GetAIGCServicesRequest{}
			resp := dto.GetAIGCServicesResponse{}

			if len(args) > 0 {
				req.ServiceName = args[0]
			}

			c := config.NewAOGClient()
			routerPath := fmt.Sprintf("/aog/%s/service", version.AOGVersion)

			err := c.Client.Do(context.Background(), http.MethodGet, routerPath, req, &resp)
			if err != nil {
				fmt.Printf("\rGet service list failed: %s", err.Error())
				return
			}

			fmt.Printf("%-10s %-15s %-15s %-15s %-5s %-15s %-15s\n", "SERVICE NAME", "HYBRID POLICY", "REMOTE PROVIDER", "LOCAL PROVIDER", "STATUS", "CREATE AT", "UPDATE AT") // 表头

			for _, service := range resp.Data {
				serviceStatus := "healthy"
				if service.Status == 0 {
					serviceStatus = "unhealthy"
				}

				fmt.Printf("%-10s %-15s %-15s %-15s %-5s %-15s %-15s\n",
					service.ServiceName,
					service.HybridPolicy,
					service.RemoteProvider,
					service.LocalProvider,
					serviceStatus,
					service.CreatedAt.Format(time.RFC3339),
					service.UpdatedAt.Format(time.RFC3339),
				)
			}
		},
	}

	return listModelCmd
}

func NewListModelsCommand() *cobra.Command {
	var providerName string

	listModelCmd := &cobra.Command{
		Use:    "models",
		Short:  "List models for a specific service",
		Long:   `List models for a specific service.`,
		PreRun: CheckAOGServer,
		Run: func(cmd *cobra.Command, args []string) {
			req := dto.GetModelsRequest{}
			resp := dto.GetModelsResponse{}

			if providerName != "" {
				req.ProviderName = providerName
			}

			c := config.NewAOGClient()
			routerPath := fmt.Sprintf("/aog/%s/model", version.AOGVersion)

			err := c.Client.Do(context.Background(), http.MethodGet, routerPath, req, &resp)
			if err != nil {
				fmt.Printf("\rGet model list failed: %s", err.Error())
				return
			}

			fmt.Printf("%-30s %-25s %-10s %-25s\n", "MODEL NAME", "PROVIDER NAME", "STATUS", "CREATE AT") // 表头

			for _, model := range resp.Data {
				fmt.Printf("%-30s %-20s %-15s %-25s\n",
					model.ModelName,
					model.ProviderName,
					model.Status,
					model.CreatedAt.Format(time.RFC3339),
				)
			}
		},
	}

	listModelCmd.Flags().StringVarP(&providerName, "provider", "p", "", "Name of the service provider, e.g: local_ollama_chat")

	return listModelCmd
}

func NewListProvidersCommand() *cobra.Command {
	var serviceName string
	var providerName string
	var remote string

	listModelCmd := &cobra.Command{
		Use:    "service_providers",
		Short:  "List models for a specific service",
		Long:   `List models for a specific service.`,
		PreRun: CheckAOGServer,
		Run: func(cmd *cobra.Command, args []string) {
			req := dto.GetServiceProvidersRequest{}
			resp := dto.GetServiceProvidersResponse{}

			if serviceName != "" {
				req.ServiceName = serviceName
			}
			if providerName != "" {
				req.ProviderName = providerName
			}
			if remote != "" && (remote == types.ServiceSourceRemote || remote == types.ServiceSourceLocal) {
				req.ServiceSource = remote
			}

			c := config.NewAOGClient()
			routerPath := fmt.Sprintf("/aog/%s/service_provider", version.AOGVersion)

			err := c.Client.Do(context.Background(), http.MethodGet, routerPath, req, &resp)
			if err != nil {
				fmt.Printf("\rGet service provider list failed: %s", err.Error())
				return
			}

			fmt.Printf("%-20s %-10s %-10s %-10s %-10s %-15s %-25s\n", "PROVIDER NAME", "SERVICE NAME", "SERVICE SOURCE", "FLAVOR", "AUTH TYPE", "STATUS", "UPDATE AT") // 表头

			for _, p := range resp.Data {
				providerStatus := "healthy"
				if p.Status == 0 {
					providerStatus = "unhealthy"
				}

				fmt.Printf("%-20s %-10s %-10s %-10s %-10s %-15s %-25s\n",
					p.ProviderName,
					p.ServiceName,
					p.ServiceSource,
					p.Flavor,
					p.AuthType,
					providerStatus,
					p.UpdatedAt.Format(time.RFC3339),
				)
			}
		},
	}

	listModelCmd.Flags().StringVarP(&serviceName, "service", "s", "", "Name of the service to list models for, e.g: chat/embed ")
	listModelCmd.Flags().StringVarP(&providerName, "provider", "p", "", "Name of the service provider, e.g: local_ollama_chat")
	listModelCmd.Flags().StringVarP(&remote, "remote", "r", "", "")

	return listModelCmd
}

func installServiceProviderHandler(configFile string) error {
	if configFile == "" {
		return fmt.Errorf("configuration file is required")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %w", err)
	}

	var spConf dto.CreateServiceProviderRequest
	err = json.Unmarshal(data, &spConf)
	if err != nil {
		return fmt.Errorf("please check your json format: %w", err)
	}

	if spConf.ServiceName == "" || spConf.ServiceSource == "" || spConf.ApiFlavor == "" {
		return fmt.Errorf("service_name, service_source, flavor_name are required")
	}

	if spConf.AuthType != "none" && spConf.AuthKey == "" {
		return fmt.Errorf("auth_key is required when auth_type is not none")
	}

	resp := dto.CreateServiceProviderResponse{}
	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	wg.Add(1)
	msg := "Service provider installing"
	go progress.ShowLoadingAnimation(stopChan, &wg, msg)

	c := config.NewAOGClient()
	routerPath := fmt.Sprintf("/aog/%s/service_provider", version.AOGVersion)

	err = c.Client.Do(context.Background(), http.MethodPost, routerPath, spConf, &resp)
	if err != nil {
		fmt.Printf("\rService provider install failed: %s", err.Error())
		return err
	}

	close(stopChan)
	wg.Wait()

	if resp.HTTPCode > 200 {
		fmt.Printf("\rService provider install failed: %s", resp.Message)
		return err
	}

	fmt.Println("Service provider install success!")

	return nil
}

func InstallServiceHandler(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	if serviceName == "service_provider" {
		filePath, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Println("Error: failed to get file path")
			return
		}
		if filePath == "" {
			fmt.Println("Error: file path is required for service_provider")
			return
		}
		err = installServiceProviderHandler(filePath)
		if err != nil {
			fmt.Println("Error: service provider install failed", err.Error())
			return
		}
	} else {
		remote, err := cmd.Flags().GetBool("remote")
		if err != nil {
			fmt.Println("Error: failed to get remote flag")
			return
		}
		providerName, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Println("Error: failed to get provider name")
			return
		}

		if !utils.Contains(types.SupportService, serviceName) {
			fmt.Printf("\rUnsupported service types: %s", serviceName)
			return
		}

		req := dto.CreateAIGCServiceRequest{}
		resp := bcode.Bcode{}

		if remote {
			method, err := cmd.Flags().GetString("method")
			if err != nil {
				fmt.Println("Error: failed to get method")
				return
			}
			authKey, err := cmd.Flags().GetString("auth_key")
			if err != nil {
				fmt.Println("Error: failed to get auth_key")
				return
			}
			flavorName, err := cmd.Flags().GetString("flavor")
			if err != nil {
				fmt.Println("Error: failed to get flavor")
				return
			}

			if authKey == "" {
				fmt.Println("Error: auth_key is required when auth_type is not none")
				return
			}
			if flavorName != types.FlavorTencent && flavorName != types.FlavorDeepSeek && flavorName != types.FlavorOllama && flavorName != types.FlavorOpenAI {
				fmt.Printf("\rInvalid flavor: %s", flavorName)
				return
			}
			providerInfo := schedule.GetProviderServiceDefaultInfo(flavorName, serviceName)
			req.ServiceSource = types.ServiceSourceRemote
			req.ApiFlavor = flavorName
			req.Url = providerInfo.RequestUrl
			req.AuthType = providerInfo.AuthType
			req.AuthKey = authKey
			req.Method = method
		} else {
			req.ServiceSource = types.ServiceSourceLocal
			req.ApiFlavor = types.FlavorOllama
			if serviceName == types.ServiceTextToImage || serviceName == types.ServiceSpeechToText || serviceName == types.ServiceSpeechToTextWS {
				req.ApiFlavor = types.FlavorOpenvino
			}
			req.AuthType = types.AuthTypeNone
		}
		skipModelFlag, err := cmd.Flags().GetBool("skip_model")
		if err != nil {
			skipModelFlag = false
		}
		modelName, err := cmd.Flags().GetString("model_name")
		if err != nil {
			modelName = ""
		}
		req.SkipModelFlag = skipModelFlag
		req.ModelName = modelName
		req.ServiceName = serviceName
		req.ProviderName = providerName
		if req.ProviderName == "" {
			req.ProviderName = fmt.Sprintf("%s_%s_%s", req.ServiceSource, req.ApiFlavor, req.ServiceName)
		}

		var wg sync.WaitGroup
		stopChan := make(chan struct{})

		wg.Add(1)
		msg := "Service installing"
		go progress.ShowLoadingAnimation(stopChan, &wg, msg)

		c := config.NewAOGClient()
		routerPath := fmt.Sprintf("/aog/%s/service/install", version.AOGVersion)

		err = c.Client.Do(context.Background(), http.MethodPost, routerPath, req, &resp)
		if err != nil {
			fmt.Printf("\rService install failed: %s", err.Error())
			return
		}

		close(stopChan)
		wg.Wait()

		if resp.HTTPCode > 200 {
			fmt.Printf("\rService install failed: %s", resp.Message)
			return
		}

		fmt.Printf("Service %s install success!", serviceName)

		if !remote && serviceName == types.ServiceChat {
			askRes := askEnableRemoteService()
			if askRes {
				fmt.Println("请前往 https://platform.deepseek.com/ 网址申请 APIKEY。")
				apiKey := getAPIKey()
				if apiKey != "" {
					fmt.Printf("\r你输入的 API Key 是: %s\n", apiKey)
				}

				req := &dto.CreateAIGCServiceRequest{
					ServiceName:   "chat",
					ServiceSource: "remote",
					ApiFlavor:     "deepseek",
					ProviderName:  "remote_deepseek_chat",
					Desc:          "remote deepseek model service",
					Method:        http.MethodPost,
					Url:           "https://api.lkeap.cloud.tencent.com/v1/chat/completions",
					AuthType:      "apikey",
					AuthKey:       apiKey,
					ExtraHeaders:  "{}",
					ExtraJsonBody: "{}",
					Properties:    `{"max_input_tokens": 131072,"supported_response_mode":["stream","sync"]}`,
				}

				err := c.Client.Do(context.Background(), http.MethodPost, routerPath, req, &resp)
				if err != nil {
					fmt.Printf("\rService install failed: %s ", err.Error())
					return
				}
			} else {
				fmt.Println("下次您可以通过 aog install chat -r --flavor deepseek --auth_type apikey 来启用远程DeepSeek服务")
			}
		}
	}
}

func CheckAOGServer(cmd *cobra.Command, args []string) {
	if !utils.IsServerRunning() {
		fmt.Println("AOG server is not running, Please run 'aog server start' first")
		os.Exit(1)
		return
	}
}

func StartAOGServer(cmd *cobra.Command, args []string) {
	if utils.IsServerRunning() {
		return
	}

	fmt.Println("AOG server is not running. Starting the server...")
	if err := startAogServer(); err != nil {
		log.Fatalf("Failed to start AOG server: %s \n", err.Error())
		return
	}

	time.Sleep(6 * time.Second)

	if !utils.IsServerRunning() {
		log.Fatal("Failed to start AOG server.")
		return
	}

	err := StartModelEngine("openvino", types.EngineStartModeDaemon)
	if err != nil {
		return
	}

	err = StartModelEngine("ollama", types.EngineStartModeDaemon)
	if err != nil {
		return
	}

	fmt.Println("AOG server start successfully.")
}

func StartModelEngine(engineName, mode string) error {
	if runtime.GOOS == "darwin" {
		return nil
	}
	// Check if the model engine service is started
	engineProvider := provider.GetModelEngine(engineName)
	engineConfig := engineProvider.GetConfig()

	err := engineProvider.HealthCheck()
	if err != nil {
		cmd := exec.Command(engineConfig.ExecPath+engineConfig.ExecFile, "-h")
		err := cmd.Run()
		if err != nil {
			slog.Info("Check model engine " + engineName + " status")
			reCheckCmd := exec.Command(engineConfig.ExecPath+"/"+engineConfig.ExecFile, "-h")
			err = reCheckCmd.Run()
			_, isExistErr := os.Stat(engineConfig.ExecPath + "/" + engineConfig.ExecFile)
			if err != nil && isExistErr != nil {
				slog.Info("Model engine " + engineName + " status: not downloaded")
				return nil
			}
		}

		slog.Info("Setting env...")
		err = engineProvider.InitEnv()
		if err != nil {
			slog.Error("Setting env error: ", err.Error())
			return err
		}

		slog.Info("Start " + engineName + "...")
		err = engineProvider.StartEngine(mode)
		if err != nil {
			slog.Error("Start engine "+engineName+" error: ", err.Error())
			return err
		}

		slog.Info("Waiting model engine " + engineName + " start 60s...")
		for i := 60; i > 0; i-- {
			time.Sleep(time.Second * 1)
			err = engineProvider.HealthCheck()
			if err == nil {
				slog.Info("Start " + engineName + " completed...")
				break
			}
			slog.Info("Waiting "+engineName+" start ...", strconv.Itoa(i), "s")
		}
	}

	err = engineProvider.HealthCheck()
	if err != nil {
		slog.Error(engineName + " failed start, Please try again later...")
		return err
	}

	slog.Info(engineName + " start successfully.")

	return nil
}

func startAogServer() error {
	logPath := config.GlobalAOGEnvironment.ConsoleLog
	rootDir := config.GlobalAOGEnvironment.RootDir
	err := utils.StartAOGServer(logPath, rootDir)
	if err != nil {
		fmt.Printf("AOG server start failed: %s", err.Error())
		return err
	}
	return nil
}

// Get the API Key entered by the user.
func getAPIKey() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("请输入已申请的 API Key: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("读取输入时出错:", err)
		return ""
	}
	return strings.TrimSpace(input)
}

// Ask the user whether to enable remote services.
func askEnableRemoteService() bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("是否启用远程协同的DeepSeek服务？(y/n): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入时出错:", err)
			continue
		}
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "y" {
			return true
		} else if input == "n" {
			return false
		} else {
			fmt.Println("输入无效，请输入 'y' 或 'n'。")
		}
	}
}

func PullHandler(cmd *cobra.Command, args []string) {
	remote, err := cmd.Flags().GetBool("remote")
	if err != nil {
		fmt.Println("Error: failed to get remote flag")
		return
	}
	serviceName, err := cmd.Flags().GetString("for")
	if err != nil {
		fmt.Println("Error: failed to get service name")
		return
	}
	providerName, err := cmd.Flags().GetString("provider")
	if err != nil {
		fmt.Println("Error: failed to get provider name")
		return
	}
	modelName := args[0]

	req := dto.CreateModelRequest{}
	resp := bcode.Bcode{}

	req.ModelName = modelName
	req.ServiceSource = types.ServiceSourceLocal
	if remote {
		req.ServiceSource = types.ServiceSourceRemote
	}
	req.ServiceName = serviceName
	req.ProviderName = providerName

	c := config.NewAOGClient()
	routerPath := fmt.Sprintf("/aog/%s/model", version.AOGVersion)

	err = c.Client.Do(context.Background(), http.MethodPost, routerPath, req, &resp)
	if err != nil {
		fmt.Printf("\rPull model failed: %s", err.Error())
		return
	}

	if resp.HTTPCode > 200 {
		fmt.Printf("\rPull model  failed: %s", resp.Message)
		return
	}

	fmt.Println("You can use the command `aog get models` to check if the model is downloaded successfully.")
	fmt.Println("Model is downloading in the background, please wait...")
}

func DeleteModelHandler(cmd *cobra.Command, args []string) {
	remote, err := cmd.Flags().GetBool("remote")
	if err != nil {
		fmt.Println("Error: failed to get remote flag")
		return
	}
	serviceName, err := cmd.Flags().GetString("for")
	if err != nil {
		fmt.Println("Error: failed to get service name")
		return
	}
	providerName, err := cmd.Flags().GetString("provider")
	if err != nil {
		fmt.Println("Error: failed to get provider name")
		return
	}
	modelName := args[0]

	req := dto.DeleteModelRequest{}
	resp := bcode.Bcode{}

	req.ModelName = modelName
	req.ServiceSource = types.ServiceSourceLocal
	if remote {
		req.ServiceSource = types.ServiceSourceRemote
	}
	req.ServiceName = serviceName
	req.ProviderName = providerName

	c := config.NewAOGClient()
	routerPath := fmt.Sprintf("/aog/%s/model", version.AOGVersion)

	err = c.Client.Do(context.Background(), http.MethodDelete, routerPath, req, &resp)
	if err != nil {
		fmt.Printf("\rDelete model failed: %s", err.Error())
		return
	}

	if resp.HTTPCode > 200 {
		fmt.Printf("\rDelete model  failed: %s", resp.Message)
		return
	}

	fmt.Println("Delete model success!")
}

func DeleteProviderHandler(cmd *cobra.Command, args []string) {
	providerName := args[0]

	req := dto.DeleteServiceProviderRequest{}
	resp := dto.DeleteServiceProviderResponse{}

	req.ProviderName = providerName

	c := config.NewAOGClient()
	routerPath := fmt.Sprintf("/aog/%s/service_provider", version.AOGVersion)

	err := c.Client.Do(context.Background(), http.MethodDelete, routerPath, req, &resp)
	if err != nil {
		fmt.Printf("\rDelete service provider failed: %s", err.Error())
		return
	}

	if resp.HTTPCode > 200 {
		fmt.Printf("\rDelete service provider  failed: %s", resp.Message)
		return
	}

	fmt.Println("Delete service provider success!")
}

func NewImportServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "import <file_path>",
		Short:  "Import service configuration from a file",
		Long:   "Import service configuration from a file and send it to the API.",
		PreRun: CheckAOGServer,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("please provide a .aog file path")
			}
			filePath := args[0]
			// Read the file content
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			// Parse the file content into ImportServiceRequest
			var req dto.ImportServiceRequest
			var resp dto.ImportServiceResponse

			err = json.Unmarshal(data, &req)
			if err != nil {
				return fmt.Errorf("please check your json format: %w", err)
			}

			var wg sync.WaitGroup
			stopChan := make(chan struct{})
			wg.Add(1)
			msg := "Importing service configuration"
			go progress.ShowLoadingAnimation(stopChan, &wg, msg)

			c := config.NewAOGClient()
			routerPath := fmt.Sprintf("/aog/%s/service/import", version.AOGVersion)

			err = c.Client.Do(context.Background(), http.MethodPost, routerPath, req, &resp)
			if err != nil {
				fmt.Printf("\r %s", err.Error())
				return err
			}

			close(stopChan)
			wg.Wait()
			fmt.Println("\rImport service configuration succeeded")

			return nil
		},
	}
	return cmd
}

func NewExportServiceCommand() *cobra.Command {
	var service, serviceProvider, model string
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export service",
		Long:  "Export service",
	}

	exportCmd.Flags().StringVar(&service, "service", "", "Service name")
	exportCmd.Flags().StringVar(&serviceProvider, "provider", "", "Provider name")
	exportCmd.Flags().StringVar(&model, "model", "", "Model name")

	exportCmd.AddCommand(NewExportServiceToFileCommand(service, serviceProvider, model))
	exportCmd.AddCommand(NewExportServiceToStdoutCommand(service, serviceProvider, model))

	return exportCmd
}

func NewExportServiceToFileCommand(service, provider, model string) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:    "to-file",
		Short:  "Export service to file",
		Long:   "Export service to file",
		PreRun: CheckAOGServer,
		Run: func(cmd *cobra.Command, args []string) {
			req := &dto.ExportServiceRequest{
				ServiceName:  service,
				ProviderName: provider,
				ModelName:    model,
			}
			resp := &dto.ExportServiceResponse{}

			c := config.NewAOGClient()
			routerPath := fmt.Sprintf("/aog/%s/service/export", version.AOGVersion)

			err := c.Client.Do(context.Background(), http.MethodPost, routerPath, req, &resp)
			if err != nil {
				fmt.Println("Error exporting service:", err)
				return
			}

			data, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}

			err = os.WriteFile(filePath, data, 0o600)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
			fmt.Println("Exported to file successfully.")
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "./.aog", "Output file path")

	return cmd
}

func NewExportServiceToStdoutCommand(service, provider, model string) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "to-stdout",
		Short:  "Export service to stdout",
		Long:   "Export service to stdout",
		PreRun: CheckAOGServer,
		Run: func(cmd *cobra.Command, args []string) {
			req := &dto.ExportServiceRequest{
				ServiceName:  service,
				ProviderName: provider,
				ModelName:    model,
			}
			resp := &dto.ExportServiceResponse{}

			c := config.NewAOGClient()
			routerPath := fmt.Sprintf("/aog/%s/service/export", version.AOGVersion)

			err := c.Client.Do(context.Background(), http.MethodPost, routerPath, req, &resp)
			if err != nil {
				fmt.Println("Error exporting service:", err)
				return
			}

			data, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}
			fmt.Println(string(data))
		},
	}
	return cmd
}

func ListenModelEngineHealth() {
	ds := datastore.GetDefaultDatastore()

	sp := &types.ServiceProvider{
		ServiceSource: types.ServiceSourceLocal,
	}

	OpenVINOEngine := provider.GetModelEngine(types.FlavorOpenvino)
	OllamaEngine := provider.GetModelEngine(types.FlavorOllama)

	for {
		list, err := ds.List(context.Background(), sp, &datastore.ListOptions{Page: 0, PageSize: 100})
		if err != nil {
			logger.EngineLogger.Error("[Engine Listen]List service provider failed: ", err.Error())
			continue
		}

		if len(list) == 0 {
			continue
		}

		engineList := make([]string, 0)
		for _, item := range list {
			sp := item.(*types.ServiceProvider)
			if utils.Contains(engineList, sp.Flavor) {
				continue
			}

			engineList = append(engineList, sp.Flavor)
		}

		for _, engine := range engineList {
			if engine == types.FlavorOllama {
				err := OllamaEngine.HealthCheck()
				if err != nil {
					logger.EngineLogger.Error("[Engine Listen]Ollama engine health check failed: ", err.Error())
					err := OllamaEngine.InitEnv()
					if err != nil {
						logger.EngineLogger.Error("[Engine Listen]Ollama engine init env failed: ", err.Error())
						return
					}
					err = OllamaEngine.StartEngine(types.EngineStartModeDaemon)
					if err != nil {
						logger.EngineLogger.Error("[Engine Listen]Ollama engine start failed: ", err.Error())
						continue
					}
				}
			} else if engine == types.FlavorOpenvino {
				err := OpenVINOEngine.HealthCheck()
				if err != nil {
					logger.EngineLogger.Error("[Engine Listen]Openvino engine health check failed: ", err.Error())
					err := OpenVINOEngine.StartEngine(types.EngineStartModeDaemon)
					if err != nil {
						logger.EngineLogger.Error("[Engine Listen]Openvino engine start failed: ", err.Error())
						continue
					}
				}
			}
		}

		time.Sleep(60 * time.Second)
	}
}
