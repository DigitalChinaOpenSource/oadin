package config

import (
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"oadin/internal/utils"
	"oadin/internal/utils/client"
	"oadin/version"

	"github.com/MatusOllah/slogcolor"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

var GlobalOadinEnvironment *OadinEnvironment

type OadinEnvironment struct {
	ApiHost           string // host
	Datastore         string // path to the datastore
	DatastoreType     string // type of the datastore
	Verbose           string // debug, info or warn
	RootDir           string // root directory for all assets such as config files
	WorkDir           string // current work directory
	APIVersion        string // version of this core app layer (gateway etc.)
	SpecVersion       string // version of the core specification this app layer supports
	UpdateDir         string // Installation package storage path
	LogDir            string // logs dir
	LogHTTP           string // path to the http log
	LogLevel          string // log level
	LogFileExpireDays int    // log file expiration time
	ConsoleLog        string // oadin server console log path
}

var (
	once         sync.Once
	envSingleton *OadinEnvironment
)

type OadinClient struct {
	client.Client
}

func NewOadinClient() *OadinClient {
	return &OadinClient{
		Client: *client.NewClient(Host(), http.DefaultClient),
	}
}

// Host returns the scheme and host. Host can be configured via the Oadin_HOST environment variable.
// Default is scheme "http" and host "127.0.0.1:16688"
func Host() *url.URL {
	defaultPort := "16688"

	s := strings.TrimSpace(Var("Oadin_HOST"))
	scheme, hostport, ok := strings.Cut(s, "://")
	switch {
	case !ok:
		scheme, hostport = "http", s
	case scheme == "http":
		defaultPort = "80"
	case scheme == "https":
		defaultPort = "443"
	}

	hostport, path, _ := strings.Cut(hostport, "/")
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		host, port = "127.0.0.1", defaultPort
		if ip := net.ParseIP(strings.Trim(hostport, "[]")); ip != nil {
			host = ip.String()
		} else if hostport != "" {
			host = hostport
		}
	}

	if n, err := strconv.ParseInt(port, 10, 32); err != nil || n > 65535 || n < 0 {
		slog.Warn("invalid port, using default", "port", port, "default", defaultPort)
		port = defaultPort
	}

	return &url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, port),
		Path:   path,
	}
}

// Var returns an environment variable stripped of leading and trailing quotes or spaces
func Var(key string) string {
	return strings.Trim(strings.TrimSpace(os.Getenv(key)), "\"'")
}

func NewOadinEnvironment() *OadinEnvironment {
	once.Do(func() {
		env := OadinEnvironment{
			ApiHost:           "0.0.0.0:16688",
			Datastore:         "oadin.db",
			DatastoreType:     "sqlite",
			UpdateDir:         "updates",
			LogDir:            "logs",
			LogHTTP:           "server.log",
			LogLevel:          "DEBUG",
			LogFileExpireDays: 7,
			Verbose:           "info",
			RootDir:           "./",
			WorkDir:           "./",
			APIVersion:        version.OadinVersion,
			SpecVersion:       version.OadinVersion,
			ConsoleLog:        "console.log",
		}
		cwd, err := os.Getwd()
		if err != nil {
			panic("[GetEnv] Failed to get current working directory")
		}
		env.WorkDir = cwd

		env.RootDir, err = utils.GetOadinDataDir()
		if err != nil {
			panic("[Init Env] get user dir failed: " + err.Error())
		}
		env.Datastore = filepath.Join(env.RootDir, env.Datastore)
		env.UpdateDir = filepath.Join(env.RootDir, env.UpdateDir)
		env.LogDir = filepath.Join(env.RootDir, env.LogDir)
		env.LogHTTP = filepath.Join(env.LogDir, env.LogHTTP)
		env.ConsoleLog = filepath.Join(env.LogDir, env.ConsoleLog)

		if err := os.MkdirAll(env.LogDir, 0o750); err != nil {
			panic("[Init Env] create logs path : " + err.Error())
		}

		envSingleton = &env
	})
	return envSingleton
}

// FlagSets Define a struct to hold the flag sets and their order
type FlagSets struct {
	Order    []string
	FlagSets map[string]*pflag.FlagSet
}

// NewFlagSets Initialize the FlagSets struct
func NewFlagSets() *FlagSets {
	return &FlagSets{
		Order:    []string{},
		FlagSets: make(map[string]*pflag.FlagSet),
	}
}

// AddFlagSet Add a flag set to the struct and maintain the order
func (fs *FlagSets) AddFlagSet(name string, flagSet *pflag.FlagSet) {
	if _, exists := fs.FlagSets[name]; !exists {
		fs.Order = append(fs.Order, name)
	}
	fs.FlagSets[name] = flagSet
}

// GetFlagSet Get the flag set by name, creating it if it doesn't exist
func (fs *FlagSets) GetFlagSet(name string) *pflag.FlagSet {
	if _, exists := fs.FlagSets[name]; !exists {
		fs.FlagSets[name] = pflag.NewFlagSet(name, pflag.ExitOnError)
		fs.Order = append(fs.Order, name)
	}
	return fs.FlagSets[name]
}

// Flags returns the flag sets for the OadinEnvironment.
func (s *OadinEnvironment) Flags() *FlagSets {
	fss := NewFlagSets()
	fs := fss.GetFlagSet("generic")
	fs.StringVar(&s.ApiHost, "app-host", s.ApiHost, "API host")
	fs.StringVar(&s.Datastore, "datastore", s.Datastore, "Datastore path")
	fs.StringVar(&s.DatastoreType, "datastore-type", s.DatastoreType, "Datastore type")
	fs.StringVar(&s.LogHTTP, "log-http", s.LogHTTP, "HTTP log path")
	fs.StringVar(&s.Verbose, "verbose", s.Verbose, "Log verbosity level")
	fs.StringVar(&s.RootDir, "root-dir", s.RootDir, "Root directory")
	fs.StringVar(&s.WorkDir, "work-dir", s.WorkDir, "Work directory")
	fs.StringVar(&s.APIVersion, "app-layer-version", s.APIVersion, "API layer version")
	fs.StringVar(&s.SpecVersion, "spec-version", s.SpecVersion, "Specification version")
	return fss
}

func (s *OadinEnvironment) SetSlogColor() {
	opts := slogcolor.DefaultOptions
	if s.Verbose == "debug" {
		opts.Level = slog.LevelDebug
	} else if s.Verbose == "warn" {
		opts.Level = slog.LevelWarn
	} else {
		opts.Level = slog.LevelInfo
	}
	opts.SrcFileMode = slogcolor.Nop
	opts.MsgColor = color.New(color.FgHiYellow)

	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, opts)))
	_, _ = color.New(color.FgHiCyan).Println(">>>>>> Oadin Open Gateway Starting : " + time.Now().Format("2006-01-02 15:04:05") + "\n\n")
	defer func() {
		_, _ = color.New(color.FgHiGreen).Println("\n\n<<<<<< Oadin Open Gateway Stopped : " + time.Now().Format("2006-01-02 15:04:05"))
	}()
}
