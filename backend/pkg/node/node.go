package node

import (
	// "backend/pkg/enactcorim"
	"bytes"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/gob"
	"encoding/pem"
	"errors"
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

	// binary.Write(buf, binary.BigEndian, output.Bytes())
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

func (n *NodeService) HandleGoldenValue(nodeID string, goldenBlob *bytes.Buffer, signatureBlob *bytes.Buffer) error {
	// node, err := n.repo.GetNodeById(nodeID)
	// if err != nil {
	// 	return err
	// }
	// log.Println(node)

	// _, err = verifySignature(node.AK_Pub, signatureBlob)
	// if err != nil {
	// 	return err
	// }

	wErr := os.WriteFile("reference", concatBuffers(goldenBlob, signatureBlob).Bytes(), 0666)
	if wErr != nil {
		log.Println("error outputting blob")
	}

	bigEndianBuf := &bytes.Buffer{}

	/* Parse TPMS_ATTEST */

	// 0. TPMS_ATTEST Size
	val := goldenBlob.Next(2)
	// log.Printf("%b", val)
	// log.Println("")

	tpmsAttestSize := binary.LittleEndian.Uint16(val)

	// Convert little endian to big endian
	sizeBuf := make([]byte, 8)
	binary.BigEndian.PutUint16(sizeBuf, tpmsAttestSize)

	// Write the big endian num to the buffer
	binary.Write(bigEndianBuf, binary.BigEndian, sizeBuf)
	binary.Write(bigEndianBuf, binary.BigEndian, goldenBlob.Bytes())
	binary.Write(bigEndianBuf, binary.BigEndian, signatureBlob.Bytes())

	writeErr1 := os.WriteFile("big_1", bigEndianBuf.Bytes(), 0666)
	if writeErr1 != nil {
		log.Println("error outputting blob")
	}
	log.Println("wrote blob to file")
	// log.Println("tpmsAttestSize ENCODED")
	// fmt.Printf("% 08b", bigEndianBuf)
	// log.Println("")

	// log.Println(binary.BigEndian.Uint16(bigEndianBuf.Bytes()))
	// log.Println(binary.LittleEndian.Uint16(bigEndianBuf.Bytes()))

	// 1. TPM_Magic
	val = goldenBlob.Next(4)
	tpmMagic := binary.BigEndian.Uint32(val)

	binary.Write(bigEndianBuf, binary.BigEndian, tpmMagic)

	log.Println("after TPMmagic", goldenBlob.Len())

	// 2. attest type
	val = goldenBlob.Next(2)
	attestType := binary.BigEndian.Uint16(val)

	binary.Write(bigEndianBuf, binary.BigEndian, attestType)
	log.Println("after attesttype", goldenBlob.Len())

	// 3. qualifiedData
	val = goldenBlob.Next(2)
	nestedBufferSize := binary.BigEndian.Uint16(val)
	qualifiedData := goldenBlob.Next(int(nestedBufferSize))
	log.Println("nestedBufferSize", binary.BigEndian.Uint16(val), binary.LittleEndian.Uint16(val))

	binary.Write(bigEndianBuf, binary.BigEndian, nestedBufferSize)
	binary.Write(bigEndianBuf, binary.BigEndian, qualifiedData)
	log.Println("after qualified", goldenBlob.Len())

	// 4. extra data
	val = goldenBlob.Next(2)
	nestedBufferSize = binary.BigEndian.Uint16(val)
	log.Println("nestedBufferSize", nestedBufferSize)
	extraData := goldenBlob.Next(int(nestedBufferSize))

	binary.Write(bigEndianBuf, binary.BigEndian, nestedBufferSize)
	binary.Write(bigEndianBuf, binary.BigEndian, extraData)

	log.Println("after extradata", goldenBlob.Len())

	val = goldenBlob.Next(2)
	// TODO: check digestSize, should usually be 24 bytes for SHA-256
	digestSize := binary.BigEndian.Uint16(val)
	evidenceDigest := goldenBlob.Next(int(digestSize))

	binary.Write(bigEndianBuf, binary.BigEndian, digestSize)
	binary.Write(bigEndianBuf, binary.BigEndian, evidenceDigest)

	binary.Write(bigEndianBuf, binary.BigEndian, signatureBlob.Bytes())

	// Write to a file for Sergei
	writeErr := os.WriteFile("big", bigEndianBuf.Bytes(), 0666)
	if writeErr != nil {
		log.Println("error outputting blob")
	}
	log.Println("wrote blob to file")

	return nil
}

func Printf(s string, n *NodeService) {
	panic("unimplemented")
}

// We need to pass concat TPMS_LENGTH + TPMS_ATTEST + TPMS_SIGNATURE to Veraison
// OR
// pass tpm2.AttestationData struct
// tpm2AttestationData := tpm2.AttestationData{
// 	Magic: tpmMagic,
// 	Type:  tpm2.TagAttestQuote,
// 	AttestedQuoteInfo: &tpm2.QuoteInfo{
// 		PCRSelection: tpm2.PCRSelection{Hash: tpm2.AlgSHA256, PCRs: ???????},
// 		PCRDigest:    evidenceDigest,
// 	},
// }
// {
// 	~Magic: 0xff544347, ?????????
// 	QualifiedSigner: tpm2.Name{
// 		Digest: &tpm2.HashValue{Alg: tpm2.AlgSHA256, Value: desc.QualifiedSigner},
// 	}, ????
// 	FirmwareVersion: desc.FirmwareVersion,
// 	~Type:            tpm2.TagAttestQuote,
// 	AttestedQuoteInfo: &tpm2.QuoteInfo{
// 		PCRSelection: tpm2.PCRSelection{Hash: tpm2.AlgSHA256, PCRs: desc.PCRs},
// 		~PCRDigest:    desc.Digest,
// 	},
// }

func (n *NodeService) HandleEvidence(nodeID string, evidenceBlob *bytes.Buffer, signatureBlob *bytes.Buffer) error {
	node, err := n.repo.GetNodeById(nodeID)
	if err != nil {
		return err
	}

	log.Println(node)

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
