package dupefinder

import (
	"github.com/ndyb/utils/debug"
	"os"
	"path/filepath"
)

var dbg debug.Debug
var files = make(Files)

// Find duplicates starting from path and send list of duplicate files to channel output
func FindDuplicates(path string, minSize uint64, ext bool, output chan ([]*File)) {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {

		file := *NewFile(path, info)

		if !file.IsRegular() {
			dbg.Printf("Not regular file %s. Ignoring!\n", file.Path)
			// return filepath.SkipDir
			return nil
		}

		if file.IsDir() {
			dbg.Printf("Walking into %s\n", file.Path)
			return nil
		}

		if file.FileSize() < minSize {
			return nil
		}

		h := file.Hash(ext)

		_, exists := files[h]
		if exists {
			for i := range files[h] {
				equals, err := files[h][i].Equal(&file)
				if err != nil {
					dbg.Printf("Error accessing %s\n", file.Path)
					return nil
				}
				if equals {
					dbg.Printf("%s == %s\n", file.Path, files[h][i].Path)
					output <- []*File{&file, &files[h][i]}
					return nil
				}
			}
		}

		files[h] = append(files[h], file)

		return nil
	})

	close(output)
}
