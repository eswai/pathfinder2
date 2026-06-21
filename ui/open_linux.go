package ui

import "os/exec"

func openFile(path string) {
	exec.Command("xdg-open", path).Start()
}
