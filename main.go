package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"time"

	"gitlab.org/go-unicord-phat-lucian/ovapi"
	"gitlab.org/go-unicord-phat-lucian/unicornphat"
)

const (
	busLineNumber      = "35"
	busDestinationCode = "OLPP"
	secondsToPullOV    = 60
	minutesFilterETA   = 35
)

var (
	mapPanel0 = map[int]func(c *unicorn.Client, r, g, b uint){
		0: draw0Panel0,
		1: draw1Panel0,
		2: draw2Panel0,
		3: draw3Panel0,
		4: draw4Panel0,
		5: draw5Panel0,
		6: draw6Panel0,
		7: draw7Panel0,
		8: draw8Panel0,
		9: draw9Panel0,
	}
	mapPanel1 = map[int]func(c *unicorn.Client, r, g, b uint){
		0: draw0Panel1,
		1: draw1Panel1,
		2: draw2Panel1,
		3: draw3Panel1,
		4: draw4Panel1,
		5: draw5Panel1,
		6: draw6Panel1,
		7: draw7Panel1,
		8: draw8Panel1,
		9: draw9Panel1,
	}
)

type BusOfInterest struct {
	LineNumber           string    `json:"line_number"`
	TripStatus           string    `json:"trip_status"`
	FinalDestinationName string    `json:"final_destination_name"`
	FinalDestinationCode string    `json:"final_destination_code"`
	CompiledArrivalTime  time.Time `json:"compiled_arrival_time"` // compiled ETA
	PlannedArrivalTime   time.Time `json:"planned_arrival_time"`  // scheduled ETA
	LastUpdateTimestamp  time.Time `json:"last_update_timestamp"`
}

func transferOVAPItoBusOfInterest(ov *ovapi.OVAPIResponse) []BusOfInterest {
	var buses []BusOfInterest
	timeLayout := "2006-01-02T15:04:05-07:00"
	localTime := time.Now()

	// Process and filter each bus timetable.
	for _, pass := range ov.Station.DirectionOlofPalmeplein.Passes {
		if pass.LinePublicNumber != busLineNumber {
			//fmt.Printf("Filtered bus line number (%v)\n", pass.LinePublicNumber)
			continue
		}

		if pass.DestinationCode != busDestinationCode {
			fmt.Printf("Filtered bus destination code (%v)\n", pass.DestinationCode)
			continue
		}

		arrivalTime, err := time.Parse(timeLayout, pass.ExpectedArrivalTime+"+01:00")
		if err != nil {
			fmt.Printf("\nUnable to parse arrival time for pass (%#v)\n\n", pass)
			continue
		}

		/*
			Important. This constraints the number of buses to be displayed
		*/
		if arrivalTime.Sub(localTime).Minutes() > minutesFilterETA {
			//h, m, s := arrivalTime.Clock()
			//fmt.Printf("Filtered bus too far in the future %v:%v:%v (%v)\n", h, m, s, arrivalTime.Sub(localTime))
			continue
		}

		// Past events.
		if arrivalTime.Sub(localTime) < 0 {
			//h, m, s := arrivalTime.Clock()
			//fmt.Printf("Filtered bus in the past %v:%v:%v (%v)\n", h, m, s, arrivalTime.Sub(localTime))
			continue
		}

		// Last update timestamp.
		pass.LastUpdateTimeStamp = strings.Replace(pass.LastUpdateTimeStamp, "+0100", "+01:00", -1)
		lastUpdateTimestamp, err := time.Parse(timeLayout, pass.LastUpdateTimeStamp)
		if err != nil {
			fmt.Printf("Skipped faulty pass LastUpdateTimeStamp (%v): %v\n", pass.LastUpdateTimeStamp, err)
			continue
		}

		// Planned time arrival.
		plannedArrivalTime, err := time.Parse(timeLayout, pass.TargetArrivalTime+"+01:00")
		if err != nil {
			fmt.Printf("Skipped faulty pass TargetArrivalTime (%v): %v\n", pass.TargetArrivalTime, err)
			continue
		}

		var bus BusOfInterest
		bus.LineNumber = pass.LinePublicNumber
		bus.TripStatus = pass.TripStopStatus
		bus.FinalDestinationName = pass.DestinationName50
		bus.FinalDestinationCode = pass.DestinationCode

		bus.LastUpdateTimestamp = lastUpdateTimestamp
		bus.CompiledArrivalTime = arrivalTime
		bus.PlannedArrivalTime = plannedArrivalTime

		buses = append(buses, bus)
	}

	return buses
}

func pullOV(out chan []int) {
	ov, err := ovapi.RequestDataFromOV()
	if err != nil {
		log.Fatal(err)
	}

	frames := make([]int, 0)
	buses := transferOVAPItoBusOfInterest(ov)
	for _, bus := range buses {
		s := fmt.Sprintf("> Bus (%v) is (%v) w/ ETA: (%v) aka (%v), offset (%v)",
			bus.LineNumber,
			bus.TripStatus,
			bus.CompiledArrivalTime,
			bus.CompiledArrivalTime.Sub(time.Now()),
			bus.PlannedArrivalTime.Sub(bus.CompiledArrivalTime),
		)
		fmt.Println(s)

		min := bus.CompiledArrivalTime.Sub(time.Now()).Minutes()
		frames = append(frames, int(min))
	}

	out <- frames
}

func sendToDisplay(num int, c *unicorn.Client) {
	c.Clear()

	switch {
	case num <= 3:
		mapPanel0[num](c, 255, 0, 0) // red
	case num > 3 && num <= 5:
		mapPanel0[num](c, 230, 150, 0) // orange
	case num > 5 && num < 10:
		mapPanel1[num](c, 0, 255, 0) // green. Sweet spot.
	case num >= 10:
		s := strconv.Itoa(num)
		firstNumStr := s[:1]
		SecondNumStr := s[1:]
		firstNumInt, _ := strconv.Atoi(firstNumStr)
		SecondNumInt, _ := strconv.Atoi(SecondNumStr)
		mapPanel0[int(firstNumInt)](c, 255, 255, 255)
		mapPanel1[int(SecondNumInt)](c, 255, 255, 255)
	default:
		drawExclamationMark(c, 255, 0, 0)
	}

	c.Show()
}

func main() {
	log.Println("Starting up...")
	dataChan := make(chan []int)
	go func() {
		for {
			pullOV(dataChan)
			time.Sleep(time.Second * secondsToPullOV)
		}
	}()

	fmt.Println("Starting unicorn client...")
	c := unicorn.NewClient(false, "")
	if err := c.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	pixels := [64]unicorn.Pixel{}
	for i := range pixels {
		pixels[i] = unicorn.Pixel{
			R: 0,
			G: 0,
			B: 0,
		}
	}

	c.SetBrightness(10)
	c.SetAllPixels(pixels)
	c.Show()
	time.Sleep(time.Second * 2)
	c.Clear()
	log.Println("Display init OK. Waiting for input...")

	go func() {
		for {
			minutes := <-dataChan
			sort.Ints(minutes)
			go func() {
				ts := time.Now()
				for {
					for _, item := range minutes {
						// Kill goroutine before the new call is made.
						if time.Since(ts).Seconds() >= secondsToPullOV-1 {
							log.Println("Closed. Waiting for new data.")
							return
						}

						c.Clear()
						sendToDisplay(item, c)
						time.Sleep(time.Second * 2)
					}
				}
			}()
		}
	}()

	// Block and clear display on closing down.
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, os.Kill)
	for {
		<-ch
		c.Clear()
		os.Exit(0)
	}
}

func drawExclamationMark(c *unicorn.Client, r, g, b uint) {
	for i := 0; i <= 7; i++ {
		if i == 5 {
			c.SetPixel(0, uint(i), 0, 0, 0)
		} else {
			if err := c.SetPixel(0, uint(i), r, g, b); err != nil {
				log.Fatalf("Error setting pixel: %v", err)
			}
		}
	}

	for j := 7; j >= 0; j-- {
		if j == 2 {
			c.SetPixel(1, uint(j), 0, 0, 0)
		} else {
			if err := c.SetPixel(1, uint(j), r, g, b); err != nil {
				log.Fatalf("Error setting pixel: %v", err)
			}
		}
	}
}

func px(c *unicorn.Client, x, y, r, g, b uint) {
	c.SetPixel(uint(x), uint(y), r, g, b)
}

func draw0Panel0(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 7, r, g, b)
	px(c, 3, 6, r, g, b)
	px(c, 3, 5, r, g, b)
	px(c, 2, 0, r, g, b)
	px(c, 2, 2, r, g, b)
	px(c, 1, 7, r, g, b)
	px(c, 1, 5, r, g, b)
	px(c, 0, 0, r, g, b)
	px(c, 0, 1, r, g, b)
	px(c, 0, 2, r, g, b)
}

func draw0Panel1(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 2, r, g, b)
	px(c, 3, 1, r, g, b)
	px(c, 3, 0, r, g, b)
	px(c, 2, 5, r, g, b)
	px(c, 2, 7, r, g, b)
	px(c, 1, 2, r, g, b)
	px(c, 1, 0, r, g, b)
	px(c, 0, 5, r, g, b)
	px(c, 0, 6, r, g, b)
	px(c, 0, 7, r, g, b)
}

func draw1Panel0(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 6, r, g, b)
	px(c, 2, 0, r, g, b)
	px(c, 2, 1, r, g, b)
	px(c, 1, 6, r, g, b)
	px(c, 0, 1, r, g, b)
}

func draw1Panel1(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 1, r, g, b)
	px(c, 2, 5, r, g, b)
	px(c, 2, 6, r, g, b)
	px(c, 1, 1, r, g, b)
	px(c, 0, 6, r, g, b)
}

func draw2Panel0(c *unicorn.Client, r, g, b uint) {
	px(c, 2, 0, r, g, b)
	px(c, 3, 6, r, g, b)
	px(c, 3, 5, r, g, b)
	px(c, 2, 2, r, g, b)
	px(c, 1, 6, r, g, b)
	px(c, 0, 0, r, g, b)
	px(c, 0, 1, r, g, b)
	px(c, 0, 2, r, g, b)
}

func draw2Panel1(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 1, r, g, b)
	px(c, 3, 0, r, g, b)
	px(c, 2, 5, r, g, b)
	px(c, 2, 7, r, g, b)
	px(c, 1, 1, r, g, b)
	px(c, 0, 5, r, g, b)
	px(c, 0, 6, r, g, b)
	px(c, 0, 7, r, g, b)
}

func draw3Panel0(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 7, r, g, b)
	px(c, 3, 6, r, g, b)
	px(c, 3, 5, r, g, b)
	px(c, 2, 2, r, g, b)
	px(c, 2, 1, r, g, b)
	px(c, 1, 5, r, g, b)
	px(c, 0, 2, r, g, b)
	px(c, 0, 1, r, g, b)
	px(c, 0, 0, r, g, b)
}

func draw3Panel1(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 2, r, g, b)
	px(c, 3, 1, r, g, b)
	px(c, 3, 0, r, g, b)
	px(c, 2, 6, r, g, b)
	px(c, 2, 7, r, g, b)
	px(c, 1, 0, r, g, b)
	px(c, 0, 5, r, g, b)
	px(c, 0, 6, r, g, b)
	px(c, 0, 7, r, g, b)
}

func draw4Panel0(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 7, r, g, b)
	px(c, 2, 0, r, g, b)
	px(c, 2, 1, r, g, b)
	px(c, 2, 2, r, g, b)
	px(c, 1, 5, r, g, b)
	px(c, 0, 2, r, g, b)
}

func draw4Panel1(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 2, r, g, b)
	px(c, 2, 5, r, g, b)
	px(c, 2, 6, r, g, b)
	px(c, 2, 7, r, g, b)
	px(c, 1, 0, r, g, b)
	px(c, 0, 7, r, g, b)
}

func draw5Panel0(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 7, r, g, b)
	px(c, 3, 6, r, g, b)
	px(c, 3, 5, r, g, b)
	px(c, 2, 0, r, g, b)
	px(c, 1, 6, r, g, b)
	px(c, 0, 0, r, g, b)
	px(c, 0, 1, r, g, b)
	px(c, 0, 2, r, g, b)
}

func draw5Panel1(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 2, r, g, b)
	px(c, 3, 1, r, g, b)
	px(c, 3, 0, r, g, b)
	px(c, 2, 5, r, g, b)
	px(c, 1, 1, r, g, b)
	px(c, 0, 5, r, g, b)
	px(c, 0, 6, r, g, b)
	px(c, 0, 7, r, g, b)
}

func draw6Panel0(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 7, r, g, b)
	px(c, 3, 6, r, g, b)
	px(c, 3, 5, r, g, b)
	px(c, 2, 0, r, g, b)
	px(c, 1, 7, r, g, b)
	px(c, 1, 6, r, g, b)
	px(c, 1, 5, r, g, b)
	px(c, 0, 0, r, g, b)
	px(c, 0, 1, r, g, b)
	px(c, 0, 2, r, g, b)
}

func draw6Panel1(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 2, r, g, b)
	px(c, 3, 1, r, g, b)
	px(c, 3, 0, r, g, b)
	px(c, 2, 5, r, g, b)
	px(c, 1, 2, r, g, b)
	px(c, 1, 1, r, g, b)
	px(c, 1, 0, r, g, b)
	px(c, 0, 5, r, g, b)
	px(c, 0, 6, r, g, b)
	px(c, 0, 7, r, g, b)
}

func draw7Panel0(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 7, r, g, b)
	px(c, 3, 6, r, g, b)
	px(c, 3, 5, r, g, b)
	px(c, 2, 2, r, g, b)
	px(c, 1, 6, r, g, b)
	px(c, 0, 0, r, g, b)
}

func draw7Panel1(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 2, r, g, b)
	px(c, 3, 1, r, g, b)
	px(c, 3, 0, r, g, b)
	px(c, 2, 7, r, g, b)
	px(c, 1, 1, r, g, b)
	px(c, 0, 5, r, g, b)
}

func draw8Panel0(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 7, r, g, b)
	px(c, 3, 6, r, g, b)
	px(c, 3, 5, r, g, b)
	px(c, 2, 0, r, g, b)
	px(c, 2, 2, r, g, b)
	px(c, 1, 7, r, g, b)
	px(c, 1, 6, r, g, b)
	px(c, 1, 5, r, g, b)
	px(c, 0, 0, r, g, b)
	px(c, 0, 1, r, g, b)
	px(c, 0, 2, r, g, b)
}

func draw8Panel1(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 2, r, g, b)
	px(c, 3, 1, r, g, b)
	px(c, 3, 0, r, g, b)
	px(c, 2, 5, r, g, b)
	px(c, 2, 7, r, g, b)
	px(c, 1, 2, r, g, b)
	px(c, 1, 1, r, g, b)
	px(c, 1, 0, r, g, b)
	px(c, 0, 5, r, g, b)
	px(c, 0, 6, r, g, b)
	px(c, 0, 7, r, g, b)
}

func draw9Panel0(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 7, r, g, b)
	px(c, 3, 6, r, g, b)
	px(c, 3, 5, r, g, b)
	px(c, 2, 0, r, g, b)
	px(c, 2, 1, r, g, b)
	px(c, 2, 2, r, g, b)
	px(c, 1, 5, r, g, b)
	px(c, 0, 2, r, g, b)
}

func draw9Panel1(c *unicorn.Client, r, g, b uint) {
	px(c, 3, 2, r, g, b)
	px(c, 3, 1, r, g, b)
	px(c, 3, 0, r, g, b)
	px(c, 2, 5, r, g, b)
	px(c, 2, 6, r, g, b)
	px(c, 2, 7, r, g, b)
	px(c, 1, 0, r, g, b)
	px(c, 0, 7, r, g, b)
}
