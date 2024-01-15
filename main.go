package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/olekukonko/tablewriter"
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
	for {
		runPomodoro()
	}
}

func initializePomodoro() Pomodoro {
start:
	fmt.Print("Type a number of minutes to Pomodoro. Default 30: ")
	reader := bufio.NewReader(os.Stdin)

	text, err := reader.ReadString('\n')
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
	text, err = reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	text = strings.Trim(text, "\n\r")

	return Pomodoro{Start: time.Now(),
		End:          time.Now().Add(time.Minute * time.Duration(length)),
		TargetLength: length,
		Activity:     text,
		Complete:     true,
	}

}

func runPomodoro() {
	pomo := initializePomodoro()
	paused := false
	interruptChan, killChan := getInterruptChan()
	var pause Pause

	if err := applyBlocksToActivity(pomo.Activity); err != nil {
		log.Fatal(err)
	}

	func() {
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
					return

				case "complete":
					pomo.End = time.Now()
					pomo.Mute = true
					close(killChan)
					return
				}
			default:

				time.Sleep(time.Millisecond * 500)
				if paused {
					fmt.Printf("\rPaused %v       \r", fmtDuration(time.Now().Sub(pause.Start)))
					continue
				}

				d := pomo.End.Sub(time.Now())
				if d < 0 {
					close(killChan)
					return
				}
				fmt.Printf("\r%v            \r", fmtDuration(d))
			}
		}
	}()

	terminateAndLog(pomo)
}

func terminateAndLog(pomo Pomodoro) {
	if err := removeBlocksToActivity(pomo.Activity); err != nil {
		log.Fatal(err)
	}

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

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func endNoise() {
	cmd := exec.Command("python", "py.py")
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func getInterruptChan() (chan string, chan interface{}) {
	ret := make(chan string)
	tmpChan := make(chan string)
	killChan := make(chan interface{})

	go func() {
		err := keyboard.Open()
		if err != nil {
			panic(err)
		}
		defer keyboard.Close()

		go func() {
			for {
				char, key, err := keyboard.GetKey()
				if err != nil {
					return
				}

				switch {
				case key == keyboard.KeySpace:
					tmpChan <- "pause"
				case char == 't':
					tmpChan <- "today"
				case char == 'x':
					tmpChan <- "kill"
					return
				case char == 'c':
					tmpChan <- "complete"
					return
				}
			}
		}()

		for {
			select {
			case _ = <-killChan:
				return
			case tmp := <-tmpChan:
				ret <- tmp
			}
		}
	}()

	return ret, killChan
}

func todaySummary() {
	logFile, err := os.OpenFile("log.json", os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	var tmpEntries logEntries
	err = json.NewDecoder(logFile).Decode(&tmpEntries)
	if err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Count", "Start", "Length", "Complete", "Category", "Activity"})

	count := 0
	for _, t := range tmpEntries {
		if DateEqual(t.Start, time.Now()) {
			count++
			table.Append([]string{fmt.Sprint(count), t.Start.Format(time.Kitchen), fmt.Sprint(t.TargetLength), fmt.Sprint(t.Complete), t.ActivityCategory, t.Activity})
		}
	}

	if count > 0 {
		table.Render()
	} else {
		fmt.Println("No activity today")
	}
}

func DateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
