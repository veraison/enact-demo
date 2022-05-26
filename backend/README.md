# EnactTrust - Veraison Backend

## Misc

1. From agent: `POST /node/pem, Body: { nodeID, AK_pub, EK_pub }`
2. Generate node_id (UUID v4)
3. Store node_id, AK_pub, EK_pub
4. Repackage node_id and AK pub as CoRIM
5. `POST /submit, Body: { CoRIM }` to veraison backend and forward response to agent


https://medium.com/webauthnworks/verifying-fido-tpm2-0-attestation-fc7243847498

https://dox.ipxe.org/structTPMS__ATTEST.html


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


https://dox.ipxe.org/structTPM2B__ATTEST.html

https://github.com/EnactTrust/enact/blob/8f4d4f741238e015d51aaa9d97ab0a7b94b635a0/enact.h#L93

https://github.com/EnactTrust/enact/blob/8f4d4f741238e015d51aaa9d97ab0a7b94b635a0/agent.c#L352


https://github.com/EnactTrust/enact/blob/8f4d4f741238e015d51aaa9d97ab0a7b94b635a0/agent.c#L387


little: 1110001
big: 01001101

### Print formatted bytes
fmt.Printf("% 08b", bigEndianBuf)

log.Printf("%b", val)