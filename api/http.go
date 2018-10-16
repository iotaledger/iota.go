package api

import (
	"bytes"
	"encoding/json"
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/pow"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

const DefaultLocalIRIURI = "http://localhost:14265"

func NewHttpClient(settings interface{}) (*httpclient, error) {
	httpClient := &httpclient{}
	if err := httpClient.SetSettings(settings); err != nil {
		return nil, err
	}
	return httpClient, nil
}

type HttpClientSettings struct {
	URI          string
	Client       HttpClient
	LocalPowFunc pow.PowFunc
}

func (hcs HttpClientSettings) PowFunc() pow.PowFunc {
	return hcs.LocalPowFunc
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type httpclient struct {
	client   HttpClient
	endpoint string
}

func (hc *httpclient) SetSettings(settings interface{}) error {
	httpSettings, ok := settings.(HttpClientSettings)
	if !ok {
		return errors.Wrapf(ErrInvalidSettingsType, "expected %T", HttpClientSettings{})
	}
	if len(httpSettings.URI) == 0 {
		hc.endpoint = DefaultLocalIRIURI
	} else {
		if _, err := url.Parse(httpSettings.URI); err != nil {
			return errors.Wrap(ErrInvalidURI, httpSettings.URI)
		}
		hc.endpoint = httpSettings.URI
	}
	if httpSettings.Client != nil {
		hc.client = httpSettings.Client
	}else{
		hc.client = http.DefaultClient
	}
	return nil
}

func (hc *httpclient) Send(cmd interface{}, out interface{}) error {
	b, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	rd := bytes.NewReader(b)
	req, err := http.NewRequest("POST", hc.endpoint, rd)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-IOTA-API-Version", "1")
	resp, err := hc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		errResp := &ErrorResponse{}
		err = json.Unmarshal(bs, errResp)
		return handleError(errResp, err, errors.Wrapf(ErrNonOKStatusCodeFromAPIRequest, "http code %d", resp.StatusCode))
	}

	if bytes.Contains(bs, []byte(`"error"`)) || bytes.Contains(bs, []byte(`"exception"`)) {
		errResp := &ErrorResponse{}
		err = json.Unmarshal(bs, errResp)
		return handleError(errResp, err, ErrUnknownErrorFromAPIRequest)
	}

	if out == nil {
		return nil
	}
	return json.Unmarshal(bs, out)
}

func handleError(err *ErrorResponse, err1, err2 error) error {
	switch {
	case err.Error != "":
		return errors.New(err.Error)
	case err.Exception != "":
		return errors.New(err.Exception)
	case err1 != nil:
		return err1
	}

	return err2
}
