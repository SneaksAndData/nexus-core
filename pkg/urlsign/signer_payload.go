package urlsign

import (
	"encoding/binary"
	"net/url"
	"strconv"
	"time"
)

const (
	ValidFromQueryParamName = "from"
	ValidToQueryParamName   = "to"
	TenantQueryParamName    = "tid"
	ChecksumQueryParamName  = "chk"
)

type signerPayload struct {
	validFrom int64
	validTo   int64
	path      string
	tenantId  string
	checksum  string
}

func newSignerPayload(unsigned url.URL, tenantId string, expiresIn time.Duration, checksum string) *signerPayload {
	startTime := time.Now()
	return &signerPayload{
		validFrom: startTime.UTC().Unix(),
		validTo:   startTime.Add(expiresIn).UTC().Unix(),
		path:      unsigned.Path,
		tenantId:  tenantId,
		checksum:  checksum,
	}
}

func newSignerPayloadFromSigned(signed url.URL) (*signerPayload, error) {
	validFrom, err := strconv.ParseInt(signed.Query().Get(ValidFromQueryParamName), 10, 64)

	if err != nil {
		return nil, err
	}

	validTo, err := strconv.ParseInt(signed.Query().Get(ValidToQueryParamName), 10, 64)

	return &signerPayload{
		validFrom: validFrom,
		validTo:   validTo,
		path:      signed.Path,
		tenantId:  signed.Query().Get(TenantQueryParamName),
		checksum:  signed.Query().Get(ChecksumQueryParamName),
	}, nil
}

func (p *signerPayload) GetPartsToSign() [][]byte {
	fromBytes := make([]byte, 8)
	toBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(fromBytes, uint64(p.validFrom))
	binary.LittleEndian.PutUint64(toBytes, uint64(p.validTo))

	return [][]byte{
		[]byte(p.path),
		[]byte(p.tenantId),
		[]byte(p.checksum),
		fromBytes,
		toBytes,
	}
}
