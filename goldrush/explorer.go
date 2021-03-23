package main

import "sync/atomic"

const (
	HEIGHT    = 3500
	WIDTH     = 3500
	MAX_DEPTH = 10
)

type Point struct {
	x      int
	y      int
	amount int
}

type Explorer struct {
	c *Client
	d *Digger

	emptyAreasCount int32
	areasCount int32
}

func (e *Explorer) getAreaAmount(a, b Point) (int, error) {
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

func (e *Explorer) checkPoint(point Point) (int, error) {
	amount, err := e.getAreaAmount(point, point)
	if err != nil {
		return 0, err
	}
	if amount != 0 {
		point.amount = amount
		e.d.Find(point)
	}
	return amount, nil
}

func (e *Explorer) checkArea(a Point, b Point) error {
	amount, err := e.getAreaAmount(a, b)
	if err != nil {
		return err
	}

	atomic.AddInt32(&e.areasCount, 1)

	if amount != 0 {
		for i := a.x; i <= b.x; i++ {
			for j := a.y; j <= b.y; j++ {
				pointAmount, err := e.checkPoint(Point{x: i, y: j})
				if err != nil {
					continue
				}
				amount -= pointAmount
				if amount == 0 {
					return nil
				}
			}
		}
	} else {
		atomic.AddInt32(&e.emptyAreasCount, 1)
	}

	return nil
}

func (e *Explorer) Run(from, to, size int) {
	for i := from; i < to; i += size {
		for j := 0; j < WIDTH; j += size {
			a := Point{x: i, y: j}
			b := Point{x: i + size - 1, y: j + size - 1}
			_ = e.checkArea(a, b)
		}
	}
}

func NewExplorer(client *Client, digger *Digger) *Explorer {
	//client.SetRPSLimit("explore", 499)
	return &Explorer{
		c: client,
		d: digger,
	}
}
