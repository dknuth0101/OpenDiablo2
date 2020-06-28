package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/profile"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2engine"
	// "github.com/OpenDiablo2/OpenDiablo2/d2common"
	// "github.com/OpenDiablo2/OpenDiablo2/d2common/d2data"
	// "github.com/OpenDiablo2/OpenDiablo2/d2common/d2data/d2datadict"
	// "github.com/OpenDiablo2/OpenDiablo2/d2common/d2resource"
	// "github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"
	// "github.com/OpenDiablo2/OpenDiablo2/d2core/d2audio"
	// ebiten2 "github.com/OpenDiablo2/OpenDiablo2/d2core/d2audio/ebiten"
	// "github.com/OpenDiablo2/OpenDiablo2/d2core/d2config"
	// "github.com/OpenDiablo2/OpenDiablo2/d2core/d2gui"
	// "github.com/OpenDiablo2/OpenDiablo2/d2core/d2input"
	// ebiten_input "github.com/OpenDiablo2/OpenDiablo2/d2core/d2input/ebiten"
	// "github.com/OpenDiablo2/OpenDiablo2/d2core/d2interface"
	// "github.com/OpenDiablo2/OpenDiablo2/d2core/d2inventory"
	// "github.com/OpenDiablo2/OpenDiablo2/d2core/d2render"
	// "github.com/OpenDiablo2/OpenDiablo2/d2core/d2render/ebiten"
	// "github.com/OpenDiablo2/OpenDiablo2/d2core/d2screen"
	// "github.com/OpenDiablo2/OpenDiablo2/d2core/d2term"
	// "github.com/OpenDiablo2/OpenDiablo2/d2core/d2ui"
	// "github.com/OpenDiablo2/OpenDiablo2/d2game/d2gamescreen"
	// "github.com/OpenDiablo2/OpenDiablo2/d2script"
)

// GitBranch is set by the CI build process to the name of the branch
var GitBranch string

// GitCommit is set by the CI build process to the commit hash
var GitCommit string

// Version is set by humans, we should update this on git tags or
// major/minor revisions that break compatability
const Version string = "0.1" // update this on breaking changes

func main() {

	region := kingpin.Arg("region", "Region type id").Int()
	preset := kingpin.Arg("preset", "Level preset").Int()
	profileOption := kingpin.Flag("profile", "Profiles the program, one of (cpu, mem, block, goroutine, trace, thread, mutex)").String()

	kingpin.Parse()

	log.SetFlags(log.Lshortfile)
	log.Println("OpenDiablo2 - Open source Diablo 2 engine")

	if od2, err := d2engine.Create(Version, GitBranch, GitCommit); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(*profileOption) > 0 {
		profiler := enableProfiler(*profileOption)
		if profiler != nil {
			defer profiler.Stop()
		}
	}

	if *region == 0 {
		od2.CreateMainMenu()
	} else {
		od2.CreateMapEngineTest(*region, *preset)
	}

	od2.Run()
}

// enableProfiler enables the program profiler
func enableProfiler(profileOption string) interface{ Stop() } {
	var options []func(*profile.Profile)
	switch strings.ToLower(strings.Trim(profileOption, " ")) {
	case "cpu":
		log.Printf("CPU profiling is enabled.")
		options = append(options, profile.CPUProfile)
	case "mem":
		log.Printf("Memory profiling is enabled.")
		options = append(options, profile.MemProfile)
	case "block":
		log.Printf("Block profiling is enabled.")
		options = append(options, profile.BlockProfile)
	case "goroutine":
		log.Printf("Goroutine profiling is enabled.")
		options = append(options, profile.GoroutineProfile)
	case "trace":
		log.Printf("Trace profiling is enabled.")
		options = append(options, profile.TraceProfile)
	case "thread":
		log.Printf("Thread creation profiling is enabled.")
		options = append(options, profile.ThreadcreationProfile)
	case "mutex":
		log.Printf("Mutex profiling is enabled.")
		options = append(options, profile.MutexProfile)
	}
	options = append(options, profile.ProfilePath("./pprof/"))

	if len(options) > 1 {
		return profile.Start(options...)
	}
	return nil
}
