package core

import (
	"os/exec"
)

// MoveToTrash moves the given path to the Windows Recycle Bin via PowerShell.
func MoveToTrash(path string) error {
	script := `$shell = New-Object -ComObject Shell.Application; $item = $shell.Namespace(0).ParseName("` + path + `"); $item.InvokeVerb("delete")`
	return exec.Command("powershell", "-NoProfile", "-Command", script).Run()
}
