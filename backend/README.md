# EnactTrust - Veraison Backend

## Misc

1. From agent: `POST /node/pem, Body: { nodeID, AK_pub, EK_pub }`
2. Generate node_id (UUID v4)
3. Store node_id, AK_pub, EK_pub
4. Repackage node_id and AK pub as CoRIM
5. `POST /submit, Body: { CoRIM }` to veraison backend and forward response to agent