package veraison

import (
	"fmt"
	"log"

	"github.com/veraison/apiclient/common"
	"github.com/veraison/apiclient/provisioning"
	"github.com/veraison/apiclient/verification"
)

// TODO: this needs to contain the concatenated nodeID in the beginning of
// TPMS_ATTEST. This should be called after processing the evidence
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
		// TODO: replace this with an env var URL
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

// TODO: SESSION - call api-agent - should get 201 response from Veraison frontend
func CreateVeraisonSession() (*verification.ChallengeResponseSession, string, error) {
	// localhost?
	var sessionURI = "http://veraison.example/challenge-response/v1/newSession"

	cfg := verification.ChallengeResponseConfig{
		NonceSz:       32,
		NewSessionURI: sessionURI,
		Client:        common.NewClient(),
		DeleteSession: true,
	}

	newSession, sessionURI, err := cfg.NewSession()
	if err != nil {
		return nil, "", fmt.Errorf("new session failed: %v", err)
	}

	return newSession, sessionURI, nil
}

func EncryptChallenge(nonce []byte) ([]byte, error) {
	var encryptedChallenge = []byte{}

	// TODO: encrypt

	return encryptedChallenge, nil
}
