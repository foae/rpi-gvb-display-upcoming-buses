package mapsgvbnl

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// WebsocketMessage defines the structure of a GVB websocket json.
type WebsocketMessage struct {
	Trip struct {
		ID            string `json:"id,omitempty"`
		Number        string `json:"number,omitempty"`
		OperatorKey   string `json:"operatorKey,omitempty"`
		OperatingDate string `json:"operatingDate,omitempty"`

		LastCallMade struct {
			StopCode           string     `json:"stopCode,omitempty"`
			StopName           string     `json:"stopName,omitempty"`
			CallOrder          int        `json:"callOrder,omitempty"`
			Status             string     `json:"status,omitempty"`
			PlannedArrivalAt   string     `json:"plannedArrivalAt,omitempty"`
			PlannedDepartureAt string     `json:"plannedDepartureAt,omitempty"`
			Delay              int        `json:"delay,omitempty"`
			LiveArrivalAt      string     `json:"liveArrivalAt,omitempty"`
			LiveDepartureAt    string     `json:"liveDepartureAt,omitempty"`
			IsExtrapolated     bool       `json:"isExtrapolated,omitempty"`
			LastUpdateAt       *time.Time `json:"lastUpdateAt,omitempty"`
		} `json:"lastCallMade,omitempty"`
	} `json:"trip,omitempty"`

	Journey struct {
		ID          int    `json:"id,omitempty"`
		Code        string `json:"code,omitempty"`
		LineID      int    `json:"lineId,omitempty"`
		LineNumber  string `json:"lineNumber,omitempty"`
		Vehicletype string `json:"vehicletype,omitempty"`
		Destination string `json:"destination,omitempty"`
	} `json:"journey,omitempty"`

	Calls []struct {
		StopCode           string     `json:"stopCode,omitempty"`
		StopName           string     `json:"stopName,omitempty"`
		CallOrder          int        `json:"callOrder,omitempty"`
		Status             string     `json:"status,omitempty"`
		PlannedArrivalAt   string     `json:"plannedArrivalAt,omitempty"`
		PlannedDepartureAt string     `json:"plannedDepartureAt,omitempty"`
		Delay              int        `json:"delay,omitempty"`
		LiveArrivalAt      string     `json:"liveArrivalAt,omitempty"`
		LiveDepartureAt    string     `json:"liveDepartureAt,omitempty"`
		IsExtrapolated     bool       `json:"isExtrapolated,omitempty"`
		LastUpdateAt       *time.Time `json:"lastUpdateAt,omitempty"`
	} `json:"calls,omitempty"`

	Time int `json:"time,omitempty"`
}

func cleanJSON(b []byte) []byte {
	str := string(b)
	str = strings.Replace(str, `\`, "", -1)
	str = strings.Replace(str, `[8,"/stops/01346","`, "", 1)
	str = strings.Replace(str, `}"]`, "}", 1)

	return []byte(str)
}

func FromClockToTime(s string) (*time.Time, error) {
	strsTime := strings.Split(s, ":")
	if len(strsTime) != 3 {
		return nil, errors.New("fromClockToTime: need time in format `23:45:30`")
	}

	hours, err := strconv.ParseInt(strsTime[0], 10, 32)
	if err != nil {
		return nil, err
	}

	mins, _ := strconv.ParseInt(strsTime[1], 10, 32)
	if err != nil {
		return nil, err
	}

	secs, _ := strconv.ParseInt(strsTime[2], 10, 32)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Europe/Amsterdam")
	t := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), int(hours), int(mins), int(secs), time.Now().Nanosecond(), loc)
	return &t, nil
}

func ConnectAndListen() chan WebsocketMessage {
	log.Println("Connecting to websocket...")
	d := websocket.Dialer{
		EnableCompression: true,
		HandshakeTimeout:  time.Minute * 15,
	}

	ws, _, err := d.Dial("wss://maps-wss.gvb.nl/", http.Header{
		"User-Agent": []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.102 Safari/537.36"},
	})
	if err != nil {
		log.Printf("Could not connect to ws: %v", err)
	}
	defer ws.Close()
	log.Println("Connected to websocket")

	msgChan := make(chan WebsocketMessage)
	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Fatalf("SHIT HIT THE FAN: %v", err)
			}

			if strings.Contains(string(message), "Ratchet") {
				log.Println("Subscribing via websocket...")
				ws.WriteMessage(1, []byte(`[5,"/stops/01346"]`))
			} else {
				var msg WebsocketMessage
				if err := json.Unmarshal(cleanJSON(message), &msg); err != nil {
					fmt.Printf("Could not unmarshal websocket message into struct: %v", err)
				}
				msgChan <- msg
			}
		}
	}()

	go func() {
		for {
			m := <-msgChan

			// Data of interest: line number, final destination (end station)
			v1 := m.Journey.LineNumber == "35" && m.Journey.Destination == "Olof Palmeplein"
			v2 := len(m.Calls) > 0 && m.Calls[0].PlannedArrivalAt != ""
			if v1 == false && v2 == false {
				continue
			}

			// Decide which timestamp to use.
			// Ideally we would use directly the provided "planned" time,
			// which includes delays or advances in the schedule.
			arrivalTime := m.Calls[0].PlannedArrivalAt
			if m.Calls[0].LiveArrivalAt != "" {
				arrivalTime = m.Calls[0].LiveArrivalAt
			}

			loc, _ := time.LoadLocation("Europe/Amsterdam")
			busTime, _ := FromClockToTime(arrivalTime)
			homeTime := time.Now().In(loc)

			// Filter past events.
			if busTime.Sub(homeTime) < 0 {
				h, m, s := busTime.Clock()
				fmt.Printf("Filtered bus in past: %v:%v:%v (%v)\n", h, m, s, busTime.Sub(homeTime))
				continue
			}

			// Filter buses that are too far in the future: 60sec*30=30min
			if busTime.Sub(homeTime).Seconds() > 60*30 {
				h, m, s := busTime.Clock()
				fmt.Printf("Filtered bus too far in the future: %v:%v:%v (%v)\n", h, m, s, busTime.Sub(homeTime))
				continue
			}

			fmt.Println("-------------------->")
			//fmt.Printf("%s\n", b)
			log.Printf("> Next bus 35 arrives in (%s)/(%v)\n", busTime.Sub(homeTime), arrivalTime)
			fmt.Println("-------------------->")
		}
	}()

	return msgChan
}
