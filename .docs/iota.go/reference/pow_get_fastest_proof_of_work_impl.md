# GetFastestProofOfWorkImpl()
GetFastestProofOfWorkImpl returns the fastest Proof-of-Work implementation. All returned Proof-of-Work implementations returned are "sync", meaning that they only run one Proof-of-Work task simultaneously.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.




## Output

| Return type     | Description |
|:---------------|:--------|
| string | The name of the Proof-of-Work function. |
| ProofOfWorkFunc | The actual Proof-of-Work implementation. |



