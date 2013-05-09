package dupefinder

import (
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/ndyb/utils/debug"
	"github.com/ndyb/utils/profiling"
	"log"
	"os"
	"path/filepath"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to file")
	action     = flag.String("action", "print", "action to take: print, delete, verbose")
	size       = flag.String("s", "0", "minimum file size to check in bytes")
	path       = flag.String("p", "", "path to start the search")
	help       = flag.Bool("help", false, "show help")
	ext        = flag.Bool("e", true, "don't compare files with different extensions")
	verbose    = flag.Bool("v", false, "debug output")
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

	fileaction, err = getActionFor(*action)
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

var (
	files = make(Files)
	queue = make(chan []*File, 1000)
	stats struct {
		files int
		size  uint64
	}
)

// Find duplicates starting from path and send list of duplicate files to channel output
func FindDuplicates(path string, output *chan ([]*File)) {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {

		file := *NewFile(path, info)

		if !file.IsRegular() {
			dbg.Printf("Not regular file %s. Ignoring!\n", file.path)
			return filepath.SkipDir
		}

		if file.IsDir() {
			dbg.Printf("Walking into %s\n", file.path)
			return nil
		}

		if file.FileSize() < minSize {
			return nil
		}

		h := file.Hash(*ext)

		_, exists := files[h]
		if exists {
			for i := range files[h] {
				equals, err := files[h][i].Equal(&file)
				if err != nil {
					dbg.Printf("Error accessing %s\n", file.path)
					return nil
				}
				if equals {
					dbg.Printf("%s == %s\n", file.path, files[h][i].path)
					queue <- []*File{&file, &files[h][i]}
					return nil
				}
			}
		}

		files[h] = append(files[h], file)

		return nil
	})

	close(queue)
}

// test
func main() {

	go FindDuplicates(*path, &queue)

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
