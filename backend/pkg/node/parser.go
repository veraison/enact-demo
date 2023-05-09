// Copyright 2023 EnactTrust LTD All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and

package node

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/google/uuid"
)

func parseEvidenceAndSignatureBlobs(evidenceBlob *bytes.Buffer, signatureBlob *bytes.Buffer) (*bytes.Buffer, uuid.UUID, error) {
	bigEndianBuf := &bytes.Buffer{}

	log.Println(`Beginning evidence and signature processing`)

	/* Parse TPMS_ATTEST */

	// First 16 bytes are the node_id

	// Read 16 byte node_id from the beginning of the whole blob
	val := evidenceBlob.Next(16)
	uuidNodeId, err := uuid.FromBytes(val)
	if err != nil {
		log.Println(err)
		return nil, uuid.UUID{}, err
	}
	fmt.Sprintf("%x", uuidNodeId)

	// 0. TPMS_ATTEST Size
	val = evidenceBlob.Next(2)
	tpmsAttestSize := binary.LittleEndian.Uint16(val)
	// Convert to Big Endian
	sizeBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeBuf, tpmsAttestSize)
	// Write the Big Endian byte array to the buffer
	binary.Write(bigEndianBuf, binary.BigEndian, sizeBuf)

	binary.Write(bigEndianBuf, binary.BigEndian, evidenceBlob.Next(int(tpmsAttestSize)))

	// /* Parse signature struct */
	// // https://dox.ipxe.org/structTPMT__SIGNATURE.html

	// 0. TPMI_ALG_SIG_SCHEME
	val = signatureBlob.Next(2)
	sigAlgId := binary.LittleEndian.Uint16(val)
	// Convert to Big Endian
	tempBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(tempBuf, sigAlgId)
	// Write the Big Endian byte array to buffer
	binary.Write(bigEndianBuf, binary.BigEndian, tempBuf)

	val = signatureBlob.Next(2)
	tpmiAlgHash := binary.LittleEndian.Uint16(val)
	// Swap endiannnes of tpmiAlgHash
	tempBuf = make([]byte, 2)
	binary.BigEndian.PutUint16(tempBuf, tpmiAlgHash)
	binary.Write(bigEndianBuf, binary.BigEndian, tempBuf)

	// Parsing signature R
	val = signatureBlob.Next(2)
	sizeR := binary.LittleEndian.Uint16(val)
	sigR := signatureBlob.Next(int(sizeR))

	// Swap endiannnes of sizeR
	tempBuf = make([]byte, 2)
	binary.BigEndian.PutUint16(tempBuf, sizeR)
	// Size field
	binary.Write(bigEndianBuf, binary.BigEndian, tempBuf)
	// SigR part
	binary.Write(bigEndianBuf, binary.BigEndian, sigR)

	// Parsing signature S
	val = signatureBlob.Next(2)
	sizeS := binary.LittleEndian.Uint16(val)
	sigS := signatureBlob.Next(int(sizeS))

	// Swap endiannnes of sizeS
	tempBuf = make([]byte, 2)
	binary.BigEndian.PutUint16(tempBuf, sizeS)
	// Size field
	binary.Write(bigEndianBuf, binary.BigEndian, tempBuf)
	// SigS part
	binary.Write(bigEndianBuf, binary.BigEndian, sigS)

	log.Println(`Finished processing evidence with byte size`, len(bigEndianBuf.Bytes()))

	return bigEndianBuf, uuidNodeId, nil
}
