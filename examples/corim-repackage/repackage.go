package main

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
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
func attesterVerificationKey(
	akPub string, nodeID uuid.UUID, comidTemplate, corimTemplate string,
) (*corim.UnsignedCorim, error) {
	c := comid.Comid{}

	if err := c.FromJSON([]byte(comidTemplate)); err != nil {
		return nil, fmt.Errorf("parsing CoMID JSON template: %s (%w)", comidTemplate, err)
	}

	avk := (*c.Triples.AttestVerifKeys)[0]

	if avk.Environment.Instance.SetUUID(nodeID) == nil {
		return nil, fmt.Errorf("cannot set nodeID")
	}

	if avk.VerifKeys[0].SetKey(akPub) == nil {
		return nil, fmt.Errorf("cannot set AK_pub")
	}

	return buildCorim(corimTemplate, &c)
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

	if gv.Measurements[0].AddDigest(algID, digest) == nil {
		return nil, fmt.Errorf("cannot set golden value")
	}

	return buildCorim(corimTemplate, &c)
}
