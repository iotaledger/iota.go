// Package gen/cmd allows to generate .go source files using the go:generate command.
// Use with "cmd [templateFile] [outputFile] [typeName] [typeReceiver] [Features separated by space]".
//
// In the template the following are available:
// - typeName as {{.Name}}
// - typeReceiver as {{.Receiver}}
// - Features are available via map and can be evaluated as {{if index .Features "extended"}}
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"

	"github.com/iotaledger/hive.go/lo"
)

func main() {
	if len(os.Args) < 6 {
		printUsage("not enough parameters")
	}

	templateFilePath := os.Args[1]
	conf := newConfiguration(os.Args[2], os.Args[3], os.Args[4], os.Args[5])

	funcs := template.FuncMap{
		"firstLower": func(s string) string {
			return strings.ToLower(s[0:]) + s[:1]
		},
	}

	tmplFile := lo.PanicOnErr(os.ReadFile(templateFilePath))

	tmpl := template.Must(
		template.New("gen").
			Funcs(funcs).
			Parse(string(tmplFile)),
	)

	buffer := new(bytes.Buffer)
	panicOnError(tmpl.Execute(buffer, conf))

	formattedOutput := lo.PanicOnErr(format.Source(buffer.Bytes()))

	panicOnError(os.WriteFile(conf.FileName, formattedOutput, 0600))
}

// printUsage prints the usage of the variadic code generator in case of an error.
func printUsage(errorMsg string) {
	_, _ = fmt.Fprintf(os.Stderr, "Error:\t%s\n\n", errorMsg)
	_, _ = fmt.Fprintf(os.Stderr, "Usage of gen/cmd:\n")
	_, _ = fmt.Fprintf(os.Stderr, "\tcmd [templateFile] [outputFile] [typeName] [typeReceiver] [features separated by space]\n")

	os.Exit(2)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
