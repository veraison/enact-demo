# EnactTrust + Veraison Provisioning

The on-boarding of each node is done with two CoRIMs:

1. for golden values
2. for the AK

## CoRIM

All CoRIMs have their profile field set to `"https://enacttrust.com/veraison/1.0.0"`:

```json
{
  /* unique CoRIM id (random UUID) */
  "corim-id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "profiles": [
    "https://enacttrust.com/veraison/1.0.0"
  ]
}
```

## CoMID(s)

All CoMIDs share a common header:

```json
{
  "tag-identity": {
    /* unique CoMID id (random UUID) */
    "id": "00000000-0000-0000-0000-000000000000"
  },
  "entities": [
    {
      "name": "EnactTrust",
      "regid": "https://enacttrust.com",
      "roles": [
        "tagCreator",
        "creator",
        "maintainer"
      ]
    }
  ],
  "triples": {
    /* Reference values or key material (see below) */
  }
}
```

### Golden Values

All the golden values associated with a certain Node ID -- `ffffffff-ffff-ffff-ffff-ffffffffffff` in the example below -- are carried in one CoMID using a single reference-values triple with as many measurements as there are golden values:

```json
  /* ... */
  "triples": {
    "reference-values": [
      {
        "environment": {
          /* the node ID */
          "instance": {
            "type": "uuid",
            "value": "ffffffff-ffff-ffff-ffff-ffffffffffff"
          }
        },
        /* one or more SHA-256 hashes corresponding to the measured components */
        "measurements": [
          {
            "value": {
              "digests": [
                "sha-256:h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc="
              ]
            }
          },
          {
            "value": {
              "digests": [
                "sha-256:AmOCmYm2/ZVPcrqvL8ZLwuLwHWktTecphuqAj26ZgT8="
              ]
            }
          }
        ]
      }
    ]
  }
```

### Key Material

The public AK associated with the Node ID goes in one CoMID inside an attester-verification-keys triple.  The key must be a PEM encoded SubjectPublicKeyInfo [RFC5280].

```json
  /* ... */
  "triples": {
    "attester-verification-keys": [
      {
        "environment": {
          /* node ID */
          "instance": {
            "type": "uuid",
            "value": "ffffffff-ffff-ffff-ffff-ffffffffffff"
          }
        },
        "verification-keys": [
          {
            /* PEM encoded SubjectPublicKeyInfo containing an ECDSA public key */
            "key": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6Vwqe7hy3O8Ypa+BUETLUjBNU3rEXVUyt9XHR7HJWLG7XTKQd9i1kVRXeBPDLFnfYru1/euxRnJM7H9UoFDLdA=="
          }
        ]
      }
    ]
  }
```

## Complete examples

* Golden values

```shell
$ cat comid-golden.json
```
```json
{
  "tag-identity": {
    "id": "00000000-0000-0000-0000-000000000000"
  },
  "entities": [
    {
      "name": "EnactTrust",
      "regid": "https://enacttrust.com",
      "roles": [
        "tagCreator",
        "creator",
        "maintainer"
      ]
    }
  ],
  "triples": {
    "reference-values": [
      {
        "environment": {
          "instance": {
            "type": "uuid",
            "value": "ffffffff-ffff-ffff-ffff-ffffffffffff"
          }
        },
        "measurements": [
          {
            "value": {
              "digests": [
                "sha-256:h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc="
              ]
            }
          },
          {
            "value": {
              "digests": [
                "sha-256:AmOCmYm2/ZVPcrqvL8ZLwuLwHWktTecphuqAj26ZgT8="
              ]
            }
          }
        ]
      }
    ]
  }
}
```

* AK public key

```shell
$ cat comid-ak.json
```
```json
{
  "tag-identity": {
    "id": "00000000-0000-0000-0000-000000000000"
  },
  "entities": [
    {
      "name": "EnactTrust",
      "regid": "https://enacttrust.com",
      "roles": [
        "tagCreator",
        "creator",
        "maintainer"
      ]
    }
  ],
  "triples": {
    "attester-verification-keys": [
      {
        "environment": {
          "instance": {
            "type": "uuid",
            "value": "ffffffff-ffff-ffff-ffff-ffffffffffff"
          }
        },
        "verification-keys": [
          {
            "key": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6Vwqe7hy3O8Ypa+BUETLUjBNU3rEXVUyt9XHR7HJWLG7XTKQd9i1kVRXeBPDLFnfYru1/euxRnJM7H9UoFDLdA=="
          }
        ]
      }
    ]
  }
}
```

* CoRIM wrapper

```shell
$ cat corim.json
```
```json
{
  "corim-id": "11111111-1111-1111-1111-111111111111",
  "profiles": [
    "https://enacttrust.com/veraison/1.0.0"
  ]
}
```

* use the [`cocli`](https://github.com/veraison/corim/cocli/README.md) command to create the two CBOR encoded CoRIMs, `corim-ak.cbor` and `corim-golden.cbor`:

```shell
$ cocli comid create \
    --template=comid-ak.json
    --output=comid-ak.cbor

$ cocli corim create \
    --template=corim.json \
    --comid=comid-ak.cbor \
    --output=corim-ak.cbor
```

```shell
$ cocli comid create \
    --template=comid-golden.json
    --output=comid-golden.cbor

$ cocli corim create \
    --template=corim.json \
    --comid=comid-golden.cbor \
    --output=corim-golden.cbor
```
