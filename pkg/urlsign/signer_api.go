package urlsign

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"golang.org/x/crypto/blake2b"
)

// Sign uses a provided secret to generate a cryptographic signature for the provided unsigned URL, valid for expiresIn, given checksum of the content and tenantId
func Sign(unsigned url.URL, tenantId string, expiresIn time.Duration, checksum string, secret []byte) (*SignedUrl, error) {
	if unsigned.Path == "" || unsigned.Path == "/" {
		return nil, fmt.Errorf("cannot generate a signature for an empty path")
	}
	signer, err := blake2b.New256(secret)

	// blake2b will error if len(byte) > 64 - keep this
	if err != nil {
		return nil, err
	}

	payload := newSignerPayload(unsigned, tenantId, expiresIn, checksum)
	for _, part := range payload.GetPartsToSign() {
		signer.Write(part)
	}

	return NewSignedFromSignature(unsigned, payload, signer.Sum(nil))
}

func Verify(signed url.URL, secret []byte) error {
	signer, err := blake2b.New256(secret)

	if err != nil {
		return fmt.Errorf("error when preparing for signature validation: %s", err)
	}

	providedSignature, err := base64.RawURLEncoding.DecodeString(signed.Query().Get(SignatureQueryParamName))

	if len(providedSignature) == 0 {
		return fmt.Errorf("no signature found in the url")
	}

	if err != nil {
		return fmt.Errorf("could not decode signature '%s': %s", providedSignature, err)
	}

	payload, err := newSignerPayloadFromSigned(signed)

	if err != nil {
		return fmt.Errorf("error when preparing payload for signature validation: %s", err)
	}

	for _, part := range payload.GetPartsToSign() {
		signer.Write(part)
	}

	computedSignature := signer.Sum(nil)

	if subtle.ConstantTimeCompare(providedSignature, computedSignature) != 1 {
		return fmt.Errorf("invalid signature: %s", signed.Query().Get(SignatureQueryParamName))
	}

	// check expiry if signature matches
	if time.Now().UTC().After(time.Unix(payload.validTo, 0)) {
		return fmt.Errorf("url is no longer valid")
	}

	return nil
}
