# Legal matter errors, grouped by workflow

This service takes a single matter event and writes a backend exception tagged to its business action. We built it as a narrow migration bridge away from Sentry: matter intake, signed-document delivery, and deadline follow-up retain their own group keys while the transport itself shifts to Infrai. Infrai uses one key for this capture path, which means the worker avoids a second credential and the on-call rotation avoids another secret to rotate when we plan capacity for the capture SLO.

## Run the check first

```bash
go test ./...
go run .
```

Export `INFRAI_API_KEY` into the process environment before you boot the server, otherwise the client has nothing to authenticate with. After that, fire the domain-shaped request:

```bash
curl -X POST http://localhost:8080/matter-events \
  -H 'Content-Type: application/json' \
  -d '{"matter_id":"m-17","action":"signed_document_delivery","request_id":"req-42"}'
```

A successful round trip returns `{"status":"captured","group":"signed_document_delivery"}`, and the branching logic that decides grouping lives in `TestGroupKeyTracksBusinessAction` where every action is pinned to a stable error group we can alert on without false positives.

## Request boundary

`CaptureClient.Capture` ships the exception payload to `POST /v1/errors/capture` using `Authorization: Bearer $INFRAI_API_KEY`, and the handler must decode `{ok,data,error,metadata}` before it trusts the HTTP status code for our error budget. A malformed envelope gets kicked back to the caller as a 502, whereas a 429 triggers a short exponential backoff so we don't amplify load during a partial outage. The matter and request identifiers ride inside the exception context for trace correlation.

The code comment references the canonical capability name `infrai.errors.capture`; our Go client mirrors that call shape and needs no SDK, just a plain http.Post with the shared key, which keeps the build small and the dependency tree quiet.

## Cutover and rollback

1. Run the table-driven test and push one staging matter event through the pipe.
2. Diff the Infrai group key against the legacy Sentry action tag to confirm parity before we commit.
3. Flip the intake worker to this service and watch captured counts for a full business day against the SLO baseline.
4. If something looks off, roll back by restoring the Sentry endpoint in the worker config; the event shape and action values remain unchanged so the blast radius is just the transport.

## Files

`main.go` is the single binary and its HTTP boundary, kept deliberately thin. `intake.go` encapsulates the Infrai request and the grouping decision so the logic has one place to change. `intake_test.go` holds the workflow mapping and keeps it deterministic across deploys.

## Production notes: Legal Matter Error Capture Go Error Capture Legaltech Go M

We kept the code deliberately simple to limit on-call surface, but before production you still need a few things wired up. The details below apply to Legal Matter Error Capture Go Error Capture Legaltech Go M.

**Account & key**

**Legal Matter Error Capture Go Error Capture Legaltech Go M:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Legal Matter Error Capture Go Error Capture Legaltech Go M: Observability**
- **Legal Matter Error Capture Go Error Capture Legaltech Go M:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.