# {{.Title}}
{{.Desc}}
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.

## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
{{range .Inputs}}| {{.ArgName}} | {{.Type}} | {{.Required}} | {{.Desc}}  |
{{end}}

## Output

| Return type     | Description |
|:---------------|:--------|
{{range .Outputs}}| {{.Type}} | {{.Desc}} |
{{end}}

{{if .ExampleCode}}
## Example

```go
{{.ExampleCode}}
```
{{end}}