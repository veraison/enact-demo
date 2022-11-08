package main

import (
	"fmt"
	"log"

	"github.com/veraison/apiclient/common"
	"github.com/veraison/apiclient/verification"
)

var (
	EntryPoint           = "http://localhost:8080/challenge-response/v1/newSession"
	SessionTable         = map[string]string{}
	FakeNodeID           = "7dd5db06-d2f5-4e0d-8a9c-9baaa5a446ef"
	FakeGolden           = []byte{0x00, 0x01, 0x02, 0x03}
	TPMEvidenceMediaType = "application/vnd.enacttrust.tpm-evidence"
)

func phase1() (*verification.ChallengeResponseConfig, error) {
	cfg := verification.ChallengeResponseConfig{}

	_ = cfg.SetNonceSz(32)
	_ = cfg.SetSessionURI(EntryPoint)
	_ = cfg.SetClient(common.NewClient())
	cfg.SetDeleteSession(true)

	newSessionCtx, sessionURI, err := cfg.NewSession()
	if err != nil {
		return nil, fmt.Errorf("new session to %s failed: %v", EntryPoint, err)
	}

	// store association between nodeID and the established session URI
	SessionTable[FakeNodeID] = sessionURI

	// finalise the challenge part
	fmt.Printf("here we should send 200 to Node with nonce: %x\n", newSessionCtx.Nonce)

	return &cfg, nil
}

func phase2(cfg *verification.ChallengeResponseConfig) ([]byte, error) {
	su, ok := SessionTable[FakeNodeID]
	if !ok {
		return nil, fmt.Errorf("no session URI found for node %q", FakeNodeID)
	}

	// extract golden values from body

	ar, err := cfg.ChallengeResponse(FakeGolden, TPMEvidenceMediaType, su)
	if err != nil {
		return nil, fmt.Errorf("challenge-response session failed: %v", err)
	}

	return ar, nil
}

func main() {
	cfg, err := phase1()
	if err != nil {
		log.Fatalf("phase1: %v", err)
	}

	// wait for Node to reply to challenge
	// when node POSTs to /node/golden, extract nodeID from body and lookup the
	// active session context

	ar, err := phase2(cfg)
	if err != nil {
		log.Fatalf("phase2: %v", err)
	}

	fmt.Println("attestation result: ", string(ar))
}
