package main

import (
	"strconv"
)

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
	c       *Client
	d       *Digger
	measure *Measure
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
	if err == nil {
		area := strconv.Itoa(b.y - a.y + 1)
		e.measure.Add(area+"_count", 1)
		if response.Amount != 0 {
			e.measure.Add(area+"_not_empty", 1)
			e.measure.Add(area+"_sum", int32(response.Amount))
		} else {
			e.measure.Add(area+"_empty", 1)
		}
	}

	return response.Amount, err
}

func (e *Explorer) checkPoint(point Point, amount int) (int, error) {
	var err error
	if amount == -1 {
		amount, err = e.getAreaAmount(point, point)
		for err != nil {
			amount, err = e.getAreaAmount(point, point)
		}
	}
	if amount != 0 {
		point.amount = amount
		e.d.Find(point)
	}
	return amount, nil
}

func (e *Explorer) checkBinArea(a, b Point, amount int) (int, error) {
	if a == b {
		amount, err := e.checkPoint(a, amount)
		return amount, err
	}

	if amount == -1 {
		var err error
		amount, err = e.getAreaAmount(a, b)
		for err != nil {
			amount, err = e.getAreaAmount(a, b)
		}
	}

	if amount == 0 {
		return 0, nil
	}

	c := Point{x: b.x, y: (a.y + b.y) / 2}
	amount1, _ := e.checkBinArea(a, c, -1)
	if amount1 != amount {
		c.y += 1
		_, _ = e.checkBinArea(c, b, amount-amount1)
	}
	return amount, nil
}

func (e *Explorer) checkArea(a, b Point) error {
	amount, err := e.getAreaAmount(a, b)
	if err != nil {
		return err
	}

	if amount != 0 {
		for i := a.x; i <= b.x; i++ {
			for j := a.y; j <= b.y; j++ {
				pointAmount, err := e.checkPoint(Point{x: i, y: j}, -1)
				if err != nil {
					continue
				}
				amount -= pointAmount
				if amount == 0 {
					return nil
				}
			}
		}
	}

	return nil
}

//func (e *Explorer) Run(from, to, size int) {
//	for i := from; i < to; i += size {
//		for j := 0; j < WIDTH; j += size {
//			a := Point{x: i, y: j}
//			b := Point{x: i + size - 1, y: j + size - 1}
//			_ = e.checkArea(a, b)
//		}
//	}
//}

func (e *Explorer) Run(from, to, width int) {
	for i := from; i < to; i++ {
		for j := 0; j < WIDTH-width+1; j += width {
			a := Point{x: i, y: j}
			b := Point{x: i, y: j + width - 1}
			_, _ = e.checkBinArea(a, b, -1)
		}
	}
}

func NewExplorer(client *Client, digger *Digger) *Explorer {
	//client.SetRPSLimit("explore", 499)
	measures := make([]string, 0)
	for i := 0; i <= 3500; i++ {
		measures = append(measures, strconv.Itoa(i)+"_count")
		measures = append(measures, strconv.Itoa(i)+"_empty")
		measures = append(measures, strconv.Itoa(i)+"_not_empty")
		measures = append(measures, strconv.Itoa(i)+"_sum")
	}

	return &Explorer{
		c:       client,
		d:       digger,
		measure: NewMeasure(measures),
	}
}
