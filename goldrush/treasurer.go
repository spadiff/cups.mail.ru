package main

import "sync"

type Treasure string
type Coin int

type Treasurer struct {
	c *Client
	coins []Coin
	treasuresToCash chan Treasure
	m sync.RWMutex
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

func (t *Treasurer) GetCoins (number int) []Coin {
	t.m.Lock()
	coinsCount := len(t.coins)
	coins := make([]Coin, number)
	copy(coins, t.coins[coinsCount - number:coinsCount - 1])
	t.coins = t.coins[0:coinsCount - number]
	t.m.Unlock()
	return coins
}

func (t *Treasurer) GetCoinsCount () int {
	t.m.RLock()
	defer t.m.RUnlock()
	return len(t.coins)
}

func (t *Treasurer) run () {
	for treasure := range t.treasuresToCash {
		go t.cash(treasure)
	}
}

func (t *Treasurer) Cash(treasure Treasure) {
	t.treasuresToCash <- treasure
}

func NewTreasurer(client *Client) *Treasurer {
	client.SetRPSLimit("cash", 99)
	treasurer := Treasurer{
		c:               client,
		coins:           make([]Coin, 0),
		treasuresToCash: make(chan Treasure, 100000),
	}
	go treasurer.run()
	return &treasurer
}
