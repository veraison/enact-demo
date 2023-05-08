# EnactTrust - Veraison Backend

Launch using

```
go run ./main.go
```

Make sure you have [Veraison services](https://github.com/veraison/services/) running and the [EnactTrust agent](https://github.com/EnactTrust/enact) installed.

## Misc

### Onboarding

1. From agent: `POST /node/pem, Body: { nodeID, AK_pub, EK_pub }`
2. Generate node_id (UUID v4)
3. Store node_id, AK_pub, EK_pub
4. Repackage node_id and AK pub as CoRIM
5. `POST /submit, Body: { CoRIM }` to veraison backend and forward response to agent

### Evidence

// Table 116 - TPMS_ATTEST Structure
 typedef struct {
   TPM_GENERATED   magic;
   TPMI_ST_ATTEST  type;
   TPM2B_NAME      qualifiedSigner;
   TPM2B_DATA      extraData;
   TPMS_CLOCK_INFO clockInfo;
   UINT64          firmwareVersion;
   TPMU_ATTEST     attested;
 } TPMS_ATTEST;
 
 // Table 117 - TPM2B_ATTEST Structure
 typedef struct {
   UINT16 size;
   BYTE   attestationData[sizeof(TPMS_ATTEST)];
 } TPM2B_ATTEST;
