package veraison

import (
	"log"

	"github.com/veraison/apiclient/provisioning"
)

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
