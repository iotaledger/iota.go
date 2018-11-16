# {{.Title}}
{{.Desc}}
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.

{{if .Inputs}}
## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
{{range .Inputs}}| {{.ArgName}} | {{.Type}} | {{if .Required}}Required{{else}}Optional{{end}} | {{.Desc}}  |
{{end}}
{{end}}

{{if .Outputs}}
## Output

| Return type     | Description |
|:---------------|:--------|
{{range .Outputs}}| {{.Type}} | {{.Desc}} |
{{end}}
{{end}}

{{if .ExampleCode}}
## Example

```go
{{.ExampleCode}}
```
{{end}}