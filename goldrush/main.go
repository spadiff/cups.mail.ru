package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	HEIGHT = 3500
	WIDTH  = 3500
	DEPTH  = 10
)

type Treasure string

type Coin int

type Point struct {
	x            int
	y            int
	currentDepth int
	amount       int
}

type License struct {
	id                int
	remainingAttempts int
}

type GoldRusher struct {
	url string
}

func parseRequestError(data []byte) error {
	var errorResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	err := json.Unmarshal(data, &errorResponse)
	if err != nil {
		return fmt.Errorf("unable to parse error json: %w", err)
	}
	return fmt.Errorf(errorResponse.Message)
}

func (g *GoldRusher) explore(point *Point, sizeX, sizeY int) (int, error) {
	reqData := struct {
		X     int `json:"posX"`
		Y     int `json:"posY"`
		SizeX int `json:"sizeX"`
		SizeY int `json:"sizeY"`
	}{
		X: point.x, Y: point.y, SizeX: sizeX, SizeY: sizeY,
	}
	reqDataBytes, _ := json.Marshal(&reqData)

	res, err := http.Post(g.url+"/explore", "application/json", bytes.NewReader(reqDataBytes))
	if err != nil {
		return 0, fmt.Errorf("unable to do explore request: %w", err)
	}
	defer res.Body.Close()

	resDataBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("unable to read explore data: %w", err)
	}

	if res.StatusCode != 200 {
		err = parseRequestError(resDataBytes)
		return 0, fmt.Errorf("unable to explore %d: %w", res.StatusCode, err)
	}

	var validResponse struct {
		Amount int `json:"amount"`
	}
	err = json.Unmarshal(resDataBytes, &validResponse)
	if err != nil {
		return 0, fmt.Errorf("unable to parse valid explore json: %w", err)
	}

	return validResponse.Amount, nil
}

func (g *GoldRusher) createLicense() (License, error) {
	res, err := http.Post(g.url+"/licenses", "application/json", bytes.NewReader([]byte("[]")))
	if err != nil {
		return License{}, fmt.Errorf("unable to do license request: %w", err)
	}
	defer res.Body.Close()

	resDataBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return License{}, fmt.Errorf("unable to read explore data: %w", err)
	}

	if res.StatusCode != 200 {
		err = parseRequestError(resDataBytes)
		return License{}, fmt.Errorf("unable to create license %d: %w", res.StatusCode, err)
	}

	var validResponse struct {
		ID         int `json:"id"`
		DigAllowed int `json:"digAllowed"`
		DigUsed    int `json:"digUsed"` // TODO: is it really useless?
	}
	err = json.Unmarshal(resDataBytes, &validResponse)
	if err != nil {
		return License{}, fmt.Errorf("unable to parse valid license json: %w", err)
	}

	return License{
		id:                validResponse.ID,
		remainingAttempts: validResponse.DigAllowed - validResponse.DigUsed,
	}, nil
}

func (g *GoldRusher) dig(point *Point, license *License) ([]Treasure, error) {
	reqData := struct {
		X         int `json:"posX"`
		Y         int `json:"posY"`
		Depth     int `json:"depth"`
		LicenseID int `json:"licenseID"`
	}{
		X: point.x, Y: point.y, Depth: point.currentDepth + 1, LicenseID: license.id,
	}
	reqDataBytes, _ := json.Marshal(&reqData)

	res, err := http.Post(g.url+"/dig", "application/json", bytes.NewReader(reqDataBytes))
	if err != nil {
		return nil, fmt.Errorf("unable to do dig request: %w", err)
	}
	defer res.Body.Close()

	resDataBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read dig data: %w", err)
	}

	if res.StatusCode == 404 {
		return make([]Treasure, 0), nil
	}

	if res.StatusCode != 200 {
		err = parseRequestError(resDataBytes)

	    if strings.Contains(err.Error(), "no such license") {
			fmt.Println("no license, status code "+strconv.Itoa(res.StatusCode), license)
			return make([]Treasure, 0), nil
		}

		return nil, fmt.Errorf("unable to dig %d: %w", res.StatusCode, err)
	}

	var validResponse []Treasure
	err = json.Unmarshal(resDataBytes, &validResponse)
	if err != nil {
		return nil, fmt.Errorf("unable to parse valid dig json: %w", err)
	}

	return validResponse, nil
}

func (g *GoldRusher) cash(t Treasure) ([]Coin, error) {
	fmt.Printf("treasure %v, %v", t, []byte(t))
	res, err := http.Post(g.url+"/cash", "application/json", bytes.NewReader([]byte("\""+t+"\"")))
	if err != nil {
		return nil, fmt.Errorf("unable to do cash request: %w", err)
	}
	defer res.Body.Close()

	resDataBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read cash data: %w", err)
	}

	if res.StatusCode != 200 {
		err = parseRequestError(resDataBytes)
		return nil, fmt.Errorf("unable to cash %d: %w", res.StatusCode, err)
	}

	var validResponse []Coin
	err = json.Unmarshal(resDataBytes, &validResponse)
	if err != nil {
		return nil, fmt.Errorf("unable to parse valid cash json: %w", err)
	}

	return validResponse, nil
}

func main() {
	address := os.Getenv("ADDRESS")
	rusher := GoldRusher{url: "http://" + address + ":8000"}

	licenses := make([]License, 0)

	newLicense, err := rusher.createLicense()
	for err != nil {
		fmt.Println(err)
		time.Sleep(1 * time.Second)
		newLicense, err = rusher.createLicense()
	}
	licenses = append(licenses, newLicense)

	ctx, _ := context.WithTimeout(context.Background(), 5 * time.Minute)
	done := make(chan struct{}, 0)
	isDone := false
	sum := 0

	points := make([]Point, 0)
	for i := 0; i < HEIGHT; i++ {
		for j := 0; j < WIDTH; j++ {
			point := Point{x: i, y: j}
			go func(point *Point, done chan struct{}) {
				amount, _ := rusher.explore(point, 1, 1)
				point.amount = amount
				done <- struct{}{}
			}(&point, done)


			select {
			case <- ctx.Done():
				isDone = true
				break
			case <- done:
				if point.amount != 0 {
					points = append(points, point)
					sum += point.amount
				}
				continue
			}
		}
		if isDone {
			break
		}
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].amount > points[j].amount
	})

	fmt.Println(len(points))
	fmt.Println(sum)
	fmt.Println(points[:20])

	for _, point := range points {
		fmt.Println(point)
		for point.currentDepth < DEPTH {
			lastLicense := &licenses[len(licenses)-1]
			if lastLicense.remainingAttempts == 0 {
				newLicense, err := rusher.createLicense()
				for err != nil {
					newLicense, err = rusher.createLicense()
				}
				licenses = append(licenses, newLicense)
				lastLicense = &licenses[len(licenses)-1]
			}

			treasures, err := rusher.dig(&point, lastLicense)
			if err != nil {
				fmt.Println(err)
				break
			}

			point.currentDepth += 1
			lastLicense.remainingAttempts -= 1

			for _, treasure := range treasures {
				kek, err := rusher.cash(treasure)
				fmt.Println("coins:" + strconv.Itoa(len(kek)))
				fmt.Println(err)
			}

		}
	}
}
