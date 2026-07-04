package blocks

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var singleton *blocksState

type blocksState struct {
	activity              string
	currentActivity       string
	lock                  sync.Mutex
	workHoursLimitApplied bool
	lastLunch             time.Time
}

func init() {
	singleton = &blocksState{}
	go manageState()
}

func ActivateProfile(s string) error {
	singleton.lock.Lock()
	defer singleton.lock.Unlock()

	if s == "lunch" && time.Since(singleton.lastLunch) < time.Hour*12 {
		fmt.Println("Had lunch too recently")
		return errors.New("had lunch too recently")
	} else {
		singleton.activity = s
	}
	return nil
}

func DeactivateProfile() {
	singleton.lock.Lock()
	defer singleton.lock.Unlock()
	singleton.activity = ""
}

func manageState() {
	unblockURLs()

	for {

		time.Sleep(time.Second * 3)

		switch {
		case singleton.activity == singleton.currentActivity:
			if singleton.activity == "" && workHours() {
				killProcessesOnly("workHours")
			}
		case singleton.activity != singleton.currentActivity:
			unblockURLs()
			singleton.workHoursLimitApplied = false

			if singleton.activity != "" {
				applyBlocksToActivity(singleton.activity)
			}

			singleton.currentActivity = singleton.activity
		}

		if singleton.activity == "" && workHours() {
			if !singleton.workHoursLimitApplied {
				fmt.Println("Activating work hours")
				killProcesses("workHours")
				singleton.workHoursLimitApplied = true
			}
		}

		if singleton.activity == "" && !workHours() {
			if singleton.workHoursLimitApplied {
				unblockURLs()
				singleton.workHoursLimitApplied = false
			}
		}

	}
}

func workHours() bool {
	loc, _ := time.LoadLocation("America/Mexico_City")
	now := time.Now().In(loc)

	if int(now.Weekday()) == 6 || int(now.Weekday()) == 0 {
		return false
	}

	start := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, loc)
	end := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, loc)
	return now.After(start) && now.Before(end)
}
