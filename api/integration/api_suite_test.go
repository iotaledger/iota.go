package integration_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)


/*
type mockprovider struct {
	res interface{}
}

func (m *mockprovider) Send(cmd interface{}, out interface{}) error {
	if m.res == nil {
		return nil
	}
	if err, ok := m.res.(error); ok {
		return err
	}
	// need to unmarshal res into out since out is only a pointer
	// and we can't redirect it to simply another value
	return json.Unmarshal(m.res.([]byte), out)
}

func (m *mockprovider) SetSettings(settings interface{}) error {
	return nil
}

func mockedProvider(res interface{}) CreateProviderFunc {
	var mock *mockprovider
	if _, ok := res.(error); ok {
		mock = &mockprovider{res: res}
	} else {
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			panic(err)
		}
		mock = &mockprovider{res: jsonBytes}
	}

	return func(settings interface{}) (Provider, error) {
		return mock, nil
	}
}

func mockedAPI(res interface{}) *API {
	return &API{attachToTangle: attachToTangle, provider: provider}, nil
}
*/

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Suite")
}
