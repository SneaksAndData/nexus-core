package urlsign

import (
	"encoding/binary"
	"net/url"
	"strconv"
	"time"
)

const (
	ExpiryQueryParamName   = "exp"
	TenantQueryParamName   = "tid"
	ChecksumQueryParamName = "chk"
)

type signerPayload struct {
	expiry   int64
	path     string
	tenantId string
	checksum string
}

func newSignerPayload(unsigned url.URL, tenantId string, expiresIn time.Duration, checksum string) *signerPayload {
	return &signerPayload{
		expiry:   time.Now().Add(expiresIn).UTC().Unix(),
		path:     unsigned.Path,
		tenantId: tenantId,
		checksum: checksum,
	}
}

func newSignerPayloadFromSigned(signed url.URL) (*signerPayload, error) {
	expiry, err := strconv.ParseInt(signed.Query().Get(ExpiryQueryParamName), 10, 64)

	if err != nil {
		return nil, err
	}

	return &signerPayload{
		expiry:   expiry,
		path:     signed.Path,
		tenantId: signed.Query().Get(TenantQueryParamName),
		checksum: signed.Query().Get(ChecksumQueryParamName),
	}, nil
}

func (p *signerPayload) GetPartsToSign() [][]byte {
	expiryBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(expiryBytes, uint64(p.expiry))

	return [][]byte{
		[]byte(p.path),
		[]byte(p.tenantId),
		[]byte(p.checksum),
		expiryBytes,
	}
}
