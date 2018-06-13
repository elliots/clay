package httpruntime

import (
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/golang/protobuf/proto"
)

// Marshaler is a processor that can marshal and unmarshal data to some content-type.
type Marshaler interface {
	ContentType() string
	Unmarshal(io.Reader, proto.Message) error
	Marshal(io.Writer, proto.Message) error
}

// func to init marshaler with Content-Type/Accept params
type marshalGetterFunc = func(ContentTypeOptions) Marshaler

// ContentTypeOptions are MIME annotations provided with Content-Type or Accept
// headers.
type ContentTypeOptions map[string]string

// OverrideMarshaler replaces Marshaler for given content-type.
func OverrideMarshaler(contentType string, m Marshaler) {
	marshalDict[strings.ToLower(contentType)] = func(ContentTypeOptions) Marshaler { return m }
}

// OverrideParametrizedMarshaler replaces MarshalGetter for given content-type.
// Use it if your marshaler needs ContentTypeOptions to successfully unmarshal the request.
func OverrideParametrizedMarshaler(contentType string, f func(ContentTypeOptions) Marshaler) {
	marshalDict[strings.ToLower(contentType)] = f
}

// MarshalerForRequest returns marshalers for inbound and outbound bodies.
func MarshalerForRequest(r *http.Request) (Marshaler, Marshaler) {
	ctype, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	inbound := marshalerOrDefault(ctype, params)

	accept, aparams, _ := mime.ParseMediaType(r.Header.Get("Accept"))
	outbound := marshalerOrDefault(accept, aparams)
	return inbound, outbound
}

func marshalerOrDefault(t string, params map[string]string) Marshaler {
	t = strings.ToLower(t)

	if m, ok := marshalDict[t]; ok {
		return m(params)
	}
	return marshalDict[MarshalerPbJSON{}.ContentType()](params)
}

var marshalDict = map[string]marshalGetterFunc{
	"application/json": func(_ ContentTypeOptions) Marshaler {
		return mpbjson
	},
}