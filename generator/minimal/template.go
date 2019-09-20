package minimal

import (
	"fmt"
	"sort"
	"strings"
)

const apiTemplate = `// @flow strict
import { sendTwirpRequest } from "./twirp";
{{imports .}}
{{range .Models}}
{{- if not .Primitive}}
export type {{.Name}} = {
  {{range .Fields -}}
  {{.Name}}: {{if .IsMessage}}{{if .IsRepeated}}{{else}}?{{end}}{{end}}{{.Type}};
  {{end}}
}

export type {{.Name}}JSON = {
  {{range .Fields -}}
  {{.JSONName}}?: {{.JSONType}};
  {{end}}
}

export function {{.Name}}ToJSON(m: {{.Name}}): {{.Name}}JSON {
  return {
    {{range .Fields -}}
    {{.JSONName}}: {{stringify .}},
    {{end}}
  };
}

export function JSONTo{{.Name}}(m: {{.Name}}JSON): {{.Name}} {
  return {
    {{range .Fields -}}
    {{.Name}}: {{parse .}},
    {{end}}
  };
}
{{end -}}
{{end}}

{{- $twirpPrefix := .TwirpPrefix -}}
{{$Ctx := .}}
{{range .Services}}
export class {{.Name}} {
  hostname: string;
  pathPrefix = "{{$twirpPrefix}}/{{.Package}}.{{.Name}}/";

  constructor(hostname: string) {
    this.hostname = hostname;
  }

  {{- range .Methods}}
  {{methodBody $Ctx .}}
  {{end}}
}
{{end}}
`

func stringify(f ModelField) string {
	if f.IsRepeated {
		singularType := strings.TrimSuffix(strings.TrimPrefix(f.Type, "Array<"), ">") // strip array brackets from type

		if f.Type == "Date" {
			return fmt.Sprintf("m.%s.map((n) => n.toISOString())", f.Name)
		}

		if f.IsMessage {
			return fmt.Sprintf("m.%s.map(%sToJSON)", f.Name, singularType)
		}
	}

	if f.Type == "Date" {
		return fmt.Sprintf("m.%s.toISOString()", f.Name)
	}

	if f.IsMessage {
		return fmt.Sprintf("m.%s != null ? %sToJSON(m.%s) : undefined", f.Name, f.Type, f.Name)
	}

	return "m." + f.Name
}

func parse(f ModelField) string {
	field := "m." + f.JSONName

	if f.IsRepeated {
		singularTSType := strings.TrimSuffix(strings.TrimPrefix(f.Type, "Array<"), ">") // strip array brackets from type

		if f.Type == "Array<Date>" {
			return fmt.Sprintf("%s != null ? %s.map((n) => new Date(n)) : []", field, field)
		}

		if f.IsMessage {
			return fmt.Sprintf("%s != null ? %s.map(JSONTo%s) : []", field, field, singularTSType)
		}

		return fmt.Sprintf("%s != null ? %s : []", field, field)
	}

	if f.Type == "Date" {
		return fmt.Sprintf("new Date(%s)", field)
	}

	if f.IsMessage {
		return fmt.Sprintf("%s != null ? JSONTo%s(%s) : null", field, f.Type, field)
	}

	switch f.Type {
	case "string":
		return fmt.Sprintf(`%s != null ? %s : ""`, field, field)
	case "number":
		return fmt.Sprintf(`%s != null ? %s : 0`, field, field)
	case "boolean":
		return fmt.Sprintf(`%s != null ? %s : false`, field, field)
	}

	return field
}

func imports(ctx *APIContext) string {
	res := make(map[string]map[string]bool)

	for _, m := range ctx.Models {
		for _, f := range m.Fields {
			baseType := f.Type
			if f.IsRepeated {
				baseType = strings.TrimSuffix(strings.TrimPrefix(baseType, "Array<"), ">")
			}
			model := ctx.contexts.modelLookup[baseType]
			if model != nil {
				if ctx.filename != model.FileName {
					if res[model.FileName] == nil {
						res[model.FileName] = make(map[string]bool)
					}
					res[model.FileName][model.Name] = true
				}
			}
		}
	}

	fileNames := make([]string, 0, len(res))
	for fileName := range res {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	var items []string
	for _, fileName := range fileNames {
		modelNames := make([]string, 0, len(res[fileName]))
		for modelName := range res[fileName] {
			modelNames = append(modelNames, modelName)
		}
		sort.Strings(modelNames)

		fileName = strings.TrimSuffix(fileName, ".js")
		for _, modelName := range modelNames {
			items = append(items, fmt.Sprintf(`import type { %s, %sJSON } from "./%s";`, modelName, modelName, fileName))
			items = append(items, fmt.Sprintf(`import { %sToJSON, JSONTo%s } from "./%s";`, modelName, modelName, fileName))
		}
	}

	return strings.Join(items, "\n")
}

func methodBody(ctx *APIContext, m ServiceMethod) string {
	inputModel := ctx.contexts.modelLookup[m.InputType]
	outputModel := ctx.contexts.modelLookup[m.OutputType]
	emptyInputModel := inputModel != nil && len(inputModel.Fields) == 0
	emptyOutputModel := outputModel != nil && len(outputModel.Fields) == 0

	if !emptyInputModel && !emptyOutputModel {
		return fmt.Sprintf(`async %s(%s: %s): Promise<%s> {
    const url = this.hostname + this.pathPrefix + "%s";
    const body: %sJSON = %sToJSON(%s);
    const data = await sendTwirpRequest(url, body);
    return JSONTo%s(data);
  }`, m.Name, m.InputArg, m.InputType, m.OutputType, m.Path, m.InputType, m.InputType, m.InputArg, m.OutputType)
	}

	if emptyInputModel && !emptyOutputModel {
		return fmt.Sprintf(`async %s(): Promise<%s> {
    const url = this.hostname + this.pathPrefix + "%s";
    const data = await sendTwirpRequest(url, {});
    return JSONTo%s(data);
  }`, m.Name, m.OutputType, m.Path, m.OutputType)
	}

	if !emptyInputModel && emptyOutputModel {
		return fmt.Sprintf(`async %s(%s: %s): Promise<void> {
    const url = this.hostname + this.pathPrefix + "%s";
    const body: %sJSON = %sToJSON(%s);
    await sendTwirpRequest(url, body);
  }`, m.Name, m.InputArg, m.InputType, m.Path, m.InputType, m.InputType, m.InputArg)
	}

	return fmt.Sprintf(`async %s(): Promise<void> {
    const url = this.hostname + this.pathPrefix + "%s";
    await sendTwirpRequest(url, {});
  }`, m.Name, m.Path)
}
