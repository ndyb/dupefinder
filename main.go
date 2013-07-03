package main

/*
go-dupfind finds duplicate files in a folder.

Usage:
	go-dupfind [flags]

The flags are:
	-v
		verbose mode
	-s
		minimum file size to check
	-a
		action to take: print, delete, verbose
	-p
		path to start the search. Starts from present working directory if not
		set.
	-help
		show help
	-e=true
		don't compare files with different extensions
*/

import (
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	. "github.com/ndyb/go-dupfind/dupefinder"
	"github.com/ndyb/utils/debug"
	"github.com/ndyb/utils/profiling"
	"log"
	"os"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to file")
	action     = flag.String("a", "print", "action to take: print, delete, verbose")
	size       = flag.String("s", "0", "minimum file size to check")
	path       = flag.String("p", "", "path to start the search")
	help       = flag.Bool("help", false, "show help")
	ext        = flag.Bool("e", true, "don't compare files with different extensions")
	verbose    = flag.Bool("v", false, "verbose mode")
)

var (
	dbg        debug.Debug
	minSize    uint64
	fileaction FileAction
	err        error
)

func init() {
	flag.Parse()

	if *cpuprofile != "" {
		profiling.CpuProfile(*cpuprofile)
	}

	if *memprofile != "" {
		profiling.MemProfile(*memprofile)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *verbose {
		dbg = true
	}

	minSize, err = humanize.ParseBytes(*size)
	if err != nil {
		minSize = 0
	}

	fileaction, err = GetActionFor(*action)
	if err != nil {
		log.Fatal("Can't find action")
	}

	if *path == "" {
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		*path = pwd
	}
}

func main() {

	var queue = make(chan []*File, 1000)
	var stats struct {
		files int
		size  uint64
	}

	go FindDuplicates(*path, minSize, *ext, queue)

	for {
		f, ok := <-queue

		if !ok {
			fmt.Printf("\nFound %d duplicate(s), totaling %s bytes.\n", stats.files, humanize.Bytes(stats.size))
			os.Exit(0)
		}

		stats.files = stats.files + 1
		stats.size = stats.size + uint64(f[0].FileSize())

		fileaction(f)
	}
}
