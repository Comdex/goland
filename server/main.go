package main

import (
	"flag"
	//"github.com/aarzilli/golua/lua"
	"github.com/mischief/goland/game/gutil"
	lua "github.com/xenith-studios/golua"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

var (
	configfile = flag.String("config", "config.lua", "configuration file")

	Lua *lua.State
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	Lua = lua.NewState()

	if Lua == nil {
		log.Fatal("Can't make lua state")
	}

	Lua.OpenLibs()
}

func main() {
	flag.Parse()

	// load configuration
	ParMap, err := gutil.LuaParMapFromFile(Lua, *configfile)
	if err != nil || ParMap == nil {
		log.Fatalf("Error loading configuration file %s: %s", *configfile, err)
	}

	// setup logging
	lf, ok := ParMap.Get("logfile")
	if !ok {
		log.Printf("No logfile specified, using stdout")
	} else {
		// open log file
		f, err := os.OpenFile(lf, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Print("-- Logging started --")

	// log panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from %v", r)
		}
	}()

	log.Printf("Config loaded from %s", *configfile)

	// dump config
	it := ParMap.Iter()
	for k, v, b := it(); b != false; k, v, b = it() {
		log.Printf(" %s -> %s", k, v)
	}

	// enable profiling
	if cpuprofile, ok := ParMap.Get("cpuprofile"); ok {
		log.Printf("Starting profiling in file %s", cpuprofile)
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	gs := NewGameServer(ParMap)

	gs.Run()

	log.Println("-- Logging ended --")
}
