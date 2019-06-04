package winutils

import (
	"bytes"
	"fmt"
	"os/exec"
)

// RunPowershellScript abstract knowledge of running powershell scripts.
func RunPowershellScript(script string, args ...string) error {
	// Need to run a scripts implicitly using `powershell` command,
	// otherwise it's not working.
	// To execute script following args need to be provided:
	// -ExecutionPolicy Bypass -File <?script file path?>
	args = append([]string{"-ExecutionPolicy", "Bypass", "-File", script}, args...)
	return runPowershell(args)
}

func runPowershell(args []string) error {
	cmd := exec.Command("powershell", args...)

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	if err := cmd.Run(); err != nil {
		outStr, errStr := outbuf.String(), errbuf.String()

		return fmt.Errorf("%v\nout:\n%s\nerr:\n%s", err, outStr, errStr)
	}
	return nil
}
