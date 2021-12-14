package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/veraison/corim/comid"
	"github.com/veraison/swid"
)

var (
	corimTemplate = `
{
  "corim-id": "11111111-1111-1111-1111-111111111111",
    "profiles": [
      "https://enacttrust.com/veraison/1.0.0"
    ]
}	
`
	gvComidTemplate = `
{
  "tag-identity": {
    "id": "00000000-0000-0000-0000-000000000000"
  },
  "entities": [
    {
      "name": "EnactTrust",
      "regid": "https://enacttrust.com",
      "roles": [ "tagCreator", "creator", "maintainer" ]
    }
  ],
  "triples": {
    "reference-values": [
      {
        "environment": {
          "instance": {
            "type": "uuid",
            "value": "97ACE407-DE30-0000-0000-097ACE407DE3"
          }

        },
        "measurements": [
          {
          }
        ]
      }
    ]
  }
}
`
	avkComidTemplate = `
{
  "tag-identity": {
    "id": "00000000-0000-0000-0000-000000000000"
  },
  "entities": [
    {
      "name": "EnactTrust",
      "regid": "https://enacttrust.com",
      "roles": [ "tagCreator", "creator", "maintainer" ]
    }
  ],
  "triples": {
    "attester-verification-keys": [
      {
        "environment": {
          "instance": {
            "type": "uuid",
            "value": "97ACE407-DE30-0000-0000-097ACE407DE3"
          }
        },
        "verification-keys": [
          { "key": "PLACEHOLDER" }
        ]
      }
    ]
  }
}
`
	sampleAKPub     = `MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6Vwqe7hy3O8Ypa+BUETLUjBNU3rEXVUyt9XHR7HJWLG7XTKQd9i1kVRXeBPDLFnfYru1/euxRnJM7H9UoFDLdA==`
	sampleNodeID, _ = uuid.Parse(`ffffffff-ffff-ffff-ffff-ffffffffffff`)
	sampleAlgID     = swid.Sha256
	sampleDigest    = comid.MustHexDecode(nil, "e45b72f5c0c0b572db4d8d3ab7e97f368ff74e62347a824decb67a84e5224d75")
)

func repackageAKPub() {
	c, err := attesterVerificationKey(sampleAKPub, sampleNodeID, avkComidTemplate, corimTemplate)
	if err != nil {
		log.Fatal(err)
	}

	out, err := c.ToCBOR()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n>> AK_pub: %x\n", out)
}

func repackageGoldenValues() {
	c, err := goldenValues(sampleAlgID, sampleDigest, sampleNodeID, gvComidTemplate, corimTemplate)
	if err != nil {
		log.Fatal(err)
	}

	out, err := c.ToCBOR()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n>> golden values: %x\n", out)
}

func main() {
	repackageAKPub()
	repackageGoldenValues()
}
