package main

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

	requestsCount        int32
	successRequestsCount int32
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

func (e *Explorer) checkPoint(point Point) {
	amount, err := e.checkArea(point, point)
	if err == nil && amount != 0 {
		point.amount = amount
		e.d.Find(point)
	}
}

func (e *Explorer) Run(from, to int) {
	for i := from; i < to; i += 2 {
		for j := 0; j < WIDTH; j += 2 {
			a := Point{x: i, y: j}
			b := Point{x: i + 1, y: j + 1}
			amount, err := e.checkArea(a, b)
			if err == nil && amount != 0 {
				e.checkPoint(Point{x: i, y: j})
				e.checkPoint(Point{x: i + 1, y: j})
				e.checkPoint(Point{x: i, y: j + 1})
				e.checkPoint(Point{x: i + 1, y: j + 1})
			}
		}
	}


	//for i := 90; i < 100; i++ {
	//	for j := 0; j < WIDTH; j++ {
	//		point := Point{x: i, y: j, amount: 1}
	//		e.d.Find(point)
	//	}
	//}

	//for i := 30; i < 90; i++ {
	//	for j := 0; j < WIDTH; j++ {
	//		point := Point{x: i, y: j}
	//		amount, err := e.checkArea(point, point)
	//		if err == nil && amount != 0 {
	//			point.amount = amount
	//			e.d.Find(point)
	//		}
	//	}
	//}
}

func NewExplorer(client *Client, digger *Digger) *Explorer {
	//client.SetRPSLimit("explore", 499)
	return &Explorer{
		c: client,
		d: digger,
	}
}
