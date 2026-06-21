package ui

import "os/exec"

func openFile(path string) {
	exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
}
