package minimal

import (
	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func RuntimeLibrary() *plugin.CodeGeneratorResponse_File {
	tmpl := `// @flow strict
import axios from "axios";

type TwirpErrorJSON = {
  code: string;
  msg: string;
  meta: { [index: string]: string };
}

export class TwirpError extends Error {
  code: string;
  meta: { [index: string]: string };

  constructor(te: TwirpErrorJSON) {
    super(te.msg);

    this.code = te.code;
    this.meta = te.meta;
  }
}

export async function sendTwirpRequest<I, O>(url: string, body: I): Promise<O> {
  try {
    const resp = await axios.post(url, body);
    return (resp.data: O);
  } catch (err) {
    if (err.response) {
      throw new TwirpError((err.response.data: TwirpErrorJSON));
    }
    throw err;
  }
}`
	cf := &plugin.CodeGeneratorResponse_File{}
	cf.Name = proto.String("twirp.js")
	cf.Content = proto.String(tmpl)

	return cf
}
