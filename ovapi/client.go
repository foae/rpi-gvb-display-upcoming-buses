package ovapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type OVAPIResponse struct {
	Station struct {
		DirectionOlofPalmeplein struct {
			Stop struct {
				Longitude                       float64 `json:"Longitude"`
				Latitude                        float64 `json:"Latitude"`
				TimingPointTown                 string  `json:"TimingPointTown"`
				TimingPointName                 string  `json:"TimingPointName"`
				TimingPointCode                 string  `json:"TimingPointCode"`
				StopAreaCode                    string  `json:"StopAreaCode"`
				TimingPointWheelChairAccessible string  `json:"TimingPointWheelChairAccessible"`
				TimingPointVisualAccessible     string  `json:"TimingPointVisualAccessible"`
			} `json:"Stop"`
			Passes map[string]bus `json:"Passes"`
		} `json:"30001346"`
	} `json:"01346"`
}

type bus struct {
	// Grouped data of interest
	TargetArrivalTime   string `json:"TargetArrivalTime"`   // convert to time: 2018-11-17T17:22:16; – planned arrival time
	TargetDepartureTime string `json:"TargetDepartureTime"` // same ^
	ExpectedArrivalTime string `json:"ExpectedArrivalTime"` // same ^ – compiled arrival time
	LastUpdateTimeStamp string `json:"LastUpdateTimeStamp"` // same ^

	LinePublicNumber  string `json:"LinePublicNumber"`  // 35
	LineDirection     int    `json:"LineDirection"`     // 2
	DestinationName50 string `json:"DestinationName50"` // Olof Palmeplein
	DestinationCode   string `json:"DestinationCode"`   // OLPP
	TripStopStatus    string `json:"TripStopStatus"`    // PLANNED, DRIVING

	IsTimingStop          bool    `json:"IsTimingStop"`
	DataOwnerCode         string  `json:"DataOwnerCode"`
	OperatorCode          string  `json:"OperatorCode"`
	FortifyOrderNumber    int     `json:"FortifyOrderNumber"`
	TransportType         string  `json:"TransportType"`
	Latitude              float64 `json:"Latitude"`
	Longitude             float64 `json:"Longitude"`
	JourneyNumber         int     `json:"JourneyNumber"`
	JourneyPatternCode    int     `json:"JourneyPatternCode"`
	LocalServiceLevelCode int     `json:"LocalServiceLevelCode"`

	OperationDate                   string `json:"OperationDate"`
	TimingPointCode                 string `json:"TimingPointCode"`
	WheelChairAccessible            string `json:"WheelChairAccessible"`
	LineName                        string `json:"LineName"`
	ExpectedDepartureTime           string `json:"ExpectedDepartureTime"`
	UserStopOrderNumber             int    `json:"UserStopOrderNumber"`
	ProductFormulaType              string `json:"ProductFormulaType"`
	TimingPointName                 string `json:"TimingPointName"`
	LinePlanningNumber              string `json:"LinePlanningNumber"`
	StopAreaCode                    string `json:"StopAreaCode"`
	TimingPointDataOwnerCode        string `json:"TimingPointDataOwnerCode"`
	TimingPointTown                 string `json:"TimingPointTown"`
	UserStopCode                    string `json:"UserStopCode"`
	JourneyStopType                 string `json:"JourneyStopType"`
	TimingPointWheelChairAccessible string `json:"TimingPointWheelChairAccessible"`
	TimingPointVisualAccessible     string `json:"TimingPointVisualAccessible"`
}

func RequestDataFromOV() (*OVAPIResponse, error) {
	req, err := http.NewRequest("GET", "https://v0.ovapi.nl/stopareacode/01346/departures", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.102 Safari/537.36")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ov OVAPIResponse
	if err := json.Unmarshal(b, &ov); err != nil {
		fmt.Printf("could not unmarshal json: %v", err)
	}

	return &ov, nil
}
