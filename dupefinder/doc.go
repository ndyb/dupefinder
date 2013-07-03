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
package dupefinder
