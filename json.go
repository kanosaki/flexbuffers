package flexbuffers

// json parsing code based on https://github.com/valyala/fastjson

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/valyala/fastjson/fastfloat"
)

func FromJson(data []byte) (Raw, error) {
	b := NewBuilder()
	r := JsonReader{Output: b}
	_, err := r.parseValue(b2s(data))
	if err != nil {
		return nil, err
	}
	if err := b.Finish(); err != nil {
		return nil, err
	}
	return b.Buffer(), nil
}

type StreamReceiver interface {
	PushString(s string) error
	PushBlob(b []byte) error
	PushInt(i int64) error
	PushUint(u uint64) error
	PushFloat(f float64) error
	PushBool(b bool) error
	PushNull() error

	BeginArray() (int, error)
	EndArray(int) error

	BeginObject() (int, error)
	EndObject(int) error
	PushObjectKey(k string) error
}

type JsonReader struct {
	Output StreamReceiver
}

func skipWS(s string) string {
	if len(s) == 0 || s[0] > 0x20 {
		// Fast path.
		return s
	}
	return skipWSSlow(s)
}

func skipWSSlow(s string) string {
	if len(s) == 0 || s[0] != 0x20 && s[0] != 0x0A && s[0] != 0x09 && s[0] != 0x0D {
		return s
	}
	for i := 1; i < len(s); i++ {
		if s[i] != 0x20 && s[i] != 0x0A && s[i] != 0x09 && s[i] != 0x0D {
			return s[i:]
		}
	}
	return ""
}

func (r *JsonReader) parseValue(s string) (string, error) {
	if len(s) == 0 {
		return s, fmt.Errorf("cannot parse empty string")
	}

	if s[0] == '{' {
		ptr, err := r.Output.BeginObject()
		if err != nil {
			return s, err
		}
		tail, err := r.parseObject(s[1:])
		if err != nil {
			return tail, fmt.Errorf("cannot parse object: %s", err)
		}
		return tail, r.Output.EndObject(ptr)
	}
	if s[0] == '[' {
		ptr, err := r.Output.BeginArray()
		if err != nil {
			return s, err
		}
		tail, err := r.parseArray(s[1:])
		if err != nil {
			return tail, fmt.Errorf("cannot parse array: %s", err)
		}
		return tail, r.Output.EndArray(ptr)
	}
	if s[0] == '"' {
		ss, tail, err := parseRawString(s[1:])
		if err != nil {
			return tail, fmt.Errorf("cannot parse string: %s", err)
		}
		return tail, r.Output.PushString(unescapeStringBestEffort(ss))
	}
	if s[0] == 't' {
		if len(s) < len("true") || s[:len("true")] != "true" {
			return s, fmt.Errorf("unexpected value found: %q", s)
		}
		return s[len("true"):], r.Output.PushBool(true)
	}
	if s[0] == 'f' {
		if len(s) < len("false") || s[:len("false")] != "false" {
			return s, fmt.Errorf("unexpected value found: %q", s)
		}
		return s[len("false"):], r.Output.PushBool(false)
	}
	if s[0] == 'n' {
		if len(s) < len("null") || s[:len("null")] != "null" {
			return s, fmt.Errorf("unexpected value found: %q", s)
		}
		return s[len("null"):], r.Output.PushNull()
	}

	integral, ns, tail, err := parseRawNumber(s)
	if err != nil {
		return tail, fmt.Errorf("cannot parse number: %s", err)
	}
	if integral {
		return tail, r.Output.PushInt(fastfloat.ParseInt64BestEffort(ns))
	} else {
		return tail, r.Output.PushFloat(fastfloat.ParseBestEffort(ns))
	}
}

func (r *JsonReader) parseArray(s string) (string, error) {
	s = skipWS(s)
	if len(s) == 0 {
		return s, fmt.Errorf("missing ']'")
	}

	if s[0] == ']' {
		return s[1:], nil
	}

	for {
		var err error

		s = skipWS(s)
		s, err = r.parseValue(s)
		if err != nil {
			return s, fmt.Errorf("cannot parse array value: %s", err)
		}

		s = skipWS(s)
		if len(s) == 0 {
			return s, fmt.Errorf("unexpected end of array")
		}
		if s[0] == ',' {
			s = s[1:]
			continue
		}
		if s[0] == ']' {
			s = s[1:]
			return s, nil
		}
		return s, fmt.Errorf("missing ',' after array value")
	}
}

func (r *JsonReader) parseObject(s string) (string, error) {
	s = skipWS(s)
	if len(s) == 0 {
		return s, fmt.Errorf("missing '}'")
	}

	if s[0] == '}' {
		return s[1:], nil
	}

	for {
		var err error

		// Parse key.
		s = skipWS(s)
		if len(s) == 0 || s[0] != '"' {
			return s, fmt.Errorf(`cannot find opening '"" for object key`)
		}
		var k string
		k, s, err = parseRawKey(s[1:])
		if err != nil {
			return s, fmt.Errorf("cannot parse object key: %s", err)
		}
		if err := r.Output.PushObjectKey(k); err != nil {
			return s, err
		}
		s = skipWS(s)
		if len(s) == 0 || s[0] != ':' {
			return s, fmt.Errorf("missing ':' after object key")
		}
		s = s[1:]

		// Parse value
		s = skipWS(s)
		s, err = r.parseValue(s)
		if err != nil {
			return s, fmt.Errorf("cannot parse object value: %s", err)
		}
		s = skipWS(s)
		if len(s) == 0 {
			return s, fmt.Errorf("unexpected end of object")
		}
		if s[0] == ',' {
			s = s[1:]
			continue
		}
		if s[0] == '}' {
			return s[1:], nil
		}
		return s, fmt.Errorf("missing ',' after object value")
	}
}

func escapeString(dst []byte, s string) []byte {
	if !hasSpecialChars(s) {
		// Fast path - nothing to escape.
		dst = append(dst, '"')
		dst = append(dst, s...)
		dst = append(dst, '"')
		return dst
	}

	// Slow path.
	return strconv.AppendQuote(dst, s)
}

func hasSpecialChars(s string) bool {
	if strings.IndexByte(s, '"') >= 0 || strings.IndexByte(s, '\\') >= 0 {
		return true
	}
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 {
			return true
		}
	}
	return false
}

func unescapeStringBestEffort(s string) string {
	n := strings.IndexByte(s, '\\')
	if n < 0 {
		// Fast path - nothing to unescape.
		return s
	}

	// Slow path - unescape string.
	var b []byte
	b = b[:n]
	s = s[n+1:]
	for len(s) > 0 {
		ch := s[0]
		s = s[1:]
		switch ch {
		case '"':
			b = append(b, '"')
		case '\\':
			b = append(b, '\\')
		case '/':
			b = append(b, '/')
		case 'b':
			b = append(b, '\b')
		case 'f':
			b = append(b, '\f')
		case 'n':
			b = append(b, '\n')
		case 'r':
			b = append(b, '\r')
		case 't':
			b = append(b, '\t')
		case 'u':
			if len(s) < 4 {
				// Too short escape sequence. Just store it unchanged.
				b = append(b, "\\u"...)
				break
			}
			xs := s[:4]
			x, err := strconv.ParseUint(xs, 16, 16)
			if err != nil {
				// Invalid escape sequence. Just store it unchanged.
				b = append(b, "\\u"...)
				break
			}
			s = s[4:]
			if !utf16.IsSurrogate(rune(x)) {
				b = append(b, string(rune(x))...)
				break
			}

			// Surrogate.
			// See https://en.wikipedia.org/wiki/Universal_Character_Set_characters#Surrogates
			if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
				b = append(b, "\\u"...)
				b = append(b, xs...)
				break
			}
			x1, err := strconv.ParseUint(s[2:6], 16, 16)
			if err != nil {
				b = append(b, "\\u"...)
				b = append(b, xs...)
				break
			}
			r := utf16.DecodeRune(rune(x), rune(x1))
			b = append(b, string(r)...)
			s = s[6:]
		default:
			// Unknown escape sequence. Just store it unchanged.
			b = append(b, '\\', ch)
		}
		n = strings.IndexByte(s, '\\')
		if n < 0 {
			b = append(b, s...)
			break
		}
		b = append(b, s[:n]...)
		s = s[n+1:]
	}
	return b2s(b)
}

// parseRawKey is similar to parseRawString, but is optimized
// for small-sized keys without escape sequences.
func parseRawKey(s string) (string, string, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			// Fast path.
			return s[:i], s[i+1:], nil
		}
		if s[i] == '\\' {
			// Slow path.
			return parseRawString(s)
		}
	}
	return s, "", fmt.Errorf(`missing closing '"'`)
}

func parseRawString(s string) (string, string, error) {
	n := strings.IndexByte(s, '"')
	if n < 0 {
		return s, "", fmt.Errorf(`missing closing '"'`)
	}
	if n == 0 || s[n-1] != '\\' {
		// Fast path. No escaped ".
		return s[:n], s[n+1:], nil
	}

	// Slow path - possible escaped " found.
	ss := s
	for {
		i := n - 1
		for i > 0 && s[i-1] == '\\' {
			i--
		}
		if uint(n-i)%2 == 0 {
			return ss[:len(ss)-len(s)+n], s[n+1:], nil
		}
		s = s[n+1:]

		n = strings.IndexByte(s, '"')
		if n < 0 {
			return ss, "", fmt.Errorf(`missing closing '"'`)
		}
		if n == 0 || s[n-1] != '\\' {
			return ss[:len(ss)-len(s)+n], s[n+1:], nil
		}
	}
}

func parseRawNumber(s string) (bool, string, string, error) {
	// The caller must ensure len(s) > 0

	integral := true
	// Find the end of the number.
	for i := 0; i < len(s); i++ {
		ch := s[i]
		chIntegral := (ch >= '0' && ch <= '9') || ch == '-' || ch == '+'
		floatElem := ch == '.' || ch == 'e' || ch == 'E'
		integral = integral && !floatElem
		if chIntegral || floatElem {
			continue
		}
		if i == 0 {
			return integral, "", s, fmt.Errorf("unexpected char: %q", s[:1])
		}
		ns := s[:i]
		s = s[i:]
		return integral, ns, s, nil
	}
	return integral, s, "", nil
}
