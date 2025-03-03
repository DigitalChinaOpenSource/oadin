// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/fatih/color"

	"intel.com/opac/opac"
)

const apiLayerVersion = "v0.1"
const specVersion = "v0.1"

func main() {

	// Flag processing
	isDebug := flag.Bool("vv", false, "More verbose level. Show show much more Debug level internal")
	isInfo := flag.Bool("v", false, "Verbose mode. Show general Info level log messages")
	gatewayPort := flag.Int("port", 16688, "Port for gateway to listen on")
	configFile := flag.String("config", "./opac.json", "Gateway configuration file")
	logHTTP := flag.String("log_http", "", "Path to a log file which will"+
		"log http requests / response. This is to help debug")
	forceReload := flag.Bool("force_reload", false, "Debug purpose only and may cause BUG!"+
		" It forces reload many configurations such as conversions in flavors, and most "+
		" sections of opac.json. This makes debug much easier because the changes to "+
		" configuration files take effect on the fly without the need of restart the gateway."+
		" but it may result in inconsistent internal state before and after the change so "+
		" it should be DISABLED in production environment.")

	flag.Parse()

	// setup log system
	var verbose string
	opts := slogcolor.DefaultOptions
	if *isDebug {
		opts.Level = slog.LevelDebug
		verbose = "debug"
	} else if *isInfo {
		opts.Level = slog.LevelInfo
		verbose = "info"
	} else {
		opts.Level = slog.LevelWarn
		verbose = "warn"
	}
	opts.SrcFileMode = slogcolor.Nop
	opts.MsgColor = color.New(color.FgHiYellow)

	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, opts)))
	color.New(color.FgHiCyan).Println(">>>>>> AIPC Open Gateway Starting : " + time.Now().Format("2006-01-02 15:04:05") + "\n\n")
	defer func() {
		color.New(color.FgHiGreen).Println("\n\n<<<<<< AIPC Open Gateway Stopped : " + time.Now().Format("2006-01-02 15:04:05"))
	}()

	// setup environment
	env := opac.GetEnv()
	env.APILayerVersion = apiLayerVersion
	env.SpecVersion = specVersion
	env.Verbose = verbose
	env.ForceReload = *forceReload
	env.ConfigFile = env.GetAbsolutePath(*configFile, env.WorkDir)
	slog.Info("[Init] OPAC Environment", "env", fmt.Sprintf("%+v", env))

	// enable event with file log
	opac.InitSysEvents(*logHTTP)
	opac.SysEvents.Notify("start_app", nil)

	err := opac.LoadPlatformInfo(env.ConfigFile)
	if err != nil {
		slog.Error("[Init] Failed to load opac.json", "error", err)
		return
	}

	// load all flavors
	// this loads all config based API Flavors. You need to manually
	// create and RegisterAPIFlavor for costimized API Flavors
	err = opac.InitAPIFlavors()
	if err != nil {
		slog.Error("[Init] Failed to load API Flavors", "error", err)
		return
	}

	// start
	opac.StartScheduler("basic")
	color.New(color.FgHiGreen).Println("AOG Gateway starting on port", *gatewayPort)
	opac.StartGateway(*gatewayPort, *isDebug)

	// Create a channel to receive signals
	sigChan := make(chan os.Signal, 1)
	// Notify sigChan on SIGINT (Ctrl+C)
	signal.Notify(sigChan, syscall.SIGINT)

	sig := <-sigChan // Block until a signal is received
	slog.Info("main: Received SIGINT signal to close gateway ...", "signal", sig)
}
