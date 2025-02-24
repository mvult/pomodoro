package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/mvult/pomodoro/blocks"
)

type Pomodoro struct {
	Start            time.Time `json:"start"`
	End              time.Time `json:"end"`
	TargetLength     int       `json:"target_length"`
	Activity         string    `json:"activity"`
	ActivityCategory string    `json:"activity_category"`
	Pauses           []Pause   `json:"pauses"`
	Complete         bool      `json:"complete"`
	Mute             bool      `json:"-"`
}

type Pause struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type logEntries []Pomodoro

func main() {
	sigChan := make(chan os.Signal, 100)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived %v, Cleaning up and shutting down in 5 seconds\n", sig)

		blocks.ActivateProfile("workHours")
		time.Sleep(time.Second * 5)
		os.Exit(0)
	}()

	for {
		err = runPomodoro()
		if err != nil {
			sigChan <- syscall.SIGINT
		}
	}
}

func runPomodoro() error {
	fmt.Println("Initializing")
	pomo := initializePomodoro()
	fmt.Println("Initialized")
	paused := false
	interruptChan, killChan := getInterruptChan()
	var pause Pause

	blocks.ActivateProfile(pomo.Activity)

out:
	for {
		select {
		case i := <-interruptChan:
			switch i {
			case "today":
				todaySummary()
			case "pause":
				if paused {
					pause.End = time.Now()
					pomo.Pauses = append(pomo.Pauses, pause)
					pomo.End = pomo.End.Add(pause.End.Sub(pause.Start))
					paused = false
					continue
				} else {
					paused = true
					pause = Pause{Start: time.Now()}
				}
			case "kill":
				pomo.Complete = false
				pomo.End = time.Now()
				close(killChan)
				break out

			case "complete":
				pomo.End = time.Now()
				pomo.Mute = true
				close(killChan)
				break out
			case "unblock":
				fmt.Println("Unblocking")
			case "sigint":
				return errors.New("sigint")
			}
		default:

			time.Sleep(time.Millisecond * 500)
			if paused {
				fmt.Printf("\rPaused %v       \r", fmtDuration(time.Since(pause.Start)))
				continue
			}

			d := time.Until(pomo.End)
			if d < 0 {
				close(killChan)
				break out
			}
			fmt.Printf("\r%v            \r", fmtDuration(d))
		}
	}

	close(interruptChan)

	terminateAndLog(pomo)
	return nil
}

func initializePomodoro() Pomodoro {
start:
	fmt.Print("Type a number of minutes to Pomodoro. Default 30: ")

	text, err := readLine()
	if err != nil {
		panic(err)
	}

	length := 30

	if strings.Trim(text, "\n\r") == "t" {
		todaySummary()
		goto start
	}

	if text != "" {
		length, err = strconv.Atoi(strings.Trim(text, "\n\r"))
		if err != nil || length > 48*60 {
			length = 30
		}
	}

	fmt.Print("\rWhat do you want to accomplish? ")
	text, err = readLine()
	if err != nil {
		panic(err)
	}

	text = strings.Trim(text, "\n\r")

	return Pomodoro{
		Start:        time.Now(),
		End:          time.Now().Add(time.Minute * time.Duration(length)),
		TargetLength: length,
		Activity:     text,
		Complete:     true,
	}
}

func getInterruptChan() (chan string, chan interface{}) {
	ret := make(chan string)
	killChan := make(chan interface{})

	go func() {
		for {
			select {
			case <-killChan:
				return
			default:
				char, key, err := keyboard.GetKey()
				if err != nil {
					return
				}
				fmt.Println(char, key)

				switch {
				case key == keyboard.KeySpace:
					ret <- "pause"
				case char == 't':
					ret <- "today"
				case char == 'x':
					ret <- "kill"
					return
				case char == 'c':
					ret <- "complete"
					return
				case key == keyboard.KeyCtrlC:
					fmt.Println("getting sig int manual")
					ret <- "sigint"
					return
				}

			}
		}
	}()

	return ret, killChan
}

func terminateAndLog(pomo Pomodoro) {
	blocks.DeactivateProfile()

	if pomo.Activity == "rest" {
		return
	}

	logFile, err := os.OpenFile("log.json", os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	var tmpEntries []Pomodoro
	err = json.NewDecoder(logFile).Decode(&tmpEntries)
	if err != nil {
		panic(err)
	}

	tmpEntries = append(tmpEntries, pomo)

	b, err := json.Marshal(tmpEntries)
	if err != nil {
		panic(err)
	}

	logFile.WriteAt(b, 0)

	if pomo.Complete && !pomo.Mute {
		endNoise()
	}
}
