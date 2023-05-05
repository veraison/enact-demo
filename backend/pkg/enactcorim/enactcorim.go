package enactcorim

import (
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/swid"
)

var (
	CorimTemplate = `
{
  "corim-id": "11111111-1111-1111-1111-111111111111",
    "profiles": [
      "https://enacttrust.com/veraison/1.0.0"
    ]
}
`
	AKComidTemplate = `
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
)

func buildCorim(template string, c *comid.Comid) (*corim.UnsignedCorim, error) {
	u := corim.UnsignedCorim{}

	if c == nil {
		return nil, errors.New("nil CoMID")
	}

	if err := u.FromJSON([]byte(template)); err != nil {
		return nil, fmt.Errorf("parsing CoRIM JSON template: %s (%w)", template, err)
	}

	if u.AddComid(*c) == nil {
		return nil, fmt.Errorf("cannot create unsigned CoRIM")
	}

	return &u, nil
}

// No belts and braces (assumes a precise template shape)
func RepackageNodePEM(akPub string, nodeID uuid.UUID) (*corim.UnsignedCorim, error) {
	c := comid.Comid{}

	if err := c.FromJSON([]byte(AKComidTemplate)); err != nil {
		return nil, fmt.Errorf("parsing CoMID JSON template: %s (%w)", AKComidTemplate, err)
	}

	avk := (*c.Triples.AttestVerifKeys)[0]

	if avk.Environment.Instance.SetUUID(nodeID) == nil {
		return nil, fmt.Errorf("cannot set nodeID")
	}

	if avk.VerifKeys[0].SetKey(akPub) == nil {
		return nil, fmt.Errorf("cannot set AK_pub")
	}

	return buildCorim(CorimTemplate, &c)
}

func goldenValues(
	algID uint64, digest []byte, nodeID uuid.UUID, comidTemplate, corimTemplate string,
) (*corim.UnsignedCorim, error) {
	c := comid.Comid{}

	if err := c.FromJSON([]byte(comidTemplate)); err != nil {
		return nil, fmt.Errorf("parsing CoMID JSON template: %s (%w)", comidTemplate, err)
	}

	gv := (*c.Triples.ReferenceValues)[0]

	if gv.Environment.Instance.SetUUID(nodeID) == nil {
		return nil, fmt.Errorf("cannot set nodeID")
	}

    fmt.Println(fmt.Sprintf("algID=0x%x len of digest = %d \n digest = 0x%x", algID, len(digest)))
	if gv.Measurements[0].AddDigest(algID, digest) == nil {
		return nil, fmt.Errorf("cannot set golden value")
	}

	return buildCorim(corimTemplate, &c)
}

func RepackageEvidence(nodeID uuid.UUID, evidenceDigest []byte) ([]byte, error) {
	var algID = swid.Sha256
	corim, err := goldenValues(algID, evidenceDigest, nodeID, gvComidTemplate, CorimTemplate)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	log.Println(`successfully repacked evidence as corim`)

	cbor, err := corim.ToCBOR()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	log.Println(`successfully converted corim to cbor`)

	fmt.Printf("\n>> golden values: %x\n", cbor)
	return cbor, nil
}
