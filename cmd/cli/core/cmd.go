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
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"oadin/config"
	"oadin/console"
	"oadin/internal/api"
	"oadin/internal/api/dto"
	"oadin/internal/datastore"
	"oadin/internal/datastore/jsonds"
	jsondsTemplate "oadin/internal/datastore/jsonds/data"
	"oadin/internal/datastore/sqlite"
	"oadin/internal/event"
	"oadin/internal/provider"
	"oadin/internal/schedule"
	"oadin/internal/server"
	"oadin/internal/types"
	"oadin/internal/utils"
	"oadin/internal/utils/bcode"
	"oadin/internal/utils/progress"
	"oadin/tray"
	"oadin/version"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// ServerManager 用于管理服务器实例
type ServerManager struct {
	oadinServer   *http.Server
	consoleServer *http.Server
	cancel        context.CancelFunc
	trayManager   *tray.Manager
}

var globalServerManager *ServerManager

// StopServer 停止指定的服务器
func (sm *ServerManager) StopServer(serverType string) error {
	var srv *http.Server
	switch serverType {
	case "oadin":
		srv = sm.oadinServer
	case "console":
		srv = sm.consoleServer
	default:
		return fmt.Errorf("unknown server type: %s", serverType)
	}

	if srv == nil {
		return fmt.Errorf("server %s is not running", serverType)
	}

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv.SetKeepAlivesEnabled(false)
	if err := srv.Shutdown(ctx); err != nil {
		if err := srv.Close(); err != nil {
			return fmt.Errorf("failed to stop %s server: %v", serverType, err)
		}
	}

	switch serverType {
	case "oadin":
		sm.oadinServer = nil
	case "console":
		sm.consoleServer = nil
	}

	// 如果两个服务器都已停止，则取消主context
	if sm.oadinServer == nil && sm.consoleServer == nil && sm.cancel != nil {
		sm.cancel()
	}

	return nil
}

// NewCommand will contain all commands
func NewCommand() *cobra.Command {
	cmds := &cobra.Command{
		Use: "oadin",
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
		Use:   "provider <provider_name>",
		Short: "Edit service data",
		Long:  "Edit service status and scheduler policy",
		Args:  cobra.ExactArgs(1),
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
		Use:   "service <service_name>",
		Short: "Edit service data",
		Long:  "Edit service status and scheduler policy",
		Args:  cobra.ExactArgs(1),
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

			c := config.NewOadinClient()
			routerPath := fmt.Sprintf("/oadin/%s/service", version.OadinVersion)

			err = c.Client.Do(context.Background(), http.MethodPut, routerPath, req, &resp)
			if err != nil {
				return
			}
			if resp.HTTPCode > 200 {
				fmt.Println(resp.Message)
			}
		},
	}

	updateServiceCmd.Flags().StringVar(&hybridPolicy, "hybrid_policy", "default", "only support default/always_local/always_remote.")
	updateServiceCmd.Flags().StringVarP(&remoteProvider, "remote_provider", "", "", "remote ai service provider")
	updateServiceCmd.Flags().StringVarP(&localProvider, "local_provider", "", "", "local ai service provider")

	return updateServiceCmd
}

func Run(ctx context.Context) error {
	// Initialize the datastore
	ds, err := sqlite.New(config.GlobalOadinEnvironment.Datastore)
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
	jds := jsonds.NewJSONDatastore(jsondsTemplate.JsonDataStoreFS)

	err = jds.Init()
	if err != nil {
		fmt.Printf("Failed to initialize json file store: %v\n", err)
	}
	datastore.SetDefaultJsonDatastore(jds)

	if server.UseVSSForPlayground() {
		go func() {
			dbPath := config.GlobalOadinEnvironment.Datastore
			if err := server.InitPlaygroundVec(ctx, dbPath); err != nil {
				slog.Error("Failed to initialize VSS database", "error", err)
			} else {
				slog.Info("VSS database initialized successfully")
			}
		}()
	}

	// version.StartCheckUpdate(ctx)
	// Initialize core core app server
	oadinServer := api.NewOadinCoreServer()
	oadinServer.Register()

	event.InitSysEvents()
	event.SysEvents.Notify("start_app", nil)

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
	api.InjectRouter(oadinServer)

	// Inject all flavors to the router
	// Setup flavors
	for _, flavor := range schedule.AllAPIFlavors() {
		flavor.InstallRoutes(oadinServer.Router, config.GlobalOadinEnvironment)
		schedule.InitProviderDefaultModelTemplate(flavor)
	}

	pidFile := filepath.Join(config.GlobalOadinEnvironment.RootDir, "oadin.pid")
	err = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0o644)
	if err != nil {
		slog.Error("[Run] Failed to write pid file", "error", err)
		return err
	}

	go ListenModelEngineHealth()

	// create a cancel context
	ctx, cancel := context.WithCancel(ctx)

	// create error chan
	errChan := make(chan error, 2)

	// create server manager
	globalServerManager = &ServerManager{
		cancel: cancel,
	}

	// start oadin server
	oadinSrv := &http.Server{
		Addr:    config.GlobalOadinEnvironment.ApiHost,
		Handler: oadinServer.Router,
	}
	globalServerManager.oadinServer = oadinSrv

	go func() {
		if err := oadinSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("oadin server error: %v", err)
		}
	}()

	// start console server
	consoleSrv, err := console.StartConsoleServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to start console server: %v", err)
	}
	globalServerManager.consoleServer = consoleSrv

	_, _ = color.New(color.FgHiGreen).Println("Oadin Gateway starting on port", config.GlobalOadinEnvironment.ApiHost)
	_, _ = color.New(color.FgHiGreen).Println("Console server starting on port :16699")

	// create tray manager
	trayManager := tray.NewManager(
		func() error {
			if globalServerManager.oadinServer != nil {
				return fmt.Errorf("server is already running")
			}
			oadinSrv = &http.Server{
				Addr:    config.GlobalOadinEnvironment.ApiHost,
				Handler: oadinServer.Router,
			}
			go func() {
				if err := oadinSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					errChan <- fmt.Errorf("oadin server error: %v", err)
				}
			}()
			globalServerManager.oadinServer = oadinSrv
			return nil
		},
		func() error {
			if globalServerManager.oadinServer == nil {
				return fmt.Errorf("server is not running")
			}
			return globalServerManager.StopServer("oadin")
		},
		func() error {
			var errs []error
			if globalServerManager.oadinServer != nil {
				if err := globalServerManager.StopServer("oadin"); err != nil {
					errs = append(errs, fmt.Errorf("failed to stop oadin server: %v", err))
				}
			}
			if globalServerManager.consoleServer != nil {
				if err := globalServerManager.StopServer("console"); err != nil {
					errs = append(errs, fmt.Errorf("failed to stop console server: %v", err))
				}
			}
			if len(errs) > 0 {
				return fmt.Errorf("errors stopping servers: %v", errs)
			}
			return nil
		},
		func() error {
			stopCmd := exec.Command("oadin", "server", "stop")
			return stopCmd.Run()
		},
		true,
	)
	globalServerManager.trayManager = trayManager

	tray.StartCheckUpdate(ctx, trayManager)
	// start tray
	trayManager.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case <-sigChan:
		cancel()
		// 等待最多5秒钟让服务器优雅关闭
		select {
		case err := <-errChan:
			return err
		case <-time.After(5 * time.Second):
			return nil
		}
	}
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

	c := config.NewOadinClient()
	routerPath := fmt.Sprintf("/oadin/%s/service_provider", version.OadinVersion)

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
		Short: "Manage oadin server",
		Long:  "Manage oadin server (start, stop, etc.)",
	}

	cmd.AddCommand(
		NewStartApiServerCommand(),
		NewStopApiServerCommand(),
		NewStopServerCommand(),
	)

	return cmd
}

func NewStopApiServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop oadin server daemon",
		Long:  "Stop oadin server daemon",
		Args:  cobra.ExactArgs(0), // 不需要参数
		RunE:  stopOadinServer,
	}
}

func stopOadinServer(cmd *cobra.Command, args []string) error {
	pidFile := filepath.Join(config.GlobalOadinEnvironment.RootDir, "*.pid")
	fmt.Printf("stop oadin server daemon, pidFile: %s\n", pidFile)
	err := utils.StopOadinServer(pidFile)
	if err != nil {
		fmt.Printf("[Stop] Failed to stop oadin server err: %s", err)
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
		PreRun: CheckOadinServer,
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
		Short: "Prints oadin build version information.",
		Long:  "Prints oadin build version information.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Oadin Version:%s\nOadin SubVersion:%s",
				version.OadinVersion, version.OadinSubVersion)
		},
	}

	return ver
}

// NewStartApiServerCommand  Create a new cobra.Command Object with default values.
func NewStartApiServerCommand() *cobra.Command {
	config.GlobalOadinEnvironment = config.NewOadinEnvironment()
	cmd := &cobra.Command{
		Use:   "start",
		Short: "oadin apiserver is a aipc open gateway",
		Long:  "oadin apiserver is a aipc open gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			isDaemon, err := cmd.Flags().GetBool("d")
			if err != nil {
				return err
			}
			if isDaemon {
				CheckOadinServer(cmd, args)
				return nil
			}
			return Run(context.Background())
		},
	}
	fs := cmd.Flags()
	namedFlagSets := config.GlobalOadinEnvironment.Flags()
	for _, set := range namedFlagSets.FlagSets {
		fs.AddFlagSet(set)
	}
	fs.BoolP("d", "d", false, "Start the server in daemon mode")
	return cmd
}

func NewInstallModelCommand() *cobra.Command {
	var (
		serviceName  string
		providerName string
		remote       bool
	)

	pullModelCmd := &cobra.Command{
		Use:   "pull <model_name>",
		Short: "Pull a model for a specific service",
		Long:  `Pull a model for a specific service with optional remote flag.`,
		Args:  cobra.ExactArgs(1),
		Run:   PullHandler,
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
		Use:   "model <model_name>",
		Short: "Remove a model for a specific service",
		Long:  `Remove a model for a specific service with optional remote flag.`,
		Args:  cobra.ExactArgs(1),
		Run:   DeleteModelHandler,
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
		Use:   "service_provider <provider_name>",
		Short: "Remove a provider for a specific service",
		Long:  `Remove a provider for a specific service with optional remote flag.`,
		Args:  cobra.ExactArgs(1),
		Run:   DeleteProviderHandler,
	}

	return deleteProviderCmd
}

func NewListServicesCommand() *cobra.Command {
	listModelCmd := &cobra.Command{
		Use:   "services <service_name>",
		Short: "Display all available service information.",
		Long:  `Display all available service information.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req := dto.GetAIGCServicesRequest{}
			resp := dto.GetAIGCServicesResponse{}

			if len(args) > 0 {
				req.ServiceName = args[0]
			}

			c := config.NewOadinClient()
			routerPath := fmt.Sprintf("/oadin/%s/service", version.OadinVersion)

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
		Use:   "models",
		Short: "List models for a specific service",
		Long:  `List models for a specific service.`,
		Run: func(cmd *cobra.Command, args []string) {
			req := dto.GetModelsRequest{}
			resp := dto.GetModelsResponse{}

			if providerName != "" {
				req.ProviderName = providerName
			}

			c := config.NewOadinClient()
			routerPath := fmt.Sprintf("/oadin/%s/model", version.OadinVersion)

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
		Use:   "service_providers",
		Short: "List models for a specific service",
		Long:  `List models for a specific service.`,
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

			c := config.NewOadinClient()
			routerPath := fmt.Sprintf("/oadin/%s/service_provider", version.OadinVersion)

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
		return fmt.Errorf("failed to parse configuration file: %w", err)
	}

	if spConf.ServiceName == "" || spConf.ServiceSource == "" || spConf.ApiFlavor == "" || len(spConf.Models) == 0 {
		return fmt.Errorf("service_name, service_source, flavor_name, and models are required")
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

	c := config.NewOadinClient()
	routerPath := fmt.Sprintf("/oadin/%s/service_provider", version.OadinVersion)

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

		c := config.NewOadinClient()
		routerPath := fmt.Sprintf("/oadin/%s/service", version.OadinVersion)

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

		//if !remote && serviceName == types.ServiceChat {
		//	askRes := askEnableRemoteService()
		//	if askRes {
		//		fmt.Println("请前往 https://platform.deepseek.com/ 网址申请 APIKEY。")
		//		apiKey := getAPIKey()
		//		if apiKey != "" {
		//			fmt.Printf("\r你输入的 API Key 是: %s\n", apiKey)
		//		}
		//
		//		req := &dto.CreateAIGCServiceRequest{
		//			ServiceName:   "chat",
		//			ServiceSource: "remote",
		//			ApiFlavor:     "deepseek",
		//			ProviderName:  "remote_deepseek_chat",
		//			Desc:          "remote deepseek model service",
		//			Method:        http.MethodPost,
		//			Url:           "https://api.lkeap.cloud.tencent.com/v1/chat/completions",
		//			AuthType:      "apikey",
		//			AuthKey:       apiKey,
		//			ExtraHeaders:  "{}",
		//			ExtraJsonBody: "{}",
		//			Properties:    `{"max_input_tokens": 131072,"supported_response_mode":["stream","sync"]}`,
		//		}
		//
		//		err := c.Client.Do(context.Background(), http.MethodPost, routerPath, req, &resp)
		//		if err != nil {
		//			fmt.Printf("\rService install failed: %s ", err.Error())
		//			return
		//		}
		//	} else {
		//		fmt.Println("下次您可以通过 oadin install chat -r --flavor deepseek --auth_type apikey 来启用远程DeepSeek服务")
		//	}
		//}
	}
}

func CheckOadinServer(cmd *cobra.Command, args []string) {

	// Check if the model engine service is started
	engineProvider := provider.GetModelEngine("ollama")
	engineConfig := engineProvider.GetConfig()

	err := engineProvider.HealthCheck()
	if err != nil {
		checkCmd := exec.Command(engineConfig.ExecPath+"/"+engineConfig.ExecFile, "-h")
		err = checkCmd.Run()
		if err != nil {
			slog.Info("Model engine not exist...")
			slog.Info("model engine not exist, start download...")
			err = engineProvider.InstallEngine()
			if err != nil {
				fmt.Println("Install model engine failed :", err.Error())
				slog.Error("Install model engine failed :", err.Error())
				log.Fatalf("Install model engine failed err %s", err.Error())
				return
			}
			slog.Info("Model engine download completed...")
		}

		slog.Info("Setting env...")
		err := engineProvider.InitEnv()
		if err != nil {
			slog.Error("Setting env error: ", err.Error())
			log.Fatalf("Setting env error %s", err.Error())
			return
		}

		slog.Info("Start model engine...")
		retryCount := 1
		for {
			if retryCount > 5 {
				break
			}
			err = engineProvider.StartEngine()
			if err != nil {
				slog.Error("Start engine error: ", err.Error())
				retryCount += 1
				time.Sleep(1 * time.Second)
				continue
			}
			break
		}

		slog.Info("Waiting ollama engine start 60s...")
		for i := 60; i > 0; i-- {
			time.Sleep(time.Second * 1)
			err = engineProvider.HealthCheck()
			if err == nil {
				slog.Info("Start model engine completed...")
				break
			}
			slog.Info("Waiting ollama engine start ...", strconv.Itoa(i), "s")
		}
	}

	err = engineProvider.HealthCheck()
	if err != nil {
		slog.Error("Ollama engine failed start, Please try again later...")
		log.Fatalf("Ollama engine failed err %s", err.Error())
		return
	}

	if utils.IsServerRunning() {
		fmt.Println("Oadin server start successfully.")
		return
	}
	//userDir, _ := os.UserHomeDir()
	//logger.NewSysLogger(logger.NewLogConfig{LogLevel: "debug", LogPath: filepath.Join(userDir, "server.log")})
	fmt.Println("Oadin server is not running. Starting the server...")
	if err := startOadinServer(); err != nil {
		log.Fatalf("Failed to start Oadin server: %s \n", err.Error())
		return
	}

	time.Sleep(6 * time.Second)

	if !utils.IsServerRunning() {
		log.Fatal("Failed to start Oadin server.")
		return
	}

	fmt.Println("Oadin server start successfully.")
}

func startOadinServer() error {
	logPath := config.GlobalOadinEnvironment.ConsoleLog
	rootDir := config.GlobalOadinEnvironment.RootDir
	err := utils.StartOadinServer(logPath, rootDir)
	if err != nil {
		fmt.Printf("oadin server start failed: %s", err.Error())
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

	//var wg sync.WaitGroup
	//stopChan := make(chan struct{})
	//
	//wg.Add(1)
	//msg := "Pulling model"
	//go progress.ShowLoadingAnimation(stopChan, &wg, msg)

	c := config.NewOadinClient()
	routerPath := fmt.Sprintf("/oadin/%s/model", version.OadinVersion)

	err = c.Client.Do(context.Background(), http.MethodPost, routerPath, req, &resp)
	if err != nil {
		fmt.Printf("\rPull model failed: %s", err.Error())
		return
	}

	// close(stopChan)
	// wg.Wait()

	if resp.HTTPCode > 200 {
		fmt.Printf("\rPull model  failed: %s", resp.Message)
		return
	}

	fmt.Println("Pull model success!")
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

	c := config.NewOadinClient()
	routerPath := fmt.Sprintf("/oadin/%s/model", version.OadinVersion)

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

	c := config.NewOadinClient()
	routerPath := fmt.Sprintf("/oadin/%s/service_provider", version.OadinVersion)

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
		Use:   "import <file_path>",
		Short: "Import service configuration from a file",
		Long:  "Import service configuration from a file and send it to the API.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("please provide a .oadin file path")
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
				return fmt.Errorf("failed to parse file content: %w", err)
			}

			var wg sync.WaitGroup
			stopChan := make(chan struct{})
			wg.Add(1)
			msg := "Importing service configuration"
			go progress.ShowLoadingAnimation(stopChan, &wg, msg)

			c := config.NewOadinClient()
			routerPath := fmt.Sprintf("/oadin/%s/service/import", version.OadinVersion)

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
		Use:   "to-file",
		Short: "Export service to file",
		Long:  "Export service to file",
		Run: func(cmd *cobra.Command, args []string) {
			req := &dto.ExportServiceRequest{
				ServiceName:  service,
				ProviderName: provider,
				ModelName:    model,
			}
			resp := &dto.ExportServiceResponse{}

			c := config.NewOadinClient()
			routerPath := fmt.Sprintf("/oadin/%s/service/export", version.OadinVersion)

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

	cmd.Flags().StringVarP(&filePath, "file", "f", "./.oadin", "Output file path")

	return cmd
}

func NewExportServiceToStdoutCommand(service, provider, model string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "to-stdout",
		Short: "Export service to stdout",
		Long:  "Export service to stdout",
		Run: func(cmd *cobra.Command, args []string) {
			req := &dto.ExportServiceRequest{
				ServiceName:  service,
				ProviderName: provider,
				ModelName:    model,
			}
			resp := &dto.ExportServiceResponse{}

			c := config.NewOadinClient()
			routerPath := fmt.Sprintf("/oadin/%s/service/export", version.OadinVersion)

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
	time.Sleep(10 * time.Second)
	ds := datastore.GetDefaultDatastore()

	sp := &types.ServiceProvider{
		ServiceSource: types.ServiceSourceLocal,
	}

	OllamaEngine := provider.GetModelEngine(types.FlavorOllama)

	for {
		list, err := ds.List(context.Background(), sp, &datastore.ListOptions{Page: 0, PageSize: 100})
		if err != nil {
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
		for _, engineName := range engineList {
			if engineName == types.FlavorOllama {
				operateStatus := OllamaEngine.GetOperateStatus()
				if operateStatus == 0 {
					continue
				}
				err := OllamaEngine.HealthCheck()
				if err != nil {
					err = OllamaEngine.InitEnv()
					if err != nil {
						slog.Error("[Engine health]Setting env error: ", err.Error())
						continue
					}
					err := OllamaEngine.StartEngine()
					if err != nil {
						continue
					}
				}
			}
		}

		time.Sleep(60 * time.Second)
	}
}

// NewStopServerCommand 创建停止指定服务器的命令
func NewStopServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop [oadin|console]",
		Short: "Stop a specific server",
		Long:  "Stop either the oadin server or the console server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverType := args[0]
			if serverType != "oadin" && serverType != "console" {
				return fmt.Errorf("invalid server type: %s. Must be either 'oadin' or 'console'", serverType)
			}

			if globalServerManager == nil {
				return fmt.Errorf("server manager is not initialized")
			}

			if err := globalServerManager.StopServer(serverType); err != nil {
				return err
			}

			fmt.Printf("%s server stopped successfully\n", serverType)
			return nil
		},
	}
	return cmd
}
