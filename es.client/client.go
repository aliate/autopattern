package es

import (
	"io"
	"strings"
	"net/http"
	"io/ioutil"
)

type Client struct {
	Host	string
}

func NewClient(host string) *Client {
	return &Client{
		Host: "http://" + host,
	}
}

func (c *Client) Do(method, path string, data io.Reader) ([]byte, error) {
	client := http.Client{}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	req, err := http.NewRequest(method, c.Host + path, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-type", "application/json")	

	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) Get(path string) ([]byte, error) {
	return c.Do(http.MethodGet, path, nil)
}

func (c *Client) Put(path string) ([]byte, error) {
	return c.Do(http.MethodPut, path, nil)
}

func (c *Client) Post(path string, data io.Reader) ([]byte, error) {
	return c.Do(http.MethodPost, path, data)
}

func (c *Client) Delete(path string) ([]byte, error) {
	return c.Do(http.MethodDelete, path, nil)
}
