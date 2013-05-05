package main

import (
	"fmt"
	"hash/crc32"
	"io/ioutil"

	"os"
	"path/filepath"
)

var START = "y:"
var LIMIT int64 = 1 * 1024 * 1024 * 50

var files = make(map[int64][]File)
var queue = make(chan string, 100)

type File struct {
	path string
	hash uint32
}

func main() {

	go filepath.Walk(START, do)

	for {
		f := <-queue
		fmt.Println("Delete " + f + "?")
		var a string
		fmt.Scanln(&a)
		if a == "yes" {
			os.Remove(f)
		}
	}
}

func do(path string, info os.FileInfo, err error) error {

	if info.IsDir() || info.Size() < LIMIT {
		return nil
	}

	_, e := files[info.Size()]
	if !e {
		files[info.Size()] = append(files[info.Size()], File{path, 0})
		return nil
	}

	h, _ := hash(path)
	f := File{path, h}
	for _, i := range files[info.Size()] {
		if i.hash == 0 {
			i.hash, _ = hash(i.path)
		}

		if i.hash == f.hash {
			delete(f.path)
			return nil
		}
	}

	files[info.Size()] = append(files[info.Size()], f)

	return nil

}

func delete(filename string) {
	queue <- filename

}

func hash(filename string) (uint32, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	h := crc32.NewIEEE()
	h.Write(bs)
	return h.Sum32(), nil
}
