package blocks

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type processInfo struct {
	PID     string
	CmdLine string // The full command line that started the process
}

func killProcesses(activity string) {
	bls, ok := blocks[activity]
	if ok {
		if err := blockURLs(bls); err != nil {
			log.Fatal(err)
		}
	}

	for _, b := range bls {
		if b.IsExecutable {
			ps := getPIDs(b.ProcessOrUrl)
			for _, p := range ps {
				killProcess(p.PID)
			}
		}
	}
}

func getPIDs(s string) []processInfo {
	cmdPs := exec.Command("ps", "aux")
	var outPs, errPs bytes.Buffer
	cmdPs.Stdout = &outPs
	cmdPs.Stderr = &errPs

	if err := cmdPs.Run(); err != nil {
		fmt.Printf("Error running ps aux: %v\n%s\n", err, errPs.String())
		return []processInfo{}
	}

	// Grep for the search string
	cmdGrep := exec.Command("grep", s)
	cmdGrep.Stdin = &outPs // Pipe ps output to grep
	var outGrep, errGrep bytes.Buffer
	cmdGrep.Stdout = &outGrep
	cmdGrep.Stderr = &errGrep

	if err := cmdGrep.Run(); err != nil {
		// grep returns an error if no lines match, which is fine
		if err.Error() == "exit status 1" {
			return []processInfo{}
		}
		fmt.Printf("Error running grep: %v\n%s\n", err, errGrep.String())
		return []processInfo{}
	}

	lines := strings.Split(outGrep.String(), "\n")
	var processes []processInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "grep "+s) {
			continue
		}

		// Parse PID and the full command line
		fields := strings.Fields(line)
		if len(fields) < 11 { // ps aux output typically has more than 10 fields before command
			continue
		}

		pid := fields[1]                          // PID is usually the second field
		cmdLine := strings.Join(fields[10:], " ") // Full command line

		processes = append(processes, processInfo{PID: pid, CmdLine: cmdLine})
	}

	if len(processes) == 0 {
		return []processInfo{}
	}

	return processes
}

func killProcess(pid string) {
	cmdKill := exec.Command("kill", pid)
	var errKill bytes.Buffer
	cmdKill.Stderr = &errKill
	if err := cmdKill.Run(); err != nil {
		fmt.Printf("    Error killing PID %s: %v\n%s\n", pid, err, errKill.String())
	} else {
		fmt.Printf("    Successfully killed PID %s.\n", pid)
	}
}
