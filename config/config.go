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

	"oadin/internal/client"
	"oadin/internal/constants"
	"oadin/internal/types"
	"oadin/internal/utils"
	"oadin/version"

	"github.com/MatusOllah/slogcolor"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

const (
	// Log levels
	LogLevelDebug = "debug"
	LogLevelWarn  = "warn"
	LogLevelError = "error"

	// Default configurations
	DefaultLogLevel = "debug"
	DefaultVerbose  = "info"
	DefaultRootDir  = "./"

	// Database types
	DatastoreSQLite = "sqlite"

	// Database file
	DefaultDatabaseFile = "oadin.db"

	// Directory names
	LogsDirectory = "logs"

	// File names
	ServerLogFile  = "server.log"
	ConsoleLogFile = "console.log"

	// Time formats
	DefaultTimeFormat = "2006-01-02 15:04:05"

	// Log file expiration in days
	DefaultLogExpireDays = 7

	// Environment variable keys
	EnvOADINHost              = "OADIN_HOST"
	EnvModelIdleTimeout       = "OADIN_MODEL_IDLE_TIMEOUT"
	EnvModelCleanupInterval   = "OADIN_MODEL_CLEANUP_INTERVAL"
	EnvLocalModelQueueSize    = "OADIN_LOCAL_MODEL_QUEUE_SIZE"
	EnvLocalModelQueueTimeout = "OADIN_LOCAL_MODEL_QUEUE_TIMEOUT"
)

var GlobalEnvironment *OADINEnvironment

type OADINEnvironment struct {
	ApiHost                string        // host
	Datastore              string        // path to the datastore
	DatastoreType          string        // type of the datastore
	Verbose                string        // debug, info or warn
	RootDir                string        // root directory for all assets such as config files
	APIVersion             string        // version of this core app layer (gateway etc.)
	SpecVersion            string        // version of the core specification this app layer supports
	LogDir                 string        // logs dir
	LogHTTP                string        // path to the http log
	LogLevel               string        // log level
	LogFileExpireDays      int           // log file expiration time
	ConsoleLog             string        // oadin server console log path
	ModelIdleTimeout       time.Duration // model idle timeout duration
	ModelCleanupInterval   time.Duration // model cleanup check interval
	LocalModelQueueSize    int           // local model queue size
	LocalModelQueueTimeout time.Duration // local model queue timeout

	UpdateDir string // Installation package storage path
}

var (
	once         sync.Once
	envSingleton *OADINEnvironment
)

type OADINClient struct {
	client.Client
}

func NewOADINClient() *OADINClient {
	return &OADINClient{
		Client: *client.NewClient(Host(), http.DefaultClient),
	}
}

// Host returns the scheme and host. Host can be configured via the OADIN_HOST environment variable.
// Default is scheme host and host "127.0.0.1:16688"
func Host() *url.URL {
	defaultPort := constants.DefaultHTTPPort

	s := strings.TrimSpace(Var(EnvOADINHost))
	scheme, hostport, ok := strings.Cut(s, "://")
	switch {
	case !ok:
		scheme, hostport = types.ProtocolHTTP, s
	case scheme == types.ProtocolHTTP:
		defaultPort = constants.DefaultHTTPPort80
	case scheme == types.ProtocolHTTPS:
		defaultPort = constants.DefaultHTTPSPort
	}

	hostport, path, _ := strings.Cut(hostport, "/")
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		// host, port = "127.0.0.1", defaultPort
		host, port = constants.DefaultHost, defaultPort
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

func NewOADINEnvironment() *OADINEnvironment {
	once.Do(func() {
		env := OADINEnvironment{
			ApiHost:                constants.DefaultHost + ":" + constants.DefaultHTTPPort,
			Datastore:              DefaultDatabaseFile,
			DatastoreType:          DatastoreSQLite,
			LogDir:                 LogsDirectory,
			LogHTTP:                ServerLogFile,
			LogLevel:               DefaultLogLevel,
			LogFileExpireDays:      DefaultLogExpireDays,
			Verbose:                DefaultVerbose,
			RootDir:                DefaultRootDir,
			APIVersion:             version.OADINVersion,
			SpecVersion:            version.OADINVersion,
			ConsoleLog:             ConsoleLogFile,
			ModelIdleTimeout:       5 * time.Minute,  // 默认5分钟
			ModelCleanupInterval:   1 * time.Minute,  // 默认1分钟
			LocalModelQueueSize:    10,               // 默认队列大小10
			LocalModelQueueTimeout: 30 * time.Second, // 默认排队超时30秒
		}

		var err error
		env.RootDir, err = utils.GetOADINDataDir()
		if err != nil {
			panic("[Init Env] get user dir failed: " + err.Error())
		}
		env.Datastore = filepath.Join(env.RootDir, env.Datastore)
		env.LogDir = filepath.Join(env.RootDir, env.LogDir)
		env.LogHTTP = filepath.Join(env.LogDir, env.LogHTTP)
		env.ConsoleLog = filepath.Join(env.LogDir, env.ConsoleLog)

		env.UpdateDir = filepath.Join(env.RootDir, env.UpdateDir)

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

// Flags returns the flag sets for the OADINEnvironment.
func (s *OADINEnvironment) Flags() *FlagSets {
	fss := NewFlagSets()
	fs := fss.GetFlagSet("generic")
	fs.StringVar(&s.ApiHost, "app-host", s.ApiHost, "API host")
	fs.StringVar(&s.Verbose, "verbose", s.Verbose, "Log verbosity level")
	return fss
}

func (s *OADINEnvironment) SetSlogColor() {
	opts := slogcolor.DefaultOptions
	if s.Verbose == LogLevelDebug {
		opts.Level = slog.LevelDebug
	} else if s.Verbose == LogLevelWarn {
		opts.Level = slog.LevelWarn
	} else {
		opts.Level = slog.LevelInfo
	}
	opts.SrcFileMode = slogcolor.Nop
	opts.MsgColor = color.New(color.FgHiYellow)

	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, opts)))
	_, _ = color.New(color.FgHiCyan).Println(">>>>>> OADIN Open Gateway Starting : " + time.Now().Format(DefaultTimeFormat) + "\n\n")
	defer func() {
		_, _ = color.New(color.FgHiGreen).Println("\n\n<<<<<< OADIN Open Gateway Stopped : " + time.Now().Format(DefaultTimeFormat))
	}()
}
