package web

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type JsonBody map[string]interface{}

const (
	HeaderContentType               = "Content-Type"
	HeaderContentEncoding           = "Content-Encoding"
	HeaderAccessControlAllowOrigin  = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders = "Access-Control-Allow-Headers"
)

const (
	ContentTypePlain = "text/plain"
	ContentTypeJson  = "application/json"
	ContentTypeXml   = "application/xml"
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
	} else if req, err := http.NewRequest("GET", fmt.Sprintf("%v?%v", url, queryValues.Encode()), nil); err != nil {
		return err
	} else {
		return RoundTripJson(req, &jsonResponse)
	}
}

type HttpResponseError struct {
	StatusCode int
	Status     string
	Body       string
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

func RoundTripJson(req *http.Request, jsonResponse interface{}) error {
	if response, err := http.DefaultTransport.RoundTrip(req); err != nil {
		return err
	} else if response.StatusCode != http.StatusOK {
		return NewHttpResponseError(response)
	} else if response.Header.Get(HeaderContentType) != ContentTypeJson {
		return fmt.Errorf("Unexpected %v %v", HeaderContentType, response.Header.Get(HeaderContentType))
	} else {
		return json.NewDecoder(response.Body).Decode(jsonResponse)
	}
}

func SendRequest(req *http.Request) (JsonBody, error) {
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
		case string:
			values.Add(k, v)
		default:
			return nil, fmt.Errorf("unhandled type of %v", v)
		}
	}
	return values, nil
}

type Json struct {
	Data  interface{} `json:"data,omitempty"`
	Error Error       `json:"error,omitempty"`
}

type Error interface {
	Code() int
	Error() string
	Errors() []ErrorItem
	AddError(item ErrorItem)
}

type Err struct {
	Code_    int         `json:"code"`
	Message_ string      `json:"message"`
	Errors_  []ErrorItem `json:"errors,omitempty"`
}

func (e *Err) AddError(i ErrorItem) {
	e.Errors_ = append(e.Errors_, i)
}

func (e *Err) Code() int {
	return e.Code_
}

func (e *Err) Error() string {
	return e.Message_
}

func (e *Err) Errors() []ErrorItem {
	return e.Errors_
}

func NewError(code int, message string) Error {
	return &Err{code, message, nil}
}

func SimpleError(err error) Error {
	return &Err{http.StatusInternalServerError, err.Error(), nil}
}

type ErrorItem struct {
	Reason       string
	Message      string
	Location     string
	LocationType string
}

func NewErrorItem(reason, message, location, locationType string) ErrorItem {
	return ErrorItem{reason, message, location, locationType}
}

func WriteJsonError(w http.ResponseWriter, err error) {
	WriteJsonWebError(w, SimpleError(err))
}

func WriteJsonErrorWithCode(w http.ResponseWriter, err error, statusCode int) {
	WriteJsonWebError(w, &Err{statusCode, err.Error(), nil})
}

func WriteCorsOptionResponse(w http.ResponseWriter, method string) {
	w.Header().Add(HeaderAccessControlAllowOrigin, "*")
	w.Header().Add(HeaderAccessControlAllowMethods, fmt.Sprintf("%v,OPTIONS", method))
	w.Header().Add(HeaderAccessControlAllowHeaders, HeaderContentType)
	w.WriteHeader(http.StatusOK)
}

func WriteJsonWebError(w http.ResponseWriter, err Error) {
	log.Println(err)
	if bs, e2 := json.Marshal(Json{nil, err}); e2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Add(HeaderContentType, ContentTypeJson)
		w.Header().Add(HeaderAccessControlAllowOrigin, "*")
		w.WriteHeader(err.Code())
		if _, e3 := w.Write(bs); e3 != nil {
			log.Println("Failed to send error response:", e3)
		}
	}
}

func WriteJson(w http.ResponseWriter, data interface{}) {
	if bs, err := json.Marshal(Json{data, nil}); err != nil {
		WriteJsonError(w, err)
	} else {
		w.Header().Add(HeaderContentType, ContentTypeJson)
		w.Header().Add(HeaderAccessControlAllowOrigin, "*")
		if _, err := w.Write(bs); err != nil {
			log.Println("Failed to send response: ", err)
		}
	}
}

func WriteRawGzipJson(w http.ResponseWriter, bs []byte) {
	w.Header().Add(HeaderContentType, ContentTypeJson)
	w.Header().Add(HeaderAccessControlAllowOrigin, "*")
	w.Header().Add(HeaderContentEncoding, "gzip")
	if _, err := w.Write(bs); err != nil {
		log.Println("Failed to send response: ", err)
	}
}

func WriteGzipJson(w http.ResponseWriter, data interface{}) {
	w.Header().Add(HeaderContentType, ContentTypeJson)
	w.Header().Add(HeaderAccessControlAllowOrigin, "*")
	w.Header().Add(HeaderContentEncoding, "gzip")
	gz := gzip.NewWriter(w)
	if err := json.NewEncoder(gz).Encode(Json{data, nil}); err != nil {
		log.Println("Failed to send response: ", err)
	}
	if err := gz.Close(); err != nil {
		log.Println("Failed to close gzip writer", err)
	}
}

func WriteXml(w http.ResponseWriter, bs []byte) {
	w.Header().Add(HeaderContentType, ContentTypeXml)
	w.Header().Add(HeaderAccessControlAllowOrigin, "*")
	if _, err := w.Write(bs); err != nil {
		log.Println("failed to send response: ", err)
	}
}
