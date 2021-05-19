package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func filterDirsGlob(dir, suffix string) ([]string, error) {
	return filepath.Glob(filepath.Join(dir, suffix))
}

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)

	fmt.Println("copy ", src, dst)
	return nBytes, err
}

func isFileEquals(src, dst string) bool {
	srcSize, err := os.Stat(src)
	if err != nil {
		panic(err)
	}

	dstSize, err := os.Stat(dst)
	if err != nil {
		return false
	}
	return srcSize.Size() == dstSize.Size()
}
