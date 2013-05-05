package main

import (
	"bufio"
	"bytes"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
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

func NewFile(path string, info os.FileInfo) *File {
	return &File{path, info, 0, []byte{}}
}

func (f *File) setIntro() {
	fin, err := os.Open(f.path)
	defer fin.Close()
	if err != nil {
		log.Fatal(err)
	}
	r := bufio.NewReader(fin)

	i := make([]byte, 1024)

	log.Printf("Reading intro for %s\n", f.path)
	io.ReadFull(r, i)
	f.intro = i
}

func (f *File) setCrc() {
	fin, err := os.Open(f.path)
	defer fin.Close()
	if err != nil {
		log.Fatal(err)
	}
	r := bufio.NewReader(fin)

	h := crc32.NewIEEE()

	log.Printf("Calculating CRC for %s\n", f.path)
	io.Copy(h, r)
	f.crc = h.Sum32()
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

func (f *File) Intro() []byte {
	if len(f.intro) == 0 {
		f.setIntro()
	}
	return f.intro
}

func (f *File) Crc() uint32 {
	if f.crc == 0 {
		f.setCrc()
	}
	return f.crc
}

func (f *File) Equal(o *File) bool {

	if !bytes.Equal(f.Intro(), o.Intro()) {
		return false
	}

	if f.Crc() == o.Crc() && f.Crc() != 0 {
		return true
	}

	return false
}
