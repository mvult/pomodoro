package main

import (
	"encoding/json"
	"fmt"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/olekukonko/tablewriter"
	"log"
	"os"
	"time"
)

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
		if dateEqual(t.Start, time.Now()) {
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

func dateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
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
	f, err := os.Open("tone.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
