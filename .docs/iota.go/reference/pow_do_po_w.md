# DoPoW()
DoPoW computes the nonce field for each transaction so that the last MWM-length trits of the transaction hash are all zeroes. Starting from the 0 index transaction, the transactions get chained to each other through the trunk transaction hash field. The last transaction in the bundle approves the given branch and trunk transactions. This function also initializes the attachment timestamp fields. This function expects the passed in transaction trytes from highest to lowest index, meaning the transaction with current index 0 at the last position.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trunkTx | Trytes | Required | The trunk transaction hash.  |
| branchTx | Trytes | Required | the branch transaction hash.  |
| trytes | []Trytes | Required | The transactions Trytes slice.  |
| mwm | uint64 | Required | The minimum weight magnitude to fulfill.  |
| pow | ProofOfWorkFunc | Required | The Proof-of-Work implementation function.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []Trytes | The transaction Trytes with computed nonce fields, ready for broadcasting. |
| error | Returned for invalid transaction Trytes and internal errors. |



