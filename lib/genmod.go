package lib

import (
	"os"
)

type GenMod struct {
	Root string
}

func CloneModule(sourcePath, destPath string) (GenMod, error) {
	os.RemoveAll(destPath)
	err := os.CopyFS(destPath, os.DirFS(sourcePath))
	return GenMod{Root: destPath}, err
}
