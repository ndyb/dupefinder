package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	path  string
	info  os.FileInfo
	crc   uint32
	intro []byte
}

type Hash struct {
	size int64
	ext  string
}

type FileAction func(file []*File) error

func getActionFor(s string) (FileAction, error) {
	switch s {
	case "print":
		return func(file []*File) error {
			fmt.Printf("%s\t%s\n", humanize.Bytes(uint64(file[0].FileSize())), file[0].path)
			return nil
		}, nil
	case "delete":
		return func(file []*File) error {
			fmt.Println(file[0].path)
			return nil
		}, nil
	case "verbose":
		return func(file []*File) error {
			fmt.Println(file[0].path)
			return nil
		}, nil
	case "dontask":
		return func(file []*File) error {
			fmt.Println(file[0].path)
			return nil
		}, nil
	}
	return nil, errors.New("Undefined action")
}

func NewFile(path string, info os.FileInfo) *File {
	return &File{path, info, 0, []byte{}}
}

func (f *File) setIntro() error {
	fin, err := os.Open(f.path)
	if err != nil {
		return err
	}
	defer fin.Close()
	r := bufio.NewReader(fin)

	i := make([]byte, 1024)

	dbg.Printf("Reading intro for %s\n", f.path)
	io.ReadFull(r, i)
	f.intro = i
	return nil
}

func (f *File) IsRegular() bool {
	if strings.Contains(f.path, "Application Data") {
		return false
	}
	return (f.info.Mode()&os.ModeType == 0) || (f.info.Mode()&os.ModeType == os.ModeDir)
}

func (f *File) setCrc() error {
	fin, err := os.Open(f.path)
	if err != nil {
		return err
	}
	defer fin.Close()
	r := bufio.NewReader(fin)

	h := crc32.NewIEEE()

	dbg.Printf("Calculating CRC for %s\n", f.path)
	io.Copy(h, r)
	f.crc = h.Sum32()
	return nil
}

func (f *File) FileSize() int64 {
	return f.info.Size()
}

func (f *File) IsDir() bool {
	return f.info.IsDir()
}

func (f *File) Hash(ext bool) Hash {
	var hash Hash
	if ext {
		hash = Hash{f.info.Size(), filepath.Ext(f.path)}
	} else {
		hash = Hash{f.info.Size(), ""}
	}
	return hash
}

func (f *File) Intro() ([]byte, error) {
	if len(f.intro) == 0 {
		e := f.setIntro()
		if e != nil {
			return nil, e
		}
	}
	return f.intro, nil
}

func (f *File) Crc() uint32 {
	if f.crc == 0 {
		f.setCrc()
	}
	return f.crc
}

func (f *File) Equal(o *File) (bool, error) {

	i, e := f.Intro()
	if e != nil {
		return false, e
	}
	j, _ := o.Intro()

	if !bytes.Equal(i, j) {
		return false, nil
	}

	if f.Crc() == o.Crc() && f.Crc() != 0 {
		return true, nil
	}

	return false, nil
}
