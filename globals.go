// -*- tab-width: 2 -*-

package globals

import (
	"fmt"
	count "github.com/jayalane/go-counter"
	lll "github.com/jayalane/go-lll"
	config "github.com/jayalane/go-tinyconfig"
	"github.com/pkg/profile"
	"net/http"
	_ "net/http/pprof" // for pprof
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"unsafe"
)

// Global has a logger Ml, config Cfg, and handles reloads of config
type Global struct {
	Ml            lll.Lll
	Cfg           *config.Config
	defaultConfig string
	reload        chan os.Signal // to reload config
}

func (g *Global) reloadHandler() {
	for c := range g.reload {
		g.Ml.La("OK: Got a signal, reloading config", c)
		oldCfg := g.Cfg

		t, err := config.ReadConfig("config.txt", g.defaultConfig)
		if err != nil {
			fmt.Println("Error opening config.txt", err.Error())
			return
		}
		st := unsafe.Pointer(g.Cfg)
		atomic.StorePointer(&st, unsafe.Pointer(&t))
		fmt.Println("Old cfg", oldCfg, "new cfg", g.Cfg, "st", st)
		fmt.Println("New Config", (*g.Cfg)) // lll isn't up yet
		lll.SetLevel(&g.Ml, (*g.Cfg)["debugLevel"].StrVal)
	}
}

// NewGlobal sets up the logger, the profiler, if doProf is true, and
// reads the config
func NewGlobal(defaultConfig string, doProf bool) Global {

	res := Global{}

	// CPU profile
	if doProf {
		defer profile.Start(profile.ProfilePath(".")).Stop()
	}
	defaultProfile := "localhost:8888"
	defaultLogLevel := "all"
	// config
	if len(defaultConfig) > 0 {
		if len(os.Args) > 1 && os.Args[1] == "--dumpConfig" {
			fmt.Println("logStdout = false\n" + defaultConfig)
			os.Exit(0)
		}
		// still config
		res.Cfg = nil
		t, err := config.ReadConfig("config.txt", "logStdout = false\n"+defaultConfig)
		if err != nil {
			fmt.Println("Error opening config.txt", err.Error())
			if res.Cfg == nil {
				os.Exit(11)
			}
		}
		res.Cfg = &t
		fmt.Println("Config", (*res.Cfg)) // lll isn't up yet

		defaultProfile = (*res.Cfg)["profListen"].StrVal
		defaultLogLevel = (*res.Cfg)["debugLevel"].StrVal

		// config sig handlers - to enable log levels
		res.reload = make(chan os.Signal, 2)
		signal.Notify(res.reload, syscall.SIGHUP)
		go res.reloadHandler() // to listen to the signal

	}
	// start the profiler
	go func() {
		if len(defaultProfile) > 0 {
			fmt.Println(http.ListenAndServe(defaultProfile, nil))
		}
	}()
	// low level logging (first so everything rotates)
	if res.Cfg != nil && (*res.Cfg)["logStdout"].BoolVal {
		lll.SetWriter(os.Stdout)
	}
	res.Ml = lll.Init("MAIN", defaultLogLevel)

	// stats
	count.InitCounters()

	return res
}

// NewLogger returns a new logger that is from github.com/jayalane/go-lll
func (g Global) NewLogger(name string, defaultLogLevel string) *lll.Lll {
	ml := lll.Init(name, defaultLogLevel)
	return &ml
}
