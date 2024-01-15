package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const hosts = `C:\Windows\System32\Drivers\etc\hosts`

type block struct {
	IsExecutable bool
	ProcessOrUrl string
}

var blocks = map[string][]block{
	"Coding": {
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
		block{false, "whatsapp.com"},
		block{false, "metacritic.com"},
		block{false, "wikipedia.org"},
		block{false, "facebook.com"},
	},
}

func applyBlocksToActivity(activity string) error {
	bls, ok := blocks[activity]
	if ok {
		return blockURLs(bls)
	}
	return nil
}

func removeBlocksToActivity(activity string) error {
	bls, ok := blocks[activity]
	if ok {
		return unblockURLs(bls)
	}
	return nil
}

func blockURLs(bls []block) error {
	file, err := os.OpenFile(hosts, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, b := range bls {
		if b.IsExecutable {
			continue
		} else {
			if _, err := file.WriteString(fmt.Sprintf("\r\n127.0.0.1\t%v", b.ProcessOrUrl)); err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	file.WriteString("\r\n")

	return clearDNSCache()
}

func clearDNSCache() error {
	cmd := exec.Command("ipconfig", `/flushdns`)
	return cmd.Run()
}

func unblockURLs(bls []block) error {
	list := blocksToList(bls)

	file, err := os.OpenFile(hosts, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	keep := true
	var b bytes.Buffer

	for scanner.Scan() {
		keep = true
		for _, i := range list {
			if strings.Contains(scanner.Text(), i) {
				keep = false
				break
			}
		}
		if keep {
			b.Write(append(scanner.Bytes(), []byte("\r\n")...))
		}
	}

	file.Truncate(0)
	file.Seek(0, 0)
	if _, err := file.Write(b.Bytes()); err != nil {
		return err
	}
	return nil

}

func blocksToList(bls []block) []string {
	ret := []string{}
	for _, b := range bls {
		if !b.IsExecutable {
			ret = append(ret, b.ProcessOrUrl)
		}
	}
	return ret
}
