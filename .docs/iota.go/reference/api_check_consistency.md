# API -> CheckConsistency()
CheckConsistency checks if a transaction is consistent or a set of transactions are co-consistent.  Co-consistent transactions and the transactions that they approve (directly or indirectly), are not conflicting with each other and the rest of the ledger.  As long as a transaction is consistent, it might be accepted by the network. In case a transaction is inconsistent, it will not be accepted and a reattachment is required by calling ReplayBundle().
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| hashes | ...Hash | Required | The hashes of the transaction to check the consistency of.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| bool | Whether the transaction(s) are consistent. |
| string | The info message supplied by IRI. |
| error | Returned for invalid transaction hashes and internal errors. |




## Example

```go
func ExampleCheckConsistency() 
	txHash := "DJDMZD9G9VMGR9UKMEYJWYRLUDEVWTPQJXIQAAXFGMXXSCONBGCJKVQQZPXFMVHAAPAGGBMDXESTZ9999"
	consistent, _, err := iotaAPI.CheckConsistency(txHash)
	if err != nil {
		// handle error
		return
	}

	fmt.Println("transaction consistent?", consistent)
}

```
