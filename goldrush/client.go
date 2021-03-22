package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Client struct {
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

func (c *Client) doRequest(method string, request, response interface{}) (int, error) {
	url := c.url + "/" + method
	data, _ := json.Marshal(&request)

	res, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("unable to do %s request: %w", method, err)
	}

	defer res.Body.Close()
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

func NewClient() *Client {
	address := os.Getenv("ADDRESS")
	return &Client{url: "http://" + address + ":8000"}
}
