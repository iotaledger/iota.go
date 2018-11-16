# API -> StoreTransactions()
StoreTransactions persists a list of attached transaction trytes in the store of the connected node. Tip-selection and Proof-of-Work must be done first by calling GetTransactionsToApprove and AttachToTangle or an equivalent attach method.  Persist the transaction trytes in local storage before calling this command, to ensure reattachment is possible, until your bundle has been included.  Any transactions stored with this command will eventually be erased as a result of a snapshot.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | ...Trytes | Required | The transaction Trytes to store.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []Trytes | The stored transaction Trytes. |
| error | Returned for internal errors. |



