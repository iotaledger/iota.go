# API -> IsPromotable()
IsPromotable checks if a transaction is promotable by calling the checkConsistency IRI API command and verifying that attachmentTimestamp is above a lower bound. Lower bound is calculated based on the number of milestones issued since transaction attachment.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| tailTxHash | Hash | Required | The tail transaction to check.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| bool | Whether the transaction is promotable. |
| error | Returned for invalid parameters and internal errors. |



