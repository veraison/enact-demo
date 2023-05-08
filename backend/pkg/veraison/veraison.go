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
	"fmt"
	"log"

	"github.com/veraison/apiclient/common"
	"github.com/veraison/apiclient/provisioning"
	"github.com/veraison/apiclient/verification"
)

var TPMEvidenceMediaType = "application/vnd.enacttrust.tpm-evidence"

func SendTPMEvidenceToVeraison(cbor []byte) error {
	// vnd-enacttrust.tpm-evidence from the diagram
	return nil
}

func SendPEMToVeraison(cbor []byte) error {
	// writeErr := os.WriteFile("pem-to-veraison", cbor, 0666)
	// if writeErr != nil {
	// 	log.Println("pem")
	// }
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

	log.Println("CORIM successfully sent to Veraison")

	return nil
}

func SendEvidenceAndSignature(cfg *verification.ChallengeResponseConfig, sessionId string, data []byte) ([]byte, error) {
	// TODO: check if the session exists
	// if !ok {
	// 	return nil, fmt.Errorf("no session URI found for node %q", FakeNodeID)
	// }

	// extract golden values from body
	attestationResultJSON, err := cfg.ChallengeResponse(data, TPMEvidenceMediaType, sessionId)
	if err != nil {
		return nil, fmt.Errorf("challenge-response session failed: %v", err)
	}

	return attestationResultJSON, nil
}

func SendEvidenceCborToVeraison(cbor []byte) error {
	return nil
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
