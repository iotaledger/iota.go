package nodeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	"github.com/iotaledger/iota.go/v4/api"
)

var (
	// ErrHTTPBadRequest gets returned for 400 bad request HTTP responses.
	ErrHTTPBadRequest = ierrors.New("bad request")
	// ErrHTTPInternalServerError gets returned for 500 internal server error HTTP responses.
	ErrHTTPInternalServerError = ierrors.New("internal server error")
	// ErrHTTPNotFound gets returned for 404 not found error HTTP responses.
	ErrHTTPNotFound = ierrors.New("not found")
	// ErrHTTPUnauthorized gets returned for 401 unauthorized error HTTP responses.
	ErrHTTPUnauthorized = ierrors.New("unauthorized")
	// ErrHTTPUnknownError gets returned for unknown error HTTP responses.
	ErrHTTPUnknownError = ierrors.New("unknown error")
	// ErrHTTPNotImplemented gets returned for 501 not implemented error HTTP responses.
	ErrHTTPNotImplemented = ierrors.New("operation not implemented/supported/available")
	// ErrHTTPServiceUnavailable gets returned for 503 service unavailable error HTTP responses.
	ErrHTTPServiceUnavailable = ierrors.New("service unavailable")

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
	locationHeader = "Location"
)

func readBody(res *http.Response) ([]byte, error) {
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, ierrors.Wrap(err, "unable to read response body")
	}

	return resBody, nil
}

func interpretBody(ctx context.Context, serixAPI *serix.API, res *http.Response, decodeTo interface{}) error {
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

		return serixAPI.JSONDecode(ctx, resBody, decodeTo)
	}

	resBody, err := readBody(res)
	if err != nil {
		return err
	}

	errRes := &HTTPErrorResponseEnvelope{}
	if len(resBody) > 0 {
		if err := json.Unmarshal(resBody, errRes); err != nil {
			return ierrors.Wrap(err, "unable to read error from response body")
		}
	}

	err, ok := httpCodeToErr[res.StatusCode]
	if !ok {
		err = ErrHTTPUnknownError
	}

	return ierrors.WithMessagef(err, "url %s, error message: %s", res.Request.URL.String(), errRes.Error.Message)
}

func do(
	ctx context.Context,
	serixAPI *serix.API,
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
			data, err = serixAPI.JSONEncode(ctx, reqObj)
			if err != nil {
				return nil, ierrors.Wrap(err, "unable to serialize request object to JSON")
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
		return nil, ierrors.Wrap(err, "unable to build http request")
	}

	if userInfo != nil {
		// set the userInfo for basic auth
		req.URL.User = userInfo
	}

	if data != nil {
		if !raw {
			req.Header.Set("Content-Type", api.MIMEApplicationJSON)
		} else {
			req.Header.Set("Content-Type", api.MIMEApplicationVendorIOTASerializerV2)
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
	if err := interpretBody(ctx, serixAPI, res, resObj); err != nil {
		return nil, err
	}

	return res, nil
}

func encodeURLWithQueryParams(endpoint string, queryParams url.Values) (string, error) {
	if len(queryParams) > 0 {
		base, err := url.Parse(endpoint)
		if err != nil {
			return "", ierrors.Wrap(err, "failed to parse endpoint")
		}

		// encode the query params
		base.RawQuery = queryParams.Encode()

		// create a valid URL string
		return base.String(), nil
	}

	return endpoint, nil
}
