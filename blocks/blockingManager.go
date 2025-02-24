package blocks

import (
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

func ActivateProfile(s string) {
	singleton.lock.Lock()
	defer singleton.lock.Unlock()

	if s == "lunch" && time.Since(singleton.lastLunch) < time.Hour*12 {
		fmt.Println("Had lunch too recently")
	} else {
		singleton.activity = s
	}
}

func DeactivateProfile() {
	singleton.lock.Lock()
	defer singleton.lock.Unlock()
	singleton.activity = ""
}

func manageState() {
	for {

		time.Sleep(time.Second * 10)

		switch {
		case singleton.activity == singleton.currentActivity:

		case singleton.activity != singleton.currentActivity:
			if singleton.activity == "" {
				removeBlocksToActivity(singleton.currentActivity)
			} else {
				removeBlocksToActivity(singleton.currentActivity)
				applyBlocksToActivity(singleton.activity)
			}

			singleton.currentActivity = singleton.activity
		}

		if singleton.activity == "" && workHours() {
			if !singleton.workHoursLimitApplied {
				applyBlocksToActivity("workHours")
				singleton.workHoursLimitApplied = true
			}
		}

		if singleton.activity == "" && !workHours() {
			if singleton.workHoursLimitApplied {

				removeBlocksToActivity("workHours")
				singleton.workHoursLimitApplied = false
			}
		}

	}
}

func workHours() bool {
	loc, _ := time.LoadLocation("America/Mexico_City")
	now := time.Now().In(loc)
	start := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, loc)
	end := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, loc)
	return now.After(start) && now.Before(end)
}
