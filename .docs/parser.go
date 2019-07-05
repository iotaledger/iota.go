// Parser generates markdown documentation according to the IOTA documentation portal format.
// The parser will automatically pick up examples for functions from the "example" subfolder
// in the corresponding packages.
package main

import (
	"flag"
	"fmt"
	"github.com/apsdehal/go-logger"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"unicode"
)

type function struct {
	Name        string
	Title       string
	Desc        string
	Inputs      []input
	Outputs     []output
	ExampleCode string
	Package     string
	ForType     string
	hadExample  bool
}

type input struct {
	ArgName  string
	Type     string
	Desc     string
	Required bool
}

type output struct {
	Type string
	Desc string
}

var packageDirs = []string{
	//"../address",
	//"../api",
	//"../bundle",
	//"../checksum",
	//"../converter",
	//"../curl",
	//"../guards",
	//"../kerl",
	//"../pow",
	//"../signing",
	//"../transaction",
	//"../units",
	//"../account",
	"../account/store",
}

var verbose = flag.Bool("v", false, "")
var writeMarkdown = flag.Bool("w", false, "")
var emptyBodyWarning = flag.Bool("emptyBody", false, "")

var log *logger.Logger
var debugLog *logger.Logger
var logFormat = "%{file}:%{line} > %{lvl} %{message}"

func main() {
	flag.Parse()

	log, _ = logger.New("parser", 1, os.Stdout)
	debugLog, _ = logger.New("debug", 1, os.Stdout)

	log.SetFormat(logFormat)
	debugLog.SetFormat(logFormat)

	debugLog.SetLogLevel(logger.CriticalLevel)
	if *verbose {
		debugLog.SetLogLevel(logger.DebugLevel)
	}

	tmpl := template.Must(template.ParseFiles("template.md"))
	_ = tmpl

	// iterate over all defined packages and generate their documentation
	for _, packageDir := range packageDirs {
		set := token.NewFileSet()
		packages, err := parser.ParseDir(set, packageDir, nil, parser.ParseComments)
		if err != nil {
			log.Fatalf("failed to parse package: %s", err)
		}

		for packageName, astPack := range packages {
			// parse functions of package
			if strings.Contains(packageName, "test") && !strings.Contains(packageName, "example") {
				continue
			}

			// bread and butter
			functions := parsePackage(astPack, set)
			parseExamples(packageName, packageDir, functions)

			for _, fun := range functions {
				if fun.hadExample {
					continue
				}
				log.Warningf("missing example for function %s", fun.Name)
			}

			if *writeMarkdown {
				writeDocs(functions, tmpl)
			}
		}
	}
}

func refFileName(f *function) string {

	var name string
	for i, c := range f.Name {
		if i != 0 && unicode.ToUpper(c) == c {
			name += "_"
		}
		name += string(unicode.ToLower(c))
	}

	return fmt.Sprintf("%s_%s.md", strings.ToLower(f.Package), name)
}

const path = "./iota.go/reference/"

func writeDocs(functions map[string]*function, tmpl *template.Template) {
	for _, fun := range functions {
		fileName := path + refFileName(fun)
		// ignore error
		os.Remove(fileName)
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("unable to create ref markdown file %s", fileName)
		}
		if err := tmpl.Execute(file, fun); err != nil {
			log.Fatalf("unable to write to ref markdown file %s", fileName)
		}
		file.Close()
	}
}

// parses the given AST package and returns its functions
func parsePackage(astPack *ast.Package, set *token.FileSet) map[string]*function {
	pack := doc.New(astPack, "", 1)
	log.Infof("parsing docs for package %s...", astPack.Name)
	functions := make(map[string]*function, len(pack.Funcs))

	funcsTotal := 0

	for _, ty := range pack.Types {
		if len(ty.Methods) == 0 && len(ty.Funcs) == 0 {
			continue
		}

		if len(ty.Methods) > 0 {
			for _, f := range ty.Methods {
				if !isExported(f) {
					continue
				}
				fun := parseFunction(f, astPack, set, ty)
				if fun == nil {
					continue
				}
				fun.Package = astPack.Name
				debugLog.Debugf("%s -> %s", ty.Name, f.Name)
				functions[fun.Name] = fun
			}
		}

		if len(ty.Funcs) > 0 {
			for _, f := range ty.Funcs {
				if !isExported(f) {
					continue
				}
				fun := parseFunction(f, astPack, set, nil)
				if fun == nil {
					continue
				}
				funcsTotal++
				fun.Package = astPack.Name
				debugLog.Debugf("package [%s] -> %s", astPack.Name, f.Name)
				functions[fun.Name] = fun
			}
		}
	}

	for _, f := range pack.Funcs {
		if !isExported(f) {
			continue
		}
		debugLog.Debugf("package [%s] -> %s", astPack.Name, f.Name)
		fun := parseFunction(f, astPack, set, nil)
		if fun == nil {
			continue
		}
		fun.Package = astPack.Name
		funcsTotal++
		functions[fun.Name] = fun
	}
	log.Infof("parsed %d functions in package %s...", funcsTotal, astPack.Name)

	return functions
}

func isExported(fun *doc.Func) bool {
	return string(fun.Name[0]) == strings.ToUpper(string(fun.Name[0]))
}

// parses the exported functions from the package
func parseFunction(fun *doc.Func, astPack *ast.Package, set *token.FileSet, ty *doc.Type) *function {
	if strings.TrimSpace(fun.Doc) == "ignore" {
		return nil
	}

	cleanedDocs := strings.TrimSpace(fun.Doc)
	cleanedDocs = strings.Replace(cleanedDocs, "\n", " ", -1)

	// extract declaration for input/output types
	decl := fun.Decl
	f := &function{
		Name: fun.Name,
		Desc: cleanedDocs,
	}

	if ty != nil {
		f.ForType = ty.Name
		f.Title = ty.Name + " -> " + fun.Name + "()"
	} else {
		f.Title = fun.Name + "()"
	}

	// search for the file with the content of this function
	// to extract the function parameter types "as-is"
	// using decl.Type.Params gives an unusable representation of the types
	found := false
	for fileName := range astPack.Files {
		fileBytes, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Fatalf("unable to read file for parameter inspection: %s", fileName)
		}

		// convert tokens positions into actual byte offsets
		start := set.Position(decl.Name.Pos()).Offset
		end := set.Position(decl.Name.End()).Offset

		// check whether the file is even as large as the end of the function declaration
		if len(fileBytes) < end {
			continue
		}

		maybeName := fileBytes[start:end]
		// check whether we found the right file
		if string(maybeName) != fun.Name {
			continue
		}

		found = true

		// extract parameters of the function
		if decl.Type.Params != nil {
			f.Inputs = make([]input, len(decl.Type.Params.List))
			paraStart := set.Position(decl.Type.Params.Pos()).Offset
			paraEnd := set.Position(decl.Type.Params.End()).Offset
			parameters := strings.Split(string(fileBytes[paraStart+1:paraEnd-1]), ",")
			for i, para := range parameters {
				split := strings.Split(strings.TrimSpace(para), " ")
				if len(split) != 2 {
					continue
				}
				var typeName string
				if !strings.HasPrefix(split[1], "...") {
					tySplit := strings.Split(split[1], ".")
					if len(tySplit) == 1 {
						typeName = tySplit[0]
					} else {
						typeName = tySplit[1]
					}
				} else {
					typeName = split[1]
				}
				f.Inputs[i] = input{ArgName: split[0], Type: typeName}
			}
		}

		// extract return values of the function
		if decl.Type.Results != nil {
			f.Outputs = make([]output, len(decl.Type.Results.List))
			resultStart := set.Position(decl.Type.Results.Pos()).Offset
			resultEnd := set.Position(decl.Type.Results.End()).Offset
			var results []string
			// if there's more than one return value we need to count in for ( )
			if len(decl.Type.Results.List) > 1 {
				results = strings.Split(string(fileBytes[resultStart+1:resultEnd-1]), ",")
			} else {
				results = strings.Split(string(fileBytes[resultStart:resultEnd]), ",")
			}
			for i, result := range results {
				var typeName string
				tySplit := strings.Split(result, ".")
				if len(tySplit) == 1 {
					typeName = tySplit[0]
				} else {
					typeName = tySplit[1]
				}
				f.Outputs[i] = output{Type: strings.TrimSpace(typeName)}
			}
		}
	}

	if !found {
		log.Fatalf("unable to find source file containing function %s of package %s", fun.Name, astPack.Name)
	}

	return f
}

// parses the corresponding example folder for the given package
func parseExamples(packageName string, packageDir string, functions map[string]*function) {
	// parse examples
	examplePackagePath := fmt.Sprintf("%s/.examples", packageDir)
	if _, err := os.Stat(examplePackagePath); os.IsNotExist(err) {
		log.Warningf("missing examples for package: %s (!)", packageName)
		return
	}

	// create a new file set for the examples
	examplesSet := token.NewFileSet()
	// parse example dir
	examplePackages, err := parser.ParseDir(examplesSet, examplePackagePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("failed to parse example packages package: %s", err)
	}

	log.Infof("parsing examples for package %s...", packageName)
	for _, exampleASTPack := range examplePackages {
		for fileName, file := range exampleASTPack.Files {
			fileBytes, err := ioutil.ReadFile(fileName)
			if err != nil {
				log.Fatalf("unable to read example file: %s", fileName)
			}
			// extract example code
			examples := doc.Examples(file)
			// extract example documentation
			examplePackage := doc.New(exampleASTPack, ".", 0)
			for _, example := range examples {
				for _, fun := range examplePackage.Funcs {
					if fun.Name != "Example"+example.Name {
						continue
					}

					f, ok := functions[example.Name]
					if !ok {
						log.Warningf("no source function found for example function %s", example.Name)
						continue
					}

					extendedDoc := example.Doc
					if len(extendedDoc) == 0 {
						debugLog.Debugf("no extended doc for function %s found", example.Name)
					} else {
						addExtendedDocs(extendedDoc, f)
					}
					f.hadExample = true

					exampleCode := fileBytes[fun.Decl.Pos():fun.Decl.End()]
					exampleBody := fileBytes[example.Code.Pos():example.Code.End()]
					if len(strings.TrimSpace(string(exampleBody))) == 1 {
						if *emptyBodyWarning {
							log.Debugf("skipping example code for function %s (empty body)", fun.Name)
						}
						continue
					}

					f.ExampleCode = "f" + string(exampleCode) + string(exampleBody)
				}
			}
		}
	}
}

// parses the expended documentation and adds it to the given function
func addExtendedDocs(extendedDocs string, f *function) {
	lines := strings.Split(extendedDocs, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		descSep := strings.Index(line, ",")
		paraSep := strings.Index(line, ":")
		// i = input
		// o = output
		switch line[0] {
		case 'i':
			required := false
			if len(line) > 5 && line[2:5] == "req" {
				required = true
			}
			paraName := strings.TrimSpace(line[paraSep+1 : descSep])
			desc := strings.TrimSpace(line[descSep+1:])
			found := false
			for j := range f.Inputs {
				input := &f.Inputs[j]
				if input.ArgName == paraName {
					found = true
					input.Desc = desc
					input.Required = required
					continue
				}
			}
			if !found {
				log.Warningf("no matching parameter '%s' in function %s", paraName, f.Name)
			}
		case 'o':
			typeName := strings.TrimSpace(line[paraSep+1 : descSep])
			desc := strings.TrimSpace(line[descSep+1:])
			found := false
			for j := range f.Outputs {
				output := &f.Outputs[j]
				if output.Type == typeName {
					found = true
					output.Desc = desc
					continue
				}
			}
			if !found {
				log.Warningf("no matching return value '%s' in function %s", typeName, f.Name)
			}
		}
	}
}
