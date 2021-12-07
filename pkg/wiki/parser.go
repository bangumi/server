// nolint:gomnd
package wiki

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/valyala/bytebufferpool"
)

var ErrWikiSyntax = errors.New("invalid wiki syntax")

type Item struct {
	Key   string
	Value string
}

func (f Field) marshalJSON() []byte {
	b := bytes.NewBuffer(nil)

	b.WriteString(strconv.Quote(f.Key))
	b.WriteString(":")
	if f.Array {
		s := make([]string, len(f.Values))
		for j, value := range f.Values {
			s[j] = "[" + strconv.Quote(value.Key) + "," + strconv.Quote(value.Value) + "]"
		}

		b.WriteString("[")
		b.WriteString(strings.Join(s, ","))
		b.WriteString("]")
	} else {
		b.WriteString(strconv.Quote(f.Value))
	}

	return b.Bytes()
}

type Field struct {
	Key    string
	Value  string
	Values []Item
	Null   bool
	Array  bool
}

type Wiki struct {
	Type   string
	Fields []Field
}

func (w Wiki) MarshalJSON() ([]byte, error) {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	b.WriteByte(byte('{'))
	bb := make([][]byte, 0, len(w.Fields))

	for _, f := range w.Fields {
		if f.Null {
			continue
		}

		bb = append(bb, f.marshalJSON())
	}

	b.Write(bb[0])
	for _, p := range bb[1:] {
		b.WriteByte(byte(','))
		b.Write(p)
	}
	b.WriteByte(byte('}'))

	return b.Bytes(), nil
}

func Parse(s string) (w Wiki, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}

	s = strings.ReplaceAll(s, "\r\n", "\n")

	lines := strings.Split(s, "\n")
	firstLine := lines[0]

	if !strings.HasPrefix(firstLine, "{{Infobox") {
		return w, ErrWikiSyntax
	}

	if lines[len(lines)-1] != "}}" {
		return w, errors.Wrap(ErrWikiSyntax, "missing '}}' at the end of text")
	}

	subjectType := strings.TrimSpace(firstLine[len("{{Infobox"):])

	components := strings.Split(strings.Join(lines[1:len(lines)-1], "\n")[1:], "\n|")

	var fields = make([]Field, 0, len(components))
	for _, component := range components {
		component = trim(component)
		if component == "" {
			continue
		}

		field, err := parseComponent(component)
		if err != nil {
			return w, errors.Wrap(ErrWikiSyntax, "failed to parse "+strconv.Quote(component))
		}

		fields = append(fields, *field)
	}

	return Wiki{
		Type:   subjectType,
		Fields: fields,
	}, nil
}

func trim(s string) string {
	return strings.Trim(s, trimCutSet)
}

const trimCutSet = " \t\r\n"

// s should be trimmed.
func parseComponent(s string) (*Field, error) {
	lines := strings.Split(s, "\n")

	// single line case, can't be a list
	if len(lines) == 1 {
		f := strings.SplitN(lines[0], "=", 2)
		f[0] = strings.TrimRight(f[0], trimCutSet)

		switch len(f) {
		case 1:
			return &Field{Key: f[0], Null: true}, nil
		case 2:
			v := strings.TrimSpace(f[1])
			if v == "" {
				return &Field{Key: f[0], Null: true}, nil
			}

			return &Field{Key: f[0], Value: v}, nil
		default:
			return nil, errors.New("code not reachable")
		}
	}

	// multi line case
	f := strings.SplitN(s, "=", 2)

	key := f[0]
	field := Field{Array: true, Key: strings.TrimSpace(key)}
	rawValue := strings.TrimLeft(f[1], trimCutSet)

	if !(strings.HasPrefix(rawValue, "{") && strings.HasSuffix(rawValue, "}")) {
		return nil, errors.Wrap(ErrWikiSyntax, "multi line content should wapped by '{}'")
	}

	v, err := parseMultiValue(rawValue)

	field.Values = v

	return &field, err
}

func parseMultiValue(s string) ([]Item, error) {
	lines := strings.Split(s[2:len(s)-2], "\n")
	a := make([]Item, 0, len(lines))
	for _, line := range lines {
		v := strings.TrimSpace(line)
		if v == "" {
			continue
		}

		if !(strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]")) {
			return nil, ErrWikiSyntax
		}
		inner := strings.TrimSpace(v[1 : len(v)-1])
		innerS := strings.SplitN(inner, "|", 2)

		i := Item{}

		switch len(innerS) {
		case 1:
			i.Value = innerS[0]
		case 2:
			i.Key = strings.TrimSpace(innerS[0])
			i.Value = innerS[1]
		default:
			panic(fmt.Sprintf("error item value '%s'", line))
		}

		a = append(a, i)
	}

	return a, nil
}
