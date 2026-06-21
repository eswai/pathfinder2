package ui

import "os/exec"

func openFile(path string) {
	exec.Command("open", path).Start()
}
