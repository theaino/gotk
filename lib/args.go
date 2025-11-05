package lib

import (
	"bytes"
	"encoding/gob"
	"os"
	"os/exec"
)


type TkArgs struct {
	Root string
	Path string
	Source string
}

func GetArgs() (args TkArgs, err error) {
	err = gob.NewDecoder(os.Stdin).Decode(&args)
	return
}

func (a TkArgs) Call(cmdParts ...string) (cmd *exec.Cmd, err error) {
	cmdArgs := make([]string, 0)
	if len(cmdParts) > 1 {
		cmdArgs = cmdParts[1:]
	}
	cmd = exec.Command(cmdParts[0], cmdArgs...)
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(&a)
	cmd.Stdin = &buf
	return
}
