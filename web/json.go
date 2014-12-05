package web

import (
	"fmt"
	"bytes"
	"errors"
	"io/ioutil"
	"net/url"
	"net/http"
	"encoding/json"
)

type JsonBody map[string]interface{}

const (
	HeaderContentType = "Content-Type"
)

const (
	ContentTypeJson = "application/json"
	ContentTypeXml = "application/xml"
)

func SendPostJson(url string, jsonBody interface{}, jsonResponse interface{}) error {
	if jsonBytes, err := json.Marshal(jsonBody); err != nil {
		return err
	} else if req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBytes)); err != nil {
		return err
	} else {
		req.Header.Set(HeaderContentType, ContentTypeJson)
		return RoundTripJson(req, &jsonResponse)
	}
}

func SendGetJson(url string, query map[string]interface{}, jsonResponse interface{}) error {
	if queryValues, err := MarshalValues(query); err != nil {
		return err
	} else if req, err := http.NewRequest("GET", fmt.Sprintf("%v?%v",url, queryValues.Encode()), nil); err != nil {
		return err
	} else {
		return RoundTripJson(req, &jsonResponse)
	}
}

type HttpResponseError struct {
	StatusCode int
	Status string
	Body string
}

func (r *HttpResponseError) Error() string {
	return fmt.Sprintf("%v %v", r.Status, r.Body)
}

func NewHttpResponseError(resp *http.Response) *HttpResponseError {
	if bodyBytes, err := ioutil.ReadAll(resp.Body); err == nil {
		return &HttpResponseError{resp.StatusCode, resp.Status, string(bodyBytes)}
	} else {
		return &HttpResponseError{resp.StatusCode, resp.Status, ""}
	}
}

func RoundTripJson (req *http.Request, jsonResponse interface{}) error {
	if response, err := http.DefaultTransport.RoundTrip(req); err != nil {
		return err
	} else if response.StatusCode != http.StatusOK {
		return NewHttpResponseError(response)
	} else if response.Header.Get(HeaderContentType) != ContentTypeJson {
		return errors.New(fmt.Sprintf("Unexpected %v %v", HeaderContentType, response.Header.Get(HeaderContentType)))
	} else {
		return json.NewDecoder(response.Body).Decode(jsonResponse)
	}
}

func SendRequest (req *http.Request) (JsonBody, error) {
	var jsonBody JsonBody
	if err := RoundTripJson(req, &jsonBody); err != nil {
		return nil, err
	} else {
		return jsonBody, nil
	}
}

func MarshalValues(query map[string]interface{}) (url.Values, error) {
	values := url.Values{}
	for k, v := range query {
		switch v := v.(type) {
		case string: values.Add(k, v)
		default: return nil, errors.New(fmt.Sprintf("unhandled type of %v", v))
		}
	}
	return values, nil
}

type Json struct {
	Data interface{} `json:",omitempty"`
	Error Error `json:",omitempty"`
}

type Error struct {
	Code int
	Message string
	Errors []FieldError
}

type FieldError struct {
	Reason string
	Message string
	Location string
	LocationType string
}

