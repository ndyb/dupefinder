package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
)

type fileAction func(file []*File) error

func faPrint(file []*File) error {
	fmt.Println(file[0].path)
	return nil
}

func faVerbose(file []*File) error {
	fmt.Println(file[0].path)
	return nil
}

func faDontask(file []*File) error {
	fmt.Println(file[0].path)
	return nil
}

func faDelete(file []*File) error {
	fmt.Println(file[0].path)
	return nil
}

type ZeroWriter struct{}

func (zw *ZeroWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func main() {

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var (
		cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
		memprofile = flag.String("memprofile", "", "write memory profile to file")
		path       = flag.String("p", pwd, "path to start the search")
		action     = flag.String("action", "print", "action to take: print, delete, verbose")
		size       = flag.Int64("s", 1024*1024, "minimum file size to check in bytes")
		help       = flag.Bool("help", false, "show help")
		ext        = flag.Bool("e", true, "don't compare files with different extensions")
		verbose    = flag.Bool("v", false, "verbose output")
	)

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if !*verbose {
		writer := &ZeroWriter{}
		log.SetOutput(writer)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		defer f.Close()
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		defer f.Close()
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
	}

	var actionFunction fileAction

	switch *action {
	case "print":
		actionFunction = faPrint
	case "delete":
		actionFunction = faDelete
	case "verbose":
		actionFunction = faVerbose
	case "dontask":
		actionFunction = faDontask
	default:
		flag.Usage()
		os.Exit(0)
	}

	type Files map[Hash][]File

	files := make(Files)
	queue := make(chan []*File, 1000)

	go func() {
		filepath.Walk(*path, func(path string, info os.FileInfo, err error) error {

			file := *NewFile(path, info)

			if file.IsDir() {
				log.Printf("Walking into %s\n", file.path)
				return nil
			}
			if file.FileSize() < *size {
				return nil
			}

			h := file.Hash(*ext)

			_, e := files[h]
			if e {
				for i := range files[h] {
					if files[h][i].Equal(&file) {
						log.Printf("%s == %s\n", files[h][i].path, file.path)
						queue <- []*File{&file, &files[h][i]}
						return nil
					}
				}
			}

			files[h] = append(files[h], file)

			return nil
		})

		close(queue)

	}()

	for {
		f, ok := <-queue
		if !ok {
			os.Exit(0)
		}
		actionFunction(f)
	}
}
