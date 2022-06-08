package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type httpRequest struct {
	Host          string
	Client        *http.Client
	Context       context.Context
	baseURL       *url.URL
	RequestHeader http.Header

	LastResponseStatus int
	LastResponseHeader http.Header
	LastResponseBody   []byte
	lastResponse       *http.Response
}

type Config struct {
}

func LoadConfig(jsonPath string) (*Config, error) {
	file, err := openFile(jsonPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	configBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	c := new(Config)
	err = json.Unmarshal(configBytes, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewHttpRequest(client *http.Client) (*httpRequest, error) {

	req := newRequest(client)
	req.baseURL.Scheme = "http"

	return req, nil
}

func newRequest(client *http.Client) *httpRequest {
	req := new(httpRequest)
	req.baseURL = new(url.URL)
	req.baseURL.Host = req.Host
	req.baseURL.Path = "/"

	if client == nil {
		client = new(http.Client)
	}
	req.Client = client
	req.Context = context.TODO()
	return req
}

func VerifyHTTPCode(code int) bool {
	if code < http.StatusOK || code > http.StatusIMUsed {
		return false
	}
	return true
}

func (u *httpRequest) putMessageUsedGetMethod(reqURL string) error {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}
	return u.request(req)
}

func (u *httpRequest) request(req *http.Request) error {
	resp, err := u.requestWithResp(req)
	if err != nil {
		return err
	}

	err = u.responseParse(resp)
	if err != nil {
		return err
	}

	if !VerifyHTTPCode(resp.StatusCode) {
		return fmt.Errorf("Remote response code is %d - %s not 2xx call DumpResponse(true) show details",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}

func (u *httpRequest) requestWithResp(req *http.Request) (resp *http.Response, err error) {
	req.Header.Set("User-Agent", "router/1.0")

	resp, err = u.Client.Do(req.WithContext(u.Context))
	// If we got an error, and the context has been canceled,
	// the context's error is probably more useful.
	if err != nil {
		select {
		case <-u.Context.Done():
			err = u.Context.Err()
		default:
		}
		return
	}
	return
}

func (u *httpRequest) responseParse(resp *http.Response) error {
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	u.LastResponseStatus = resp.StatusCode
	u.LastResponseHeader = resp.Header
	u.LastResponseBody = resBody
	u.lastResponse = resp
	return nil
}

func (u *httpRequest) DumpResponse(isDumpBody bool) []byte {
	var b bytes.Buffer
	if u.lastResponse == nil {
		return nil
	}
	b.WriteString(fmt.Sprintf("%s %d\n", u.lastResponse.Proto, u.LastResponseStatus))
	for k, vs := range u.LastResponseHeader {
		str := k + ": "
		for i, v := range vs {
			if i != 0 {
				str += "; " + v
			} else {
				str += v
			}
		}
		b.WriteString(str)
	}
	if isDumpBody {
		b.Write(u.LastResponseBody)
	}
	return b.Bytes()
}

func openFile(path string) (*os.File, error) {
	return os.Open(path)
}
