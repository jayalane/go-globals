// -*- tab-width: 2 -*-

package globals

import (
	"fmt"
	count "github.com/jayalane/go-counter"
	lll "github.com/jayalane/go-lll"
	config "github.com/jayalane/go-tinyconfig"
	"github.com/pkg/profile"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"unsafe"
)

type Global struct {
	Ml            lll.Lll
	Cfg           *config.Config
	defaultConfig string
	reload        chan os.Signal // to reload config
}

func (g *Global) reloadHandler() {
	for c := range g.reload {
		g.Ml.La("OK: Got a signal, reloading config", c)

		t, err := config.ReadConfig("config.txt", g.defaultConfig)
		if err != nil {
			fmt.Println("Error opening config.txt", err.Error())
			return
		}
		st := unsafe.Pointer(g.Cfg)
		atomic.StorePointer(&st, unsafe.Pointer(&t))
		fmt.Println("New Config", (*g.Cfg)) // lll isn't up yet
		lll.SetLevel(&g.Ml, (*g.Cfg)["debugLevel"].StrVal)
	}
}

func NewGlobal(defaultConfig string) Global {

	res := Global{}

	// CPU profile
	defer profile.Start(profile.ProfilePath(".")).Stop()

	// config
	if len(os.Args) > 1 && os.Args[1] == "--dumpConfig" {
		fmt.Println(defaultConfig)
		os.Exit(0)
	}
	// still config
	res.Cfg = nil
	t, err := config.ReadConfig("config.txt", defaultConfig)
	if err != nil {
		fmt.Println("Error opening config.txt", err.Error())
		if res.Cfg == nil {
			os.Exit(11)
		}
	}
	res.Cfg = &t
	fmt.Println("Config", (*res.Cfg)) // lll isn't up yet

	// start the profiler
	go func() {
		if len((*res.Cfg)["profListen"].StrVal) > 0 {
			fmt.Println(http.ListenAndServe((*res.Cfg)["profListen"].StrVal, nil))
		}
	}()

	// low level logging (first so everything rotates)
	res.Ml = lll.Init("MAIN", (*res.Cfg)["debugLevel"].StrVal)

	// config sig handlers - to enable log levels
	res.reload = make(chan os.Signal, 2)
	signal.Notify(res.reload, syscall.SIGHUP)
	go res.reloadHandler() // to listen to the signal

	// stats
	count.InitCounters()

	return res
}
