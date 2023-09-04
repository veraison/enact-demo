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

package veraison

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/veraison/apiclient/common"
	"github.com/veraison/apiclient/provisioning"
	"github.com/veraison/apiclient/verification"
	"github.com/veraison/ear"
)

var TPMEvidenceMediaType = "application/vnd.enacttrust.tpm-evidence"
var VeraisonECDSAPublicKey = `{
	"kty": "EC",
	"crv": "P-256",
	"x": "usWxHK2PmfnHKwXPS54m0kTcGJ90UiglWiGahtagnv8",
	"y": "IBOL-C3BttVivg-lSreASjpkttcsz-1rb7btKLv8EX4"
}`

func SendTPMEvidenceToVeraison(cbor []byte) error {
	// vnd-enacttrust.tpm-evidence from the diagram
	return nil
}

func SendCborToVeraison(cbor []byte) error {
	// This uses the Veraison api-client
	cfg := provisioning.SubmitConfig{
		SubmitURI: "http://localhost:8888/endorsement-provisioning/v1/submit",
	}
	// The Run method is invoked on the instantiated SubmitConfig object to
	// trigger the protocol FSM, hiding any details about the synchronus / async nature
	// of the underlying exchange.  The user must supply the byte buffer containing the
	// serialized endorsement, and the associated media type:
	err := cfg.Run(cbor, "application/corim-unsigned+cbor; profile=http://enacttrust.com/veraison/1.0.0")
	if err != nil {
		log.Println(err.Error())
		return err
	}

	if err != nil {
		log.Println(err.Error())
		return err
	}

	log.Println("CORIM cbor successfully sent to Veraison")

	return nil
}

// this corresponds to phase2 from
// https://github.com/veraison/enact-demo/blob/38e97e32d302d8627489de6127839d4929dfc819/examples/async-apiclient/main.go#L51
func SendEvidenceAndSignature(cfg *verification.ChallengeResponseConfig, sessionId string, data []byte) ([]byte, error) {
	// TODO: check if the session exists
	// if !ok {
	// 	return nil, fmt.Errorf("no session URI found for node %q", FakeNodeID)
	// }
	log.Print("sessionID")
	log.Println(sessionId)

	log.Println("sessionId: ", sessionId)
	log.Println("data length: ", len(data))

	// extract golden values from body
	fmt.Printf("\n%x\n", data)
	attestationResultRawMessage, err := cfg.ChallengeResponse(data, TPMEvidenceMediaType, sessionId)
	if err != nil {
		return nil, fmt.Errorf("challenge-response session failed: %v", err)
	}

	var attestationResultJWT string
	if err = json.Unmarshal(attestationResultRawMessage, &attestationResultJWT); err != nil {
		return nil, fmt.Errorf("challenge-response result decoding failed: %v", err)
	}

	return []byte(attestationResultJWT), nil
}

// This does POST /submit, Body: { CoRIM }`

func RepackageEvidenceAndSendToVeraison(cfg *verification.ChallengeResponseConfig) {
	// TODO: repackage as golden value CoRIM

	// TODO: POST /submit
}

func CreateVeraisonSession() (*verification.ChallengeResponseConfig, *verification.ChallengeResponseSession, string, error) {
	// localhost?
	var sessionURI = "http://localhost:8080/challenge-response/v1/newSession"

	cfg := verification.ChallengeResponseConfig{
		NonceSz:       16,
		NewSessionURI: sessionURI,
		Client:        common.NewClient(),
		DeleteSession: true,
	}

	newSession, sessionURI, err := cfg.NewSession()
	if err != nil {
		return nil, nil, "", fmt.Errorf("new session failed: %v", err)
	}

	return &cfg, newSession, sessionURI, nil
}

// This is the attestation result check
func EarCheck(b []byte) error {
	k, _ := jwk.ParseKey([]byte(VeraisonECDSAPublicKey))

	var r ear.AttestationResult

	s := fmt.Sprintf("%x", b)
	fmt.Println(s)
	log.Println("Length of EarCheck byte slice=", len(b))

	if err := r.Verify(b, jwa.KeyAlgorithmFrom(jwa.ES256), k); err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	appraisal, ok := r.Submods["TPM_ENACTTRUST"]
	if !ok {
		return errors.New("unexpected format: missing TPM_ENACTTRUST submod")
	}

	// at a minimum, one needs to check the overall status
	if *appraisal.Status != ear.TrustTierAffirming {
		return fmt.Errorf(`want "affirming", got %s`, *appraisal.Status)
	}

	return nil
}

func dumpByteSlice(b []byte) {
	var a [16]byte
	n := (len(b) + 15) &^ 15
	for i := 0; i < n; i++ {
		if i%16 == 0 {
			fmt.Printf("%4d", i)
		}
		if i%8 == 0 {
			fmt.Print(" ")
		}
		if i < len(b) {
			fmt.Printf(" %02X", b[i])
		} else {
			fmt.Print("   ")
		}
		if i >= len(b) {
			a[i%16] = ' '
		} else if b[i] < 32 || b[i] > 126 {
			a[i%16] = '.'
		} else {
			a[i%16] = b[i]
		}
		if i%16 == 15 {
			fmt.Printf("  %s\n", string(a[:]))
		}
	}
}
