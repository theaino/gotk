package lib

import (
	"os"
)

type GenMod struct {
	Root string
}

func CloneModule(sourcePath, destPath string) (GenMod, error) {
	err := os.CopyFS(destPath, os.DirFS(sourcePath))
	return GenMod{Root: destPath}, err
}
