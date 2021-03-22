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

func (d *Digger) Find(point Point, depth int) error {
	license := d.l.GetLicense()
	treasures, err := d.dig(point, depth, license)
	if err != nil {
		d.l.ReturnLicense(license)
		return err
	}
	for _, treasure := range treasures {
		d.t.Cash(treasure)
	}
	return nil
}

func NewDigger(client *Client, licenser *Licenser, treasurer *Treasurer) *Digger {
	return &Digger{c: client, l: licenser, t: treasurer}
}
