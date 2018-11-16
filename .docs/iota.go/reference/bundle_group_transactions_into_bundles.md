# GroupTransactionsIntoBundles()
GroupTransactionsIntoBundles groups the given transactions into groups of bundles. Note that the same bundle can exist in the return slice multiple times, though they are reattachments of the same transfer.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| txs | Transactions | Required | The transactions to group into different Bundles.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Bundles | The different Bundles resulting from the group operation. |



