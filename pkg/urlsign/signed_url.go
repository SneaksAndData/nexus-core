package urlsign

import (
	"encoding/base64"
	"fmt"
	"net/url"
)

const (
	SignatureQueryParamName = "sig"
)

// SignedUrl hold a url and a its signature in a separate field
type SignedUrl struct {
	Url       *url.URL
	Signature string
}

func NewSignedFromUrl(signed url.URL) (*SignedUrl, error) {
	parsed, err := signed.Parse(signed.String())

	if err != nil {
		return nil, err
	}

	signature := parsed.Query().Get(SignatureQueryParamName)

	if signature == "" {
		return nil, fmt.Errorf("cannot create a signed url reference for url %s - no `%s` parameter found", SignatureQueryParamName, signed.String())
	}

	return &SignedUrl{
		Url:       parsed,
		Signature: signature,
	}, nil
}

func NewSignedFromSignature(unsigned url.URL, sig []byte) (*SignedUrl, error) {
	sigBase64 := base64.RawURLEncoding.EncodeToString(sig)
	urlQuery := unsigned.Query()
	urlQuery.Add(SignatureQueryParamName, sigBase64)

	unsigned.RawQuery = urlQuery.Encode()

	return &SignedUrl{
		Url:       &unsigned,
		Signature: sigBase64,
	}, nil
}
