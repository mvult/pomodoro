package blocks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const riotAPIKey = "RGAPI-e5521523-3d62-4578-be70-2a50830e2ebf"
const riotPUUID = "rx45dmEeosGmsySyCJW7a5q1zLJySuHCTrmz7PQpnYZnPYMUeevoEeizSFAn79dneHXhvMRnelCOrg"

const matchIDsURLFmt = "https://americas.api.riotgames.com/lol/match/v5/matches/by-puuid/%s/ids?start=0&count=20&api_key=%s"
const matchURLFmt = "https://americas.api.riotgames.com/lol/match/v5/matches/%s?api_key=%s"

type matchInfo struct {
	Info struct {
		GameCreation int64 `json:"gameCreation"`
		GameDuration int64 `json:"gameDuration"`
	} `json:"info"`
}

var riotService *RiotService

func StartRiotMonitor() {
	// Initialize Riot service with database
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Printf("Riot: DATABASE_URL not set, blocking last game of day")
		// Continue without service - will block last game
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		service, err := NewRiotService(ctx, databaseURL)
		cancel()
		if err != nil {
			log.Printf("Riot: failed to initialize database service: %v", err)
			// Continue without service - will block last game
		} else {
			riotService = service
			defer riotService.Close()
			log.Printf("Riot: initialized database service")
		}
	}

	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		checkRiotMatches()
		<-ticker.C
	}
}

func checkRiotMatches() {
	matchIDs, err := fetchMatchIDs()
	if err != nil {
		log.Printf("Riot: failed to fetch match ids: %v", err)
		return
	}

	if len(matchIDs) == 0 {
		return
	}

	now := time.Now()
	maxGames := 2
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		maxGames = 3
	}
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	cutoff := startOfDay.UnixMilli()
	qualifying := 0
	weekAgo := now.AddDate(0, 0, -7)
	weekAgoStart := time.Date(weekAgo.Year(), weekAgo.Month(), weekAgo.Day(), 0, 0, 0, 0, now.Location())
	weekCutoff := weekAgoStart.UnixMilli()
	infoByMatch := make(map[string]matchInfo, len(matchIDs))
	lastWeekIDs := make([]string, 0, len(matchIDs))

	for _, id := range matchIDs {
		info, err := fetchMatchInfo(id)
		if err != nil {
			log.Printf("Riot: failed to fetch match %s: %v", id, err)
			continue
		}
		infoByMatch[id] = info
		if info.Info.GameCreation >= weekCutoff {
			lastWeekIDs = append(lastWeekIDs, id)
		}
	}

	// Check if last week is accounted for to allow last game of the day
	var lastWeekAccountedFor bool
	if riotService != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		accounted, err := riotService.IsLastWeekAccountedFor(ctx, lastWeekIDs)
		if err != nil {
			log.Printf("Riot: failed to check last week accounting: %v", err)
			lastWeekAccountedFor = false
		} else {
			lastWeekAccountedFor = accounted
		}
	} else {
		// No DB service - block last game
		lastWeekAccountedFor = false
		log.Printf("Riot: no database service, blocking last game of day")
	}

	for _, id := range matchIDs {
		info, ok := infoByMatch[id]
		if !ok {
			continue
		}
		if info.Info.GameCreation >= cutoff && info.Info.GameDuration > 15*60 {
			qualifying++

			// Check if we should kill
			shouldKill := false
			if qualifying >= maxGames {
				if qualifying == maxGames && lastWeekAccountedFor {
					// Last game of the day and last week is accounted for - allow it
					log.Printf("Riot: allowing last game of day (qualifying=%d, max=%d)", qualifying, maxGames)
					continue
				}
				// Exceeds daily limit or last week not accounted for
				shouldKill = true
			}

			if shouldKill {
				log.Printf("Riot: killing processes (qualifying=%d, max=%d, lastWeekAccountedFor=%v)", qualifying, maxGames, lastWeekAccountedFor)
				killRiotProcesses()
				return
			}
		}
	}
}

func fetchMatchIDs() ([]string, error) {
	url := fmt.Sprintf(matchIDsURLFmt, riotPUUID, riotAPIKey)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var ids []string
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func fetchMatchInfo(id string) (matchInfo, error) {
	url := fmt.Sprintf(matchURLFmt, id, riotAPIKey)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return matchInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return matchInfo{}, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var info matchInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return matchInfo{}, err
	}
	return info, nil
}

func killRiotProcesses() {
	needles := []string{
		"league of legends",
		"leagueclient",
		"riot client",
	}
	killed := 0
	for _, needle := range needles {
		for _, p := range getPIDsBySubstring(needle) {
			killProcess(p.PID)
			log.Printf("Riot: killed %s (pid %s)", p.CmdLine, p.PID)
			killed++
		}
	}
}

func getPIDsBySubstring(needle string) []processInfo {
	ps, err := getProcessList()
	if err != nil {
		log.Printf("riot: failed to list processes: %v", err)
		return nil
	}

	needle = strings.ToLower(needle)
	var processes []processInfo

	for _, p := range ps {
		if strings.Contains(strings.ToLower(p.CmdLine), needle) {
			processes = append(processes, p)
		}
	}

	return processes
}

func getProcessList() ([]processInfo, error) {
	cmdPs := exec.Command("ps", "aux")
	var outPs, errPs bytes.Buffer
	cmdPs.Stdout = &outPs
	cmdPs.Stderr = &errPs

	if err := cmdPs.Run(); err != nil {
		return nil, fmt.Errorf("error running ps aux: %v\n%s", err, errPs.String())
	}

	lines := strings.Split(outPs.String(), "\n")
	var processes []processInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		pid := fields[1]
		cmdLine := strings.Join(fields[10:], " ")
		processes = append(processes, processInfo{PID: pid, CmdLine: cmdLine})
	}

	return processes, nil
}
