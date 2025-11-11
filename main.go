package main

import (
	"fmt"
	"os"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/mvult/pomodoro/blocks"
)

type Pomodoro struct {
	Start         time.Time `json:"start"`
	End           time.Time `json:"end"`
	Activity      string    `json:"activity"`
	Busy          bool
	InterruptChan chan bool
}

type logEntries []Pomodoro

func main() {
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	for {
		err = eventLoop()
		if err != nil {
			fmt.Println("Event loop error", err)
		}
	}
}

func eventLoop() error {
	p := Pomodoro{InterruptChan: make(chan bool)}

	for {
		s, err := readLine()
		if err != nil {
			return err
		}

		switch s {
		case "l":
			p.Refresh("lunch", 30)

		case "zefr":
			p.Refresh("zefr", 30)

		case "I am weak and require dopamine to function":
			p.Refresh("break", 30)

		case "end":
			if p.Busy {
				fmt.Println("sending on chan")
				p.InterruptChan <- false
				fmt.Println("sent on chan")
				p.Busy = false
			}

		case "kill":
			fmt.Println("Killing manually")
			blocks.ActivateProfile("workHours")
			time.Sleep(time.Second * 5)
			os.Exit(0)

		default:
			p.Refresh(s, 30)

		}
	}
}

func (p *Pomodoro) Refresh(activity string, mins int) {
	if p.Busy {
		p.InterruptChan <- false
		p.Busy = false
	}

	p.Activity = activity
	p.Start = time.Now()
	p.End = time.Now().Add(time.Minute * time.Duration(mins))
	p.Busy = true

	if err := blocks.ActivateProfile(activity); err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		for {
			time.Sleep(time.Second)
			select {
			case <-p.InterruptChan:
				fmt.Println("Got interrupted")
				blocks.DeactivateProfile()
				return
			default:
				if time.Now().After(p.End) {
					fmt.Println("Deaqctivating", p.End)
					blocks.DeactivateProfile()
					p.Busy = false
					fmt.Print("\r                    \r") // Clear the line
					return
				}
				fmt.Printf("\r%v %v\r", p.Activity, time.Until(p.End).Round(time.Second))
			}
		}
	}()
	fmt.Println("Refresh finished")
}
