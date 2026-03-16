# Task: Add Device Identity Auth

The OpenClaw gateway requires DEVICE IDENTITY for write scopes. Without it, chat.send gets "missing scope: operator.write".

Read the current gateway client code in internal/gateway/client.go and protocol.go.

## What Needs to Change

The connect params need a "device" field with a signed payload.

### Device Identity File
Located at ~/.openclaw/identity/device.json:
```json
{
  "version": 1,
  "deviceId": "hex string",
  "publicKeyPem": "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----\n",
  "privateKeyPem": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
}
```
This is an Ed25519 keypair.

### Signed Payload
The client must sign a compact JSON payload with the private key. The payload fields must be in this exact order:
```
{"clientId":"gateway-client","clientMode":"backend","role":"operator","scopes":["operator.admin"],"signedAtMs":1234567890,"token":null,"nonce":"<nonce-from-challenge>","platform":"windows","deviceFamily":null}
```

The signature is: base64url(Ed25519Sign(privateKey, payloadJsonBytes))

### Connect Params with Device
```json
{
  "minProtocol": 3,
  "maxProtocol": 3,
  "client": {
    "id": "gateway-client",
    "displayName": "openclaw-tui",
    "version": "0.1.0",
    "platform": "windows",
    "mode": "backend"
  },
  "caps": [],
  "auth": {"token": "<token>"},
  "role": "operator",
  "scopes": ["operator.admin"],
  "device": {
    "id": "<deviceId from device.json>",
    "publicKey": "<base64url raw 32-byte public key>",
    "signature": "<base64url Ed25519 signature>",
    "signedAt": 1234567890,
    "nonce": "<nonce from challenge>"
  }
}
```

### base64url encoding
Standard base64 but replace + with -, / with _, and strip trailing = padding.

### Public Key Format
The connect sends the RAW public key bytes (32 bytes for Ed25519) in base64url, NOT the full PEM. 
Extract from PEM: parse PEM block -> parse as PKIX/DER -> last 32 bytes are the raw Ed25519 public key.

## Implementation

1. Create internal/gateway/identity.go with:
   - DeviceIdentity struct
   - LoadDeviceIdentity(path) to read the JSON file
   - DefaultIdentityPath() returning ~/.openclaw/identity/device.json
   - SignConnectPayload() to create and sign the device payload
   - base64UrlEncode helper
   - extractRawPublicKey to get 32 raw bytes from PEM

2. Update client.go Connect():
   - Add DeviceIdentity field to Client struct
   - Load device identity in NewClient or Connect
   - Store the challenge nonce from the connect.challenge event
   - Build signed device payload using the nonce
   - Include "device" field in connect params

3. Update protocol.go:
   - Add DeviceInfo struct with ID, PublicKey, Signature, SignedAt, Nonce fields
   - Add DeviceInfo field to ConnectParams (json:"device,omitempty")

4. Update cmd/openclaw-tui/main.go:
   - Load device identity path
   - Pass it to gateway.NewClient

CRITICAL:
- go build ./... must succeed
- go vet ./... must pass
- Use Go's crypto/ed25519 for signing
- Parse PEM with encoding/pem, then crypto/x509 for PKIX
- The signed payload JSON must be compact with exact field order
- Test: run cmd/debug with the device identity and verify chat.send works
- Commit when done
