package dupefinder

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

const (
	introSize = 1024 // Number of bytes to read when reading a File intro
)

// A file on the file system
type File struct {
	Path  string
	info  os.FileInfo
	crc   uint32
	intro []byte
}

// Hash used in hash table to identify similar files
type Hash struct {
	size int64
	ext  string
}

// Hash table of files indexed by Hash
type Files map[Hash][]File

// Function definition for action that can be performed on a list of duplicate files
type FileAction func(file []*File) error

func GetActionFor(s string) (FileAction, error) {
	switch s {
	case "print":
		return func(file []*File) error {
			fmt.Printf("%s\t%s\n", humanize.Bytes(uint64(file[0].FileSize())), file[0].Path)
			return nil
		}, nil
	case "delete":
		return func(file []*File) error {
			fmt.Println(file[0].Path)
			return nil
		}, nil
	case "verbose":
		return func(file []*File) error {
			fmt.Println(file[0].Path)
			return nil
		}, nil
	case "dontask":
		return func(file []*File) error {
			fmt.Println(file[0].Path)
			return nil
		}, nil
	}
	return nil, errors.New("Undefined action")
}

// Creates new file from path with given info
func NewFile(path string, info os.FileInfo) *File {
	return &File{path, info, 0, []byte{}}
}

func (f *File) setIntro() error {
	fin, err := os.Open(f.Path)
	if err != nil {
		return err
	}
	defer fin.Close()
	r := bufio.NewReader(fin)

	i := make([]byte, introSize)

	// dbg.Printf("Reading intro for %s\n", f.path)
	io.ReadFull(r, i)
	f.intro = i
	return nil
}

// Returns true if file is a regular file, or a directory.
// Windows (NTFS) junctions are not supported by standard library.
// Hack ignores "Application Data" folder
func (f *File) IsRegular() bool {
	if strings.Contains(f.Path, "Application Data") {
		return false
	}
	return (f.info.Mode()&os.ModeType == 0) || (f.info.Mode()&os.ModeType == os.ModeDir)
}

func (f *File) setCrc() error {
	fin, err := os.Open(f.Path)
	if err != nil {
		return err
	}
	defer fin.Close()
	r := bufio.NewReader(fin)

	h := crc32.NewIEEE()

	// dbg.Printf("Calculating CRC for %s\n", f.path)
	io.Copy(h, r)
	f.crc = h.Sum32()
	return nil
}

// Returns size if file in bytes
func (f *File) FileSize() uint64 {
	return uint64(f.info.Size())
}

// Returns true if File is a directory
func (f *File) IsDir() bool {
	return f.info.IsDir()
}

// Returns Hash of File
func (f *File) Hash(ext bool) Hash {
	var hash Hash
	if ext {
		hash = Hash{f.info.Size(), filepath.Ext(f.Path)}
	} else {
		hash = Hash{f.info.Size(), ""}
	}
	return hash
}

func (f *File) getIntro() ([]byte, error) {
	if len(f.intro) == 0 {
		e := f.setIntro()
		if e != nil {
			return nil, e
		}
	}
	return f.intro, nil
}

func (f *File) getCrc() uint32 {
	if f.crc == 0 {
		f.setCrc()
	}
	return f.crc
}

// Compares two files. First it compares the first few bytes of each file. If they are equal the files are compared by their CRC checksums
func (f *File) Equal(o *File) (bool, error) {

	i, e := f.getIntro()
	if e != nil {
		return false, e
	}
	j, _ := o.getIntro()

	if !bytes.Equal(i, j) {
		return false, nil
	}

	if f.getCrc() == o.getCrc() && f.getCrc() != 0 {
		return true, nil
	}

	return false, nil
}
