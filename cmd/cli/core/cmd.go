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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"intel.com/aog/config"
	"intel.com/aog/internal/api"
	"intel.com/aog/internal/api/dto"
	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/datastore/sqlite"
	"intel.com/aog/internal/event"
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
		Use: "aog",
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
				fmt.Println("获取 hybrid_policy 参数时出错:", err)
				os.Exit(1)
			}

			if !isValidHybridPolicy(hybridPolicy) {
				fmt.Printf("\r无效的 hybrid_policy 值: %s，允许的值为: %v\n", hybridPolicy, allowedHybridPolicies)
				os.Exit(1)
			}

			// todo 调用API 修改服务信息
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
			routerPath := fmt.Sprintf("/aog/%s/service", version.AOGVersion)

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
	ds, err := sqlite.New(config.GlobalAOGEnvironment.Datastore)
	if err != nil {
		slog.Error("[Init] Failed to load datastore", "error", err)
		return err
	}

	//slog.Info(fmt.Sprintf(
	//	"[Init]\nRoot Dir: %s\nWorkDir: %s\nDBPath: %s\nLogPath: %s",
	//	config.GlobalAOGEnvironment.RootDir,
	//	config.GlobalAOGEnvironment.WorkDir,
	//	config.GlobalAOGEnvironment.Datastore,
	//	config.GlobalAOGEnvironment.LogDir,
	//))

	err = ds.Init()
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return err
	}

	datastore.SetDefaultDatastore(ds)

	// Initialize core core app server
	aogServer := api.NewAOGCoreServer()
	aogServer.Register()

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
	api.InjectRouter(aogServer)

	// Inject all flavors to the router
	// Setup flavors
	for _, flavor := range schedule.AllAPIFlavors() {
		flavor.InstallRoutes(aogServer.Router, config.GlobalAOGEnvironment)
		schedule.InitProviderDefaultModelTemplate(flavor)
	}

	pidFile := filepath.Join(config.GlobalAOGEnvironment.RootDir, "aog.pid")
	err = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0o644)
	if err != nil {
		slog.Error("[Run] Failed to write pid file", "error", err)
		return err
	}

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
	routerPath := fmt.Sprintf("/aog/%s/service_provider", version.AOGVersion)

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
		Short: "Manage aog server",
		Long:  "Manage aog server (start, stop, etc.)",
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
		Short: "Stop aog server daemon",
		Long:  "Stop aog server daemon",
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

	// 遍历所有pid文件
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

		// 删除pid文件
		if err := os.Remove(pidFile); err != nil {
			fmt.Printf("Failed to remove PID file %s: %v\n", pidFile, err)
		}
	}

	return nil
}

// 定义允许的 hybrid_policy 值
var allowedHybridPolicies = []string{"default", "always_remote", "always_local"}

// 检查输入的 hybrid_policy 是否合法
func isValidHybridPolicy(policy string) bool {
	for _, allowed := range allowedHybridPolicies {
		if allowed == policy {
			return true
		}
	}
	return false
}

// NewInstallServiceCommand will install a service
func NewInstallServiceCommand() *cobra.Command {
	var (
		providerName string
		remoteFlag   bool
		remoteURL    string
		authType     string
		method       string
		authKey      string
		flavor       string
	)

	installServiceCmd := &cobra.Command{
		Use:    "install <service>",
		Short:  "Install a service or service provider",
		Long:   `Install a service by name or a service provider from a file.`,
		Args:   cobra.ExactArgs(1),
		PreRun: CheckAOGServer,
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

	return installServiceCmd
}

// NewVersionCommand print client version
func NewVersionCommand() *cobra.Command {
	ver := &cobra.Command{
		Use:   "version",
		Short: "Prints aog build version information.",
		Long:  "Prints aog build version information.",
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
	cmd := &cobra.Command{
		Use:   "start",
		Short: "aog apiserver is a aipc open gateway",
		Long:  "aog apiserver is a aipc open gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			isDaemon, err := cmd.Flags().GetBool("d")
			if err != nil {
				return err
			}
			if isDaemon {
				CheckAOGServer(cmd, args)
				return nil
			}
			return Run(context.Background())
		},
	}
	fs := cmd.Flags()
	namedFlagSets := config.GlobalAOGEnvironment.Flags()
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
		Use:   "models",
		Short: "List models for a specific service",
		Long:  `List models for a specific service.`,
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
			// todo 其实获取到Flavor后，authType method url都已经是固定了，需考虑这部分是否需要提供给用户设定
			// 代码内部需要维护一套关于Flavor的相关信息（authType method url），建议在template/flavor.yaml中设定
			// remoteURL, _ := cmd.Flags().GetString("remote_url")
			// authType, _ := cmd.Flags().GetString("auth_type")
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
		routerPath := fmt.Sprintf("/aog/%s/service", version.AOGVersion)

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

		fmt.Println("Service chat install success!")

		if !remote {
			askRes := askEnableRemoteService()
			if askRes {
				fmt.Println("请前往 https://platform.deepseek.com/ 网址申请 APIKEY。")
				apiKey := getAPIKey()
				if apiKey != "" {
					fmt.Printf("\r你输入的 API Key 是: %s\n", apiKey)
				}

				// todo 暂时默认远程参数 后期需动态获取
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
	if isServerRunning() {
		return
	}

	fmt.Println("AOG server is not running. Starting the server...")
	if err := startAogServer(); err != nil {
		log.Fatalf("Failed to start AOG server: %s \n", err.Error())
		return
	}

	time.Sleep(6 * time.Second)

	if !isServerRunning() {
		log.Fatal("Failed to start AOG server.")
		return
	}

	fmt.Println("AOG server started successfully.")
}

func isServerRunning() bool {
	serverUrl := "http://127.0.0.1:16688/aog/" + version.AOGVersion + "/health"
	resp, err := http.Get(serverUrl)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func startAogServer() error {
	// exePath := filepath.Join(config.GlobalAOGEnvironment.WorkDir, "aog.exe")
	logPath := config.GlobalAOGEnvironment.ConsoleLog

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer logFile.Close()

	cmd := exec.Command("aog.exe", "server", "start")
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start AOG server: %v", err)
	}

	// 保存 PID 到文件
	pid := cmd.Process.Pid
	pidFile := filepath.Join(config.GlobalAOGEnvironment.RootDir, "aog.pid")
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0o644); err != nil {
		return fmt.Errorf("failed to save PID to file: %v", err)
	}

	fmt.Printf("\rAOG server started with PID: %d\n", cmd.Process.Pid)
	return nil
}

// 获取用户输入的 API Key
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

// 询问用户是否启用远程服务
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

	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	wg.Add(1)
	msg := "Pulling model"
	go progress.ShowLoadingAnimation(stopChan, &wg, msg)

	c := config.NewAOGClient()
	routerPath := fmt.Sprintf("/aog/%s/model", version.AOGVersion)

	err = c.Client.Do(context.Background(), http.MethodPost, routerPath, req, &resp)
	if err != nil {
		fmt.Printf("\rPull model failed: %s", err.Error())
		return
	}

	close(stopChan)
	wg.Wait()

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
		Use:   "import <file_path>",
		Short: "Import service configuration from a file",
		Long:  "Import service configuration from a file and send it to the API.",
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
				return fmt.Errorf("failed to parse file content: %w", err)
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
