package nodeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	// ErrHTTPBadRequest gets returned for 400 bad request HTTP responses.
	ErrHTTPBadRequest = errors.New("bad request")
	// ErrHTTPInternalServerError gets returned for 500 internal server error HTTP responses.
	ErrHTTPInternalServerError = errors.New("internal server error")
	// ErrHTTPNotFound gets returned for 404 not found error HTTP responses.
	ErrHTTPNotFound = errors.New("not found")
	// ErrHTTPUnauthorized gets returned for 401 unauthorized error HTTP responses.
	ErrHTTPUnauthorized = errors.New("unauthorized")
	// ErrHTTPUnknownError gets returned for unknown error HTTP responses.
	ErrHTTPUnknownError = errors.New("unknown error")
	// ErrHTTPNotImplemented gets returned for 501 not implemented error HTTP responses.
	ErrHTTPNotImplemented = errors.New("operation not implemented/supported/available")
	// ErrHTTPServiceUnavailable gets returned for 503 service unavailable error HTTP responses.
	ErrHTTPServiceUnavailable = errors.New("service unavailable")

	httpCodeToErr = map[int]error{
		http.StatusBadRequest:          ErrHTTPBadRequest,
		http.StatusInternalServerError: ErrHTTPInternalServerError,
		http.StatusNotFound:            ErrHTTPNotFound,
		http.StatusUnauthorized:        ErrHTTPUnauthorized,
		http.StatusNotImplemented:      ErrHTTPNotImplemented,
		http.StatusServiceUnavailable:  ErrHTTPServiceUnavailable,
	}
)

const (
	MIMEApplicationJSON                   = "application/json"
	MIMEApplicationVendorIOTASerializerV1 = "application/vnd.iota.serializer-v1"
	locationHeader                        = "Location"
)

func readBody(res *http.Response) ([]byte, error) {
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}
	return resBody, nil
}

func interpretBody(res *http.Response, decodeTo interface{}) error {
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		if decodeTo == nil {
			return nil
		}

		resBody, err := readBody(res)
		if err != nil {
			return err
		}

		if rawData, ok := decodeTo.(*RawDataEnvelope); ok {
			rawData.Data = make([]byte, len(resBody))
			copy(rawData.Data, resBody)
			return nil
		}

		return json.Unmarshal(resBody, decodeTo)
	}

	resBody, err := readBody(res)
	if err != nil {
		return err
	}

	errRes := &HTTPErrorResponseEnvelope{}
	if len(resBody) > 0 {
		if err := json.Unmarshal(resBody, errRes); err != nil {
			return fmt.Errorf("unable to read error from response body: %w", err)
		}
	}

	err, ok := httpCodeToErr[res.StatusCode]
	if !ok {
		err = ErrHTTPUnknownError
	}

	return fmt.Errorf("%w: url %s, error message: %s", err, res.Request.URL.String(), errRes.Error.Message)
}

func do(
	ctx context.Context,
	httpClient *http.Client,
	baseURL string,
	userInfo *url.Userinfo,
	method string,
	route string,
	requestURLHook RequestURLHook,
	requestHeaderHook RequestHeaderHook,
	reqObj interface{},
	resObj interface{}) (*http.Response, error) {

	// marshal request object
	var data []byte
	var raw bool

	if reqObj != nil {
		var err error

		if rawData, ok := reqObj.(*RawDataEnvelope); !ok {
			data, err = json.Marshal(reqObj)
			if err != nil {
				return nil, fmt.Errorf("unable to serialize request object to JSON: %w", err)
			}
		} else {
			data = rawData.Data
			raw = true
		}
	}

	// construct request URL
	url := fmt.Sprintf("%s%s", baseURL, route)
	if requestURLHook != nil {
		url = requestURLHook(url)
	}

	// construct request
	req, err := http.NewRequestWithContext(ctx, method, url, func() io.Reader {
		if data == nil {
			return nil
		}
		return bytes.NewReader(data)
	}())
	if err != nil {
		return nil, fmt.Errorf("unable to build http request: %w", err)
	}

	if userInfo != nil {
		// set the userInfo for basic auth
		req.URL.User = userInfo
	}

	if data != nil {
		if !raw {
			req.Header.Set("Content-Type", MIMEApplicationJSON)
		} else {
			req.Header.Set("Content-Type", MIMEApplicationVendorIOTASerializerV1)
		}
	}

	if requestHeaderHook != nil {
		requestHeaderHook(req.Header)
	}

	// make the request
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// write response into response object
	if err := interpretBody(res, resObj); err != nil {
		return nil, err
	}
	return res, nil
}
