package blocks

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func getHostsFile() string {
	switch runtime.GOOS {
	case "windows":
		return `C:\Windows\System32\Drivers\etc\hosts`
	case "darwin":
		return "/etc/hosts"
	}
	log.Fatal("Unsupported platform")
	return ""
}

func getLineEnd() string {
	switch runtime.GOOS {
	case "windows":
		return "\r\n"
	case "darwin":
		return "\n"
	}
	log.Fatal("Unsupported platform")
	return ""
}

type block struct {
	IsExecutable bool
	ProcessOrUrl string
}

var blocks = map[string][]block{
	"ad": {
		block{false, "www.endoftheinter.net"},
		block{false, "boards.endoftheinter.net"},
		block{false, "www.metacritic.com"},
		block{false, "www.vox.com"},
		block{false, "www.nytimes.com"},
		block{false, "www.realclearpolitics.com"},
	},
	"zefr": {
		block{false, "www.endoftheinter.net"},
		block{false, "boards.endoftheinter.net"},
		block{false, "www.metacritic.com"},
		block{false, "www.facebook.com"},
		block{false, "www.vox.com"},
		block{false, "www.nytimes.com"},
		block{false, "www.realclearpolitics.com"},
	},
	"coding": {
		block{false, "www.youtube.com"},
		block{false, "youtube.com"},
		block{false, "youtu.be"},
		block{false, "googlevideo.com"},
		block{false, "youtube-nocookie.com"},
		block{false, "youtube.googleapis.com"},
		block{false, "youtubei.googleapis.com"},
		block{false, "ytimg.com"},
		block{false, "ytimg.l.google.com"},
		block{false, "www.endoftheinter.net"},
		block{false, "boards.endoftheinter.net"},
		block{false, "www.metacritic.com"},
		block{false, "www.wikipedia.org"},
		block{false, "www.vox.com"},
		block{false, "www.nytimes.com"},
		block{false, "www.realclearpolitics.com"},
		block{false, "www.reddit.com"},
		block{false, "www.x.com"},
		block{false, "x.com"},
	},
	"ceo": {
		block{false, "www.youtube.com"},
		block{false, "youtube.com"},
		block{false, "youtu.be"},
		block{false, "googlevideo.com"},
		block{false, "youtube-nocookie.com"},
		block{false, "youtube.googleapis.com"},
		block{false, "youtubei.googleapis.com"},
		block{false, "ytimg.com"},
		block{false, "ytimg.l.google.com"},
		block{false, "endoftheinter.net"},
		block{false, "boards.endoftheinter.net"},
		block{false, "www.metacritic.com"},
		block{false, "www.wikipedia.org"},
		block{false, "www.vox.com"},
		block{false, "www.nytimes.com"},
		block{false, "www.realclearpolitics.com"},
		block{false, "www.reddit.com"},
		block{false, "www.x.com"},
		block{false, "x.com"},
	},
	"workHours": {
		block{false, "www.youtube.com"},
		block{false, "youtube.com"},
		block{false, "youtu.be"},
		block{false, "googlevideo.com"},
		block{false, "youtube-nocookie.com"},
		block{false, "youtube.googleapis.com"},
		block{false, "youtubei.googleapis.com"},
		block{false, "ytimg.com"},
		block{false, "ytimg.l.google.com"},
		block{false, "www.endoftheinter.net"},
		block{false, "boards.endoftheinter.net"},
		block{false, "www.metacritic.com"},
		block{false, "www.vox.com"},
		block{false, "www.nytimes.com"},
		block{false, "www.realclearpolitics.com"},
		block{false, "www.facebook.com"},
		block{false, "www.reddit.com"},
		block{false, "www.x.com"},
		block{false, "x.com"},
		block{true, "/Applications/WhatsApp.app/Contents/MacOS/WhatsApp"},
	},
	"break": {},
	"lunch": {},
}

func applyBlocksToActivity(activity string) error {
	fmt.Printf("\nActivating %v profile\n", activity)
	bls, ok := blocks[activity]
	if ok {
		if err := blockURLs(bls); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func blockURLs(bls []block) error {
	file, err := os.OpenFile(getHostsFile(), os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, b := range bls {
		if b.IsExecutable {
			continue
		} else {
			// if _, err := file.WriteString(fmt.Sprintf(getLineEnd()+"127.0.0.1\t%v #pomo", b.ProcessOrUrl)); err != nil {
			if _, err := fmt.Fprintf(file, string(getLineEnd()+"127.0.0.1\t%v #pomo"), b.ProcessOrUrl); err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	file.WriteString(getLineEnd())

	return clearDNSCache()
}

func clearDNSCache() error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("ipconfig", `/flushdns`)
		return cmd.Run()
	case "darwin":
		cmd := exec.Command("sudo", "dscacheutil", "-flushcache")
		if err := cmd.Run(); err != nil {
			return err
		}

		cmd = exec.Command("sudo", "killall", "-HUP", "mDNSResponder")

		return cmd.Run()
	}
	return nil
}

func unblockURLs() error {
	fmt.Println("Unblocking")
	file, err := os.OpenFile(getHostsFile(), os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var b bytes.Buffer

	for scanner.Scan() {
		text := scanner.Text()
		if !strings.Contains(text, "#pomo") {
			b.Write(append(scanner.Bytes(), []byte(getLineEnd())...))
		}
	}

	file.Truncate(0)
	file.Seek(0, 0)
	if _, err := file.Write(b.Bytes()); err != nil {
		return err
	}
	return nil
}
