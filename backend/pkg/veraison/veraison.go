package veraison

import (
	"log"

	"github.com/veraison/corim/corim"
)

func SendPEMToVeraison(corimData *corim.UnsignedCorim) error {
	// TODO: place corim into a byte array, because that's what the http package wants to have for the POST body
	// var buf []byte
	// postBody := bytes.NewBuffer(buf)
	// resp, err := http.Post("", "application/json", postBody)
	// _ = resp
	// defer resp.Body.Close()

	// if err != nil {
	// 	log.Println(err.Error())
	// }

	// veraisonResponse, err := ioutil.ReadAll(resp.Body)
	// _ = veraisonResponse

	log.Println("Sending CORIM to Veraison")

	return nil
}
