package node

import (
	// "backend/pkg/enactcorim"
	"bytes"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/uuid"
	"github.com/veraison/enact-demo/pkg/enactcorim"
	"github.com/veraison/enact-demo/pkg/veraison"
)

type NodeService struct {
	repo NodeRepository
}
type Node struct {
	ID         uuid.UUID `db:"id"`
	AK_Pub     string    `db:"ak_pub"`
	EK_Pub     string    `db:"ek_pub"`
	Created_At string    `db:"created_at"`
}

func NewService(repo NodeRepository) *NodeService {
	return &NodeService{
		repo: repo,
	}
}

func (n *NodeService) HandleReceivePEM(akPub string, ekPub string) (uuid.UUID, error) {
	// 1. From the agent: `POST /node/pem, Body: { AK_pub, EK_pub }`
	// 2. Generate node_id (UUID v4)
	nodeID, err := uuid.NewUUID()
	if err != nil {
		log.Println("Error generating node UUID")
		return nodeID, err
	}

	// 3. Init node entity and store it in the db
	node := Node{
		ID:         nodeID,
		AK_Pub:     akPub,
		EK_Pub:     ekPub,
		Created_At: time.Now().UTC().String(),
	}

	err = n.repo.InsertNode(node)
	if err != nil {
		log.Println(err.Error())
		return nodeID, err
	}

	// 4. Repackage node_id and AK pub as CoRIM
	corim, err := enactcorim.RepackageNodePEM(akPub, nodeID)
	if err != nil {
		log.Println(err)
		return nodeID, err
	}

	log.Println(`successfully repacked node PEM as corim`)

	cbor, err := corim.ToCBOR()
	if err != nil {
		log.Fatal(err)
		return nodeID, err
	}

	log.Println(`successfully converted corim to cbor`)

	// 5. `POST /submit, Body: { CoRIM }` to veraison backend and forward response to agent
	err = veraison.SendPEMToVeraison(cbor)
	if err != nil {
		log.Println(err)
		return nodeID, err
	}

	return nodeID, nil
}

// TODO: move to util module if func is still needed
func concatBuffers(firstBlob *bytes.Buffer, secondBlob *bytes.Buffer) *bytes.Buffer {
	buf := &bytes.Buffer{}

	// read into a buffer as little endian ->
	binary.Write(buf, binary.BigEndian, firstBlob.Bytes())
	binary.Write(buf, binary.BigEndian, secondBlob.Bytes())

	return buf
}

type TokenDescription struct {
	QualifiedSigner []byte `json:"signer"`
	Digest          []byte `json:"digest"`
	PCRs            []int  `json:"pcrs"`
	FirmwareVersion uint64 `json:"firmware"`
	Algorithm       uint16 `json:"algorithm"`
	Type            uint16 `json:"type"`
}

func makeAttestationData(desc *TokenDescription) tpm2.AttestationData {
	return tpm2.AttestationData{
		Magic: 0xff544347,
		QualifiedSigner: tpm2.Name{
			Digest: &tpm2.HashValue{Alg: tpm2.AlgSHA256, Value: desc.QualifiedSigner},
		},
		FirmwareVersion: desc.FirmwareVersion,
		Type:            tpm2.TagAttestQuote,
		AttestedQuoteInfo: &tpm2.QuoteInfo{
			PCRSelection: tpm2.PCRSelection{Hash: tpm2.AlgSHA256, PCRs: desc.PCRs},
			PCRDigest:    desc.Digest,
		},
	}
}

// Hacky: there's no generic type for structs, since golang uses interfaces for abstraction.
// It's also not a good practice to convert structs to byte arrays - a format suitable for serialization should be used.
func structToByteArr(genericStruct interface{}) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(genericStruct)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	return buf.Bytes()
}

// swapUint16 converts a uint16 to network byte order and back.
func swapUint16(n uint16) uint16 {
	return (n&0x00FF)<<8 | (n&0xFF00)>>8
}

// Token is the container for the decoded EnactTrust token
type Token struct {
	// TPMS_ATTEST decoded from the token
	AttestationData *tpm2.AttestationData
	// Raw token bytes
	Raw []byte
	// TPMT_SIGNATURE decoded from the token
	Signature *tpm2.Signature
}

func (t *Token) Decode(data []byte) error {
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

	t.Raw = data[2 : 2+size]
	t.AttestationData, err = tpm2.DecodeAttestationData(t.Raw)
	if err != nil {
		return fmt.Errorf("could not decode TPMS_ATTEST: %v", err)
	}

	t.Signature, err = tpm2.DecodeSignature(bytes.NewBuffer(data[2+size:]))
	if err != nil {
		return fmt.Errorf("could not decode TPMT_SIGNATURE: %v", err)
	}

	return nil
}

func (t Token) VerifySignature(key *ecdsa.PublicKey) error {
	digest := sha256.Sum256(t.Raw)

	if !ecdsa.Verify(key, digest[:], t.Signature.ECC.R, t.Signature.ECC.S) {
		return fmt.Errorf("failed to verify signature")
	}

	return nil
}

func parseKey(keyString string) (*ecdsa.PublicKey, error) {
	buf, err := base64.StdEncoding.DecodeString(keyString)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParsePKIXPublicKey(buf)
	if err != nil {
		return nil, fmt.Errorf("could not parse public key: %v", err)
	}

	ret, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("could not extract EC public key; got [%T]: %v", key, err)
	}

	return ret, nil
}

func (n *NodeService) HandleGoldenValue(nodeID string, goldenBlob *bytes.Buffer, signatureBlob *bytes.Buffer) error {
	log.Println("goldenBlob + signature bytes:", len(goldenBlob.Bytes())+len(signatureBlob.Bytes()))

	bigEndianBuf := &bytes.Buffer{}

	log.Println(len(bigEndianBuf.Bytes()))

	/* Parse TPMS_ATTEST */

	// 0. TPMS_ATTEST Size
	val := goldenBlob.Next(2)
	tpmsAttestSize := binary.LittleEndian.Uint16(val)
	// Convert to Big Endian
	sizeBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeBuf, tpmsAttestSize)
	// Write the Big Endian byte array to the buffer
	binary.Write(bigEndianBuf, binary.BigEndian, sizeBuf)
	log.Println(len(bigEndianBuf.Bytes()))

	// 1. TPM_Magic
	val = goldenBlob.Next(4)
	tpmMagic := binary.BigEndian.Uint32(val)
	// Data is already Big Endian, writing directly to buffer
	binary.Write(bigEndianBuf, binary.BigEndian, tpmMagic)
	log.Println(len(bigEndianBuf.Bytes()))

	// 2. attest type
	val = goldenBlob.Next(2)
	attestType := binary.BigEndian.Uint16(val)
	// Data is already Big Endian
	binary.Write(bigEndianBuf, binary.BigEndian, attestType)
	log.Println(len(bigEndianBuf.Bytes()))

	// 3. qualifiedSigner
	val = goldenBlob.Next(2)
	nestedBufferSize := binary.BigEndian.Uint16(val)
	qualifiedSigner := goldenBlob.Next(int(nestedBufferSize))
	// Size field is already Big Endian
	binary.Write(bigEndianBuf, binary.BigEndian, nestedBufferSize)
	log.Println(len(bigEndianBuf.Bytes()))
	// Qualified buffer is already Big Endian
	binary.Write(bigEndianBuf, binary.BigEndian, qualifiedSigner)
	log.Println(len(bigEndianBuf.Bytes()))

	// 4. extra data - node_id
	val = goldenBlob.Next(2)
	nestedBufferSize = binary.BigEndian.Uint16(val)
	extraData := goldenBlob.Next(int(nestedBufferSize))
	// Size field is already Big Endian
	binary.Write(bigEndianBuf, binary.BigEndian, nestedBufferSize)
	log.Println(len(bigEndianBuf.Bytes()))
	// Extra data buffer is already Big Endian
	binary.Write(bigEndianBuf, binary.BigEndian, extraData)
	log.Println(len(bigEndianBuf.Bytes()))

	// 5. Digest
	val = goldenBlob.Next(2)
	// TODO: check digestSize, should usually be 24 bytes for SHA-256
	digestSize := binary.BigEndian.Uint16(val)
	evidenceDigest := goldenBlob.Next(int(digestSize))
	// Size field is already Big Endian
	binary.Write(bigEndianBuf, binary.BigEndian, digestSize)
	log.Println(len(bigEndianBuf.Bytes()))
	// Digest is already Big Endian
	binary.Write(bigEndianBuf, binary.BigEndian, evidenceDigest)
	log.Println(len(bigEndianBuf.Bytes()))

	/* Parse signature struct */
	// https://dox.ipxe.org/structTPMT__SIGNATURE.html

	// 0. TPMI_ALG_SIG_SCHEME
	val = signatureBlob.Next(2)
	sigAlgId := binary.LittleEndian.Uint16(val)
	// Convert to Big Endian
	tempBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(tempBuf, sigAlgId)
	// Write the Big Endian byte array to buffer
	binary.Write(bigEndianBuf, binary.BigEndian, tempBuf)
	log.Println(len(bigEndianBuf.Bytes()))

	// 1. TPMU_SIGNATURE
	// 		struct TPMS_SIGNATURE_ECC {
	// 			TPMI_ALG_HASH hash; - UINT16
	// 			TPM2B_ECC_PARAMETER signatureR; - UINT16 size + BYTE buffer[MAX_ECC_KEY_BYTES];
	// 			TPM2B_ECC_PARAMETER signatureS; - UINT16 size + BYTE buffer[MAX_ECC_KEY_BYTES];
	// 		}

	val = signatureBlob.Next(2)
	tpmiAlgHash := binary.LittleEndian.Uint16(val)
	// Swap endiannnes of tpmiAlgHash
	tempBuf = make([]byte, 2)
	binary.BigEndian.PutUint16(tempBuf, tpmiAlgHash)
	binary.Write(bigEndianBuf, binary.BigEndian, tempBuf)
	log.Println(len(bigEndianBuf.Bytes()))

	// Parsing signature R
	val = signatureBlob.Next(2)
	sizeR := binary.LittleEndian.Uint16(val)
	sigR := signatureBlob.Next(int(sizeR))

	// Swap endiannnes of sizeR
	tempBuf = make([]byte, 2)
	binary.BigEndian.PutUint16(tempBuf, sizeR)
	// Size field
	binary.Write(bigEndianBuf, binary.BigEndian, tempBuf)
	log.Println(len(bigEndianBuf.Bytes()))
	// SigR part
	binary.Write(bigEndianBuf, binary.BigEndian, sigR)
	log.Println(len(bigEndianBuf.Bytes()))

	// Parsing signature S
	val = signatureBlob.Next(2)
	sizeS := binary.LittleEndian.Uint16(val)
	sigS := signatureBlob.Next(int(sizeS))

	// Swap endiannnes of sizeS
	tempBuf = make([]byte, 2)
	binary.BigEndian.PutUint16(tempBuf, sizeS)
	// Size field
	binary.Write(bigEndianBuf, binary.BigEndian, tempBuf)
	log.Println(len(bigEndianBuf.Bytes()))
	// SigS part
	binary.Write(bigEndianBuf, binary.BigEndian, sigS)
	log.Println(len(bigEndianBuf.Bytes()))

	t := Token{}
	err := t.Decode(bigEndianBuf.Bytes())
	if err != nil {
		log.Println(err)
	}

	// TODO: read secret from here
	// Extract nodeID from ExtraData
	dimiNodeID, err := uuid.FromBytes(t.AttestationData.ExtraData)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Dimi's nodeID:", nodeID)
	log.Println("Dimi's nodeID:", dimiNodeID.String())

	node, err := n.repo.GetNodeById(dimiNodeID.String())
	if err != nil {
		return err
	}

	key, err := parseKey(node.AK_Pub)
	if err != nil {
		log.Println(err)
	}

	err = t.VerifySignature(key)
	if err != nil {
		return err
	} else {
		log.Println("Signature check is GOOD.")
		//Store golden PCR value in Enact DB
		//Mark node as onboarded
	}

	// Write to a file
	writeErr := os.WriteFile("big.blob", bigEndianBuf.Bytes(), 0666)
	if writeErr != nil {
		log.Println("error outputting blob")
	}
	log.Println("wrote blob to file")

	return nil
}

func (n *NodeService) HandleEvidence(nodeID string, evidenceBlob *bytes.Buffer, signatureBlob *bytes.Buffer) error {
	node, err := n.repo.GetNodeById(nodeID)
	if err != nil {
		return err
	}

	log.Println(node)

	/* Similar to Golden endpoint, parse evidenceBlog as goldenBlob and signatureBlob the same */

	// First we need to see if the data source is legit
	_, err = verifySignature(node.AK_Pub, signatureBlob)
	if err != nil {
		return err
	}

	// TODO: parse TPMS_ATTEST

	return nil
}

var ErrorPEMDecode = errors.New("not found")
var ErrorPEMNotPublicKey = errors.New("pem block is not a public key type")
var ErrorMarshallingPublicKey = errors.New("error marshalling public key type")

func verifySignature(pemStr string, evidenceBlob *bytes.Buffer) (bool, error) {
	akPemBlock, _ := pem.Decode([]byte(pemStr))
	if akPemBlock == nil {
		log.Println("failed to parse PEM block containing the public key")
		return false, ErrorPEMDecode
	}

	if akPemBlock.Type != "PUBLIC KEY" {
		log.Println("akPemBlock.Type != PUBLIC KEY, but: ", akPemBlock.Type)
		return false, ErrorPEMNotPublicKey
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(akPemBlock.Bytes)
	if err != nil {
		log.Println(err.Error())
		return false, ErrorMarshallingPublicKey
	}

	akECDSAPublicKey, _ := x509.ParsePKIXPublicKey(publicKeyBytes)
	switch pub := akECDSAPublicKey.(type) {
	case *rsa.PublicKey:
		log.Println("pub is of type RSA:", pub)
	case *dsa.PublicKey:
		log.Println("pub is of type DSA:", pub)
	case *ecdsa.PublicKey:
		log.Println("pub is of type ECDSA:", pub)
	default:
		log.Println("pub:", pub)
		log.Println("unknown type of public key")
	}

	pubKey, _ := akECDSAPublicKey.(*ecdsa.PublicKey)

	hash := sha256.Sum256(evidenceBlob.Bytes())
	var hashCast []byte = hash[:]

	// Read the r size and chunk from the signature blob
	val := evidenceBlob.Next(2)
	sizeR := binary.LittleEndian.Uint16(val)
	log.Println("Size R: ", sizeR)
	sigR := evidenceBlob.Next(int(sizeR))

	// Read the s size and chunk from the signature blob
	val = evidenceBlob.Next(2)
	sizeS := binary.LittleEndian.Uint16(val)
	log.Println("Size S: ", sizeS)
	sigS := evidenceBlob.Next(int(sizeS))

	numSigR := new(big.Int)
	numSigR.SetBytes(sigR)
	numSigS := new(big.Int)
	numSigS.SetBytes(sigS)

	isValid := ecdsa.Verify(pubKey, hashCast, numSigR, numSigS)
	if isValid {
		log.Println("Signature IS valid")
	} else {
		log.Println("Signature IS NOT valid")
	}

	return isValid, nil
}
