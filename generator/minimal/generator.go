package minimal

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func NewGenerator(twirpVersion string, p map[string]string) *Generator {
	funcMap := template.FuncMap{
		"stringify":  stringify,
		"parse":      parse,
		"imports":    imports,
		"methodBody": methodBody,
	}

	apiTemplate, err := template.New("client_api").Funcs(funcMap).Parse(apiTemplate)
	if err != nil {
		panic(err)
	}

	twirpPrefix := "/twirp"
	if twirpVersion == "v6" {
		twirpPrefix = ""
	}

	contexts := NewAPIContexts(twirpPrefix)
	contexts.modelLookup["Date"] = &Model{
		Name:      "Date",
		Primitive: true,
	}

	return &Generator{
		params:      p,
		apiTemplate: apiTemplate,
		contexts:    contexts,
	}
}

type Generator struct {
	params      map[string]string
	apiTemplate *template.Template
	contexts    *APIContexts
}

func (g *Generator) Prepare(d *descriptor.FileDescriptorProto) {
	// skip WKT Timestamp, we don't do any special serialization for jsonpb.
	if *d.Name == "google/protobuf/timestamp.proto" {
		return
	}

	g.contexts.addContext(jsModuleFilename(d))

	pkg := d.GetPackage()

	for _, m := range d.GetMessageType() {
		model := &Model{
			FileName: jsModuleFilename(d),
			Name:     m.GetName(),
		}

		for _, f := range m.GetField() {
			model.Fields = append(model.Fields, newField(f))
		}

		g.contexts.AddModel(model)
	}

	for _, s := range d.GetService() {
		service := &Service{
			Name:    s.GetName(),
			Package: pkg,
		}

		for _, m := range s.GetMethod() {
			methodPath := m.GetName()
			methodName := strings.ToLower(methodPath[0:1]) + methodPath[1:]
			in := removePkg(m.GetInputType())
			arg := strings.ToLower(in[0:1]) + in[1:]

			method := ServiceMethod{
				Name:       methodName,
				Path:       methodPath,
				InputArg:   arg,
				InputType:  in,
				OutputType: removePkg(m.GetOutputType()),
			}

			service.Methods = append(service.Methods, method)
		}

		g.contexts.AddService(service)
	}
}

func (g *Generator) Generate() (files []*plugin.CodeGeneratorResponse_File, err error) {
	for _, ctx := range g.contexts.contexts {
		b := bytes.NewBufferString("")
		err := g.apiTemplate.Execute(b, ctx)
		if err != nil {
			return nil, err
		}

		clientAPI := &plugin.CodeGeneratorResponse_File{}
		clientAPI.Name = proto.String(ctx.filename)
		clientAPI.Content = proto.String(b.String())

		files = append(files, clientAPI)
	}

	files = append(files, RuntimeLibrary())

	return files, nil
}
