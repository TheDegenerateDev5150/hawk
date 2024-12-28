package docu

import (
	"bytes"
	"errors"
	"github.com/niiigoo/hawk/proto"
	"github.com/niiigoo/hawk/proto/io"
	html "html/template"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var ErrFormat = errors.New("invalid format given")

type Service interface {
	Generate(format, version string, args ...string) error
}

type service struct {
	protoService proto.Parser
}

func NewService() Service {
	return &service{
		protoService: proto.NewService(),
	}
}

func (s *service) Generate(format, version string, args ...string) error {
	file, err := s.protoService.DetectFile(args...)
	if err != nil {
		return err
	}
	source := strings.TrimSuffix(file, ".proto")

	if err = s.protoService.Parse(file, true); err != nil {
		return err
	}

	switch format {
	case "md":
		err = s.genMD(source, version)
	case "html":
		err = s.genHTML(source, version)
	default:
		err = ErrFormat
	}

	return err
}

func (s *service) genMD(source, version string) error {
	tplBytes, err := Asset("documentation.md")
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"Escape": func(s string) string {
			s = strings.ReplaceAll(s, ">", "\\>")
			s = strings.ReplaceAll(s, "<", "\\<")
			s = strings.ReplaceAll(s, "|", "\\|")
			return s
		},
		"ToUpper":           strings.ToUpper,
		"NormalizeComments": s.normalizeComments,
	}
	tpl, err := template.New("markdown").Funcs(funcMap).Parse(string(tplBytes))
	if err != nil {
		return err
	}

	messages, referenced := s.messages()
	slices.Sort(referenced)
	referenced = slices.Compact(referenced)
	outputBuffer := bytes.NewBuffer(nil)
	err = tpl.Execute(outputBuffer, Data{
		Service:     s.protoService.Definition().Services[0],
		Messages:    messages,
		Referenced:  referenced,
		VersionName: version,
		VersionTime: time.Now().Format(time.RFC822),
	})
	if err != nil {
		return err
	}

	out, err := os.OpenFile(source+".md", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err = out.Write(outputBuffer.Bytes()); err != nil {
		return err
	}

	return nil
}

func (s *service) genHTML(source, version string) error {
	if err := os.MkdirAll("html", 0777); err != nil {
		return err
	}

	tplBytes, err := Asset("html/index.html")
	if err != nil {
		return err
	}

	funcMap := html.FuncMap{
		"ToUpper":           strings.ToUpper,
		"NormalizeComments": s.normalizeComments,
	}
	tpl, err := html.New("html").Funcs(funcMap).Parse(string(tplBytes))
	if err != nil {
		return err
	}

	messages, referenced := s.messages()
	slices.Sort(referenced)
	referenced = slices.Compact(referenced)
	outputBuffer := bytes.NewBuffer(nil)
	err = tpl.Execute(outputBuffer, Data{
		Service:     s.protoService.Definition().Services[0],
		Messages:    messages,
		Referenced:  referenced,
		VersionName: version,
		VersionTime: time.Now().Format(time.RFC822),
	})
	if err != nil {
		return err
	}

	out, err := os.OpenFile("html/index.html", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()
	if _, err = out.Write(outputBuffer.Bytes()); err != nil {
		return err
	}

	fileBytes, err := Asset("html/main.css")
	if err != nil {
		return err
	}
	out, err = os.OpenFile("html/main.css", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()
	if _, err = out.Write(fileBytes); err != nil {
		return err
	}

	return nil
}

func (s *service) messages() (map[string]Message, []string) {
	messages := make(map[string]Message)
	var referenced []string

	for _, msg := range s.protoService.Definition().MessagesMap {
		fields := make([]Field, 0)
		for _, entry := range msg.Entries {
			if entry.Field != nil {
				t, ref := s.parseType(entry.Field.Type)
				if ref != nil {
					referenced = append(referenced, ref...)
				}
				f := Field{
					Name:        entry.Field.Name,
					Description: s.normalizeComments(entry.Field.Comments),
					Type:        t,
					Optional:    entry.Field.Optional,
					Options:     make([]string, 0),
				}
				if entry.Field.Repeated {
					f.Type = "Array<" + f.Type + ">"
				}
				for _, option := range entry.Field.Options {
					str := option.Name
					if option.Attr != nil {
						str = *option.Attr
					}
					str += " = " + s.parseValue(option.Value)
					f.Options = append(f.Options, str)
				}
				fields = append(fields, f)
			}
		}
		messages[msg.Name] = Message{
			Name:        msg.Name,
			Description: s.normalizeComments(msg.Comments),
			Fields:      fields,
		}
	}

	return messages, referenced
}

func (s *service) parseType(t io.Type) (string, []string) {
	if t.Scalar != io.None {
		return t.Scalar.GoString(), nil
	} else if t.Map != nil {
		var ref []string
		k, tmp := s.parseType(*t.Map.Key)
		if tmp != nil {
			ref = append(ref, tmp...)
		}
		v, tmp := s.parseType(*t.Map.Value)
		if tmp != nil {
			ref = append(ref, tmp...)
		}
		return "map<" + k + ", " + v + ">", ref
	} else {
		return t.Reference, []string{t.Reference}
	}
}

func (s *service) parseValue(v *io.Value) string {
	if v == nil {
		return ""
	} else if v.Reference != nil {
		return *v.Reference
	} else if v.Bool != nil {
		return strconv.FormatBool(bool(*v.Bool))
	} else if v.Int != nil {
		return strconv.FormatInt(*v.Int, 10)
	} else if v.Number != nil {
		return strconv.FormatFloat(*v.Number, 'f', -1, 64)
	} else if v.String != nil {
		return *v.String
	} else if v.Array != nil {
		strValues := make([]string, len(v.Array.Elements))
		for i, v := range v.Array.Elements {
			strValues[i] = s.parseValue(v)
		}
		return "[" + strings.Join(strValues, ", ") + "]"
	} else if v.Map != nil {
		str := "{"
		for i, entry := range v.Map.Entries {
			if i > 0 {
				str += ", "
			}
			str += s.parseValue(entry.Key) + ": " + s.parseValue(entry.Value)
		}
		str += "}"
		return str
	}

	return ""
}

func (s *service) normalizeComments(c []string) string {
	comment := ""
	if c != nil {
		for i, c := range c {
			c = strings.Trim(c, "/* ")
			if i > 0 {
				comment += "\n"
			}
			comment += c
		}
	}
	return comment
}
