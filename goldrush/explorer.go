package main

import (
	"fmt"
	"go.uber.org/ratelimit"
)

const (
	HEIGHT    = 3500
	WIDTH     = 3500
	MAX_DEPTH = 10
)

type Point struct {
	x int
	y int
}

type Explorer struct {
	c *Client
	d *Digger
}

func (e *Explorer) checkArea(a, b Point) (int, error) {
	// TODO: swap points?
	request := struct {
		X     int `json:"posX"`
		Y     int `json:"posY"`
		SizeX int `json:"sizeX"`
		SizeY int `json:"sizeY"`
	}{
		X: a.x, Y: a.y, SizeX: b.x - a.x + 1, SizeY: b.y - a.y + 1,
	}

	response := struct {
		Amount int `json:"amount"`
	}{}

	_, err := e.c.doRequest("explore", &request, &response)

	return response.Amount, err
}

func (e *Explorer) Run() {
	rl := ratelimit.New(1000)

	for i := 0; i < HEIGHT; i++ {
		for j := 0; j < WIDTH; j++ {
			rl.Take()
			point := Point{x: i, y: j}
			go func(e *Explorer, point Point) {
				amount, err := e.checkArea(point, point)
				if err != nil {
					fmt.Println(err)
				} else if amount != 0 {
					e.d.Find(point)
				}
			}(e, point)
		}
	}
}

func NewExplorer(client *Client, digger *Digger) *Explorer {
	return &Explorer{c: client, d: digger}
}
