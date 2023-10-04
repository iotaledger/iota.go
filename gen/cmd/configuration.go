package main

import (
	"fmt"
	"strings"
)

type configuration struct {
	FileName string
	Name     string
	Receiver string
	Features map[string]bool
}

func newConfiguration(fileName, name, receiver, featuresStr string) *configuration {
	features := make(map[string]bool)
	for _, feature := range strings.Split(featuresStr, " ") {
		features[feature] = true
	}

	return &configuration{
		FileName: fileName,
		Name:     name,
		Receiver: strings.ToLower(receiver),
		Features: features,
	}
}

func (c *configuration) String() string {
	return fmt.Sprintf("configuration{\n\tFileName:%s,\n\tName:%s,\n\tReceiver:%s,\n\tFeatures:%v,\n}\n", c.FileName, c.Name, c.Receiver, c.Features)
}
