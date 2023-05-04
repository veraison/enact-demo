package node

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/google/go-tpm/tpm2"
)

type EnactToken struct {
	// TPMS_ATTEST decoded from the token
	AttestationData *tpm2.AttestationData
	// Raw token bytes
	Raw []byte
	// TPMT_SIGNATURE decoded from the token
	Signature *tpm2.Signature
}

func (et *EnactToken) Decode(data []byte) error {
	// The first two bytes are the SIZE of the following TPMS_ATTEST
	// structure. The following SIZE bytes are the TPMS_ATTEST structure,
	// the remaining bytes are the signature.
	if len(data) < 3 {
		return fmt.Errorf("could not get data size; token too small (%d)", len(data))
	}

	size := binary.BigEndian.Uint16(data[:2])
	if len(data) < int(2+size) {
		return fmt.Errorf("TPMS_ATTEST appears truncated; expected %d, but got %d bytes",
			size, len(data)-2)
	}

	var err error

	et.Raw = data[2 : 2+size]
	et.AttestationData, err = tpm2.DecodeAttestationData(et.Raw)
	if err != nil {
		return fmt.Errorf("could not decode TPMS_ATTEST: %v", err)
	}

	et.Signature, err = tpm2.DecodeSignature(bytes.NewBuffer(data[2+size:]))
	if err != nil {
		return fmt.Errorf("could not decode TPMT_SIGNATURE: %v", err)
	}

	return nil
}

func (et EnactToken) VerifySignature(key *ecdsa.PublicKey) error {
	digest := sha256.Sum256(et.Raw)

	if !ecdsa.Verify(key, digest[:], et.Signature.ECC.R, et.Signature.ECC.S) {
		return fmt.Errorf("failed to verify signature")
	}

	return nil
}

func convertParsedBufferToToken(buf *bytes.Buffer) (EnactToken, error) {
	et := EnactToken{}
	err := et.Decode(buf.Bytes())

	return et, err
}
