package main

import "sync"

type Treasure string
type Coin int

type Treasurer struct {
	c *Client
	coins []Coin
	treasuresToCash chan Treasure
	m sync.Mutex
}

func (t *Treasurer) cash (treasure Treasure) error {
	var coins []Coin
	_, err := t.c.doRequest("cash", &treasure, &coins)

	if err != nil {
		t.treasuresToCash <- treasure
		return err
	}

	t.m.Lock()
	t.coins = append(t.coins, coins...)
	t.m.Unlock()

	return nil
}

func (t *Treasurer) run() {
	for treasure := range t.treasuresToCash {
		go t.cash(treasure)
	}
}

func (t *Treasurer) Cash(treasure Treasure) {
	t.treasuresToCash <- treasure
}

func NewTreasurer(client *Client) *Treasurer {
	treasurer := Treasurer{
		c:               client,
		coins:           make([]Coin, 0),
		treasuresToCash: make(chan Treasure, 100000),
		m:               sync.Mutex{},
	}
	go treasurer.run()
	return &treasurer
}
