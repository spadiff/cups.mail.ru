package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Digger struct {
	c *Client
	l *Licenser
	t *Treasurer
	pointsToFind chan Point
}

func (d *Digger) dig(point Point, depth int, license int) ([]Treasure, error) {
	request := struct {
		X         int `json:"posX"`
		Y         int `json:"posY"`
		Depth     int `json:"depth"`
		LicenseID int `json:"licenseID"`
	}{
		X: point.x, Y: point.y, Depth: depth, LicenseID: license,
	}

	response := make([]Treasure, 0)

	code, err := d.c.doRequest("dig", &request, &response)
	if err != nil && code != 404 {
		if strings.Contains(err.Error(), "no such license") {
			fmt.Println("no license, status code "+strconv.Itoa(code), license)
			return response, nil
		}
		return nil, err
	}

	return response, nil
}

func (d *Digger) run() {
	for point := range d.pointsToFind {
		go func(d *Digger, point Point) {
			for depth := 0; depth < MAX_DEPTH; depth++ {
				license := d.l.GetLicense()
				treasures, err := d.dig(point, depth, license)
				if err != nil {
					d.l.ReturnLicense(license)
					fmt.Println(err)
					return
				}
				for _, treasure := range treasures {
					d.t.Cash(treasure)
				}
			}
		}(d, point)
	}
}

func (d *Digger) Find(point Point) {
	d.pointsToFind <- point
}

func NewDigger(client *Client, licenser *Licenser, treasurer *Treasurer) *Digger {
	digger := Digger{
		c:            client,
		l:            licenser,
		t:            treasurer,
		pointsToFind: make(chan Point, 100000),
	}
	go digger.run()
	return &digger
}
