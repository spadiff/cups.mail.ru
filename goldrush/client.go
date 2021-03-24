package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/ratelimit"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

type Client struct {
	url string
	isAlive int32
	rpsLimiters map[string]ratelimit.Limiter
	measure *Measure
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

func (c *Client) doRequest(method string, request, response interface{}) (int, error) {
	c.measure.Add(method + "_in_progress", 1)
	defer func() {
		c.measure.Add(method + "_in_progress", -1)
	}()

	for c.isAlive != 1 {}
	limiter, ok := c.rpsLimiters[method]
	if ok {
		limiter.Take()
	}

	url := c.url + "/" + method
	data, _ := json.Marshal(&request)

	before := time.Now()
	res, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("unable to do %s request: %w", method, err)
	}
	after := time.Now().Sub(before).Milliseconds()

	defer res.Body.Close()
	defer func(res *http.Response) {
		c.measure.Add(method + "_count", 1)
		c.measure.Add(method + "_timing", after)
		c.measure.Add(method + "_" + strconv.Itoa(res.StatusCode) + "_count", 1)
		c.measure.Add(method + "_" + strconv.Itoa(res.StatusCode) + "_timing", after)
	}(res)

	resData, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("unable to read %s data: %w", method, err)
	}

	if res.StatusCode != 200 {
		err = parseRequestError(resData)
		return res.StatusCode, fmt.Errorf("unable to %s: %w", method, err)
	}

	err = json.Unmarshal(resData, &response)
	if err != nil {
		return 0, fmt.Errorf("unable to parse valid %s json: %w", method, err)
	}

	return res.StatusCode, nil
}

func (c *Client) healthCheck() {
	ticker := time.NewTicker(1 * time.Second)
	for _ = range ticker.C {
		//fmt.Println("kek2")
		//res, err := http.Get(c.url + "/health-check")
		//fmt.Println("kek3")
		//if err != nil || res.StatusCode != 200 {
		//	atomic.CompareAndSwapInt32(&c.isAlive, 1, 0)
		//} else {
		atomic.CompareAndSwapInt32(&c.isAlive, 0, 1)
		//}
	}
}

func (c *Client) SetRPSLimit(method string, rate int) {
	c.rpsLimiters[method] = ratelimit.New(rate)
}

func NewClient() *Client {
	methods := []string{"licenses", "explore", "dig", "cash"}
	measures := make([]string, 0)
	for _, method := range methods {
		measures = append(measures, method + "_in_progress")
		measures = append(measures, method + "_count")
		measures = append(measures, method + "_timing")
		for i := 0; i < 1000; i++ {
			measures = append(measures, method + "_" + strconv.Itoa(i) + "_count")
			measures = append(measures, method + "_" + strconv.Itoa(i) + "_timing")
		}
	}


	address := os.Getenv("ADDRESS")
	client := Client{
		url:         "http://" + address + ":8000",
		isAlive:     0,
		rpsLimiters: make(map[string]ratelimit.Limiter),
		measure: NewMeasure(measures),
	}
	go client.healthCheck()
	return &client
}
