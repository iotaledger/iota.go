package mamgoiota

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnectionSend(t *testing.T) {
	assert := assert.New(t)

	c, err := NewConnection("http://node02.iotatoken.nl:14265", "SIERTBRUINSISBEZIGOMEENRONDJESAMENMETWIMAMENTTEMAKENOMZODESUBSIDIERONDTEKRIJGENH9")
	assert.Nil(err)

	var someJSON struct {
		Id        int
		Message   string
		Timestamp time.Time
	}

	someJSON.Id = 12345
	someJSON.Message = "Hello world this is a JSON"
	someJSON.Timestamp = time.Now()

	stringifiedJSON, err := json.Marshal(someJSON)
	assert.Nil(err)

	id, err := Send("RQP9IFNFGZGFKRVVKUPMYMPZMAICIGX9SVMBPNASEBWJZZAVDCMNOFLMRMFRSQVOQGUVGEETKYFCUPNDDWEKYHSALY", 0, string(stringifiedJSON), c)
	assert.Nil(err)

	t.Logf("TransactionId: %v\n", id)
}

func TestConnectionReadTransactions(t *testing.T) {
	assert := assert.New(t)

	c, err := NewConnection("http://node02.iotatoken.nl:14265", "")
	assert.Nil(err)

	ts, err := ReadTransactions("RQP9IFNFGZGFKRVVKUPMYMPZMAICIGX9SVMBPNASEBWJZZAVDCMNOFLMRMFRSQVOQGUVGEETKYFCUPNDDWEKYHSALY", c)
	assert.Nil(err)
	for i, tr := range ts {
		t.Logf("%d. %v: %d IOTA, %v to %v\n", i+1, tr.Timestamp, tr.Value, tr.Message, tr.Recipient)
	}
}
func TestConnectionReadSingleTransaction(t *testing.T) {
	assert := assert.New(t)

	c, err := NewConnection("http://node02.iotatoken.nl:14265", "")
	assert.Nil(err)

	tx, err := ReadTransaction("QFLSB9PFUYYCKUJ9JWIIHVQPZOOOQPDXMCGWAZCGLCBTODRJJQHZ9BIUEBGMNDFYOJMFGPQOUKBJ99999", c)
	assert.Nil(err)
	t.Logf("%v: %d IOTA, %v to %v\n", tx.Timestamp, tx.Value, tx.Message, tx.Recipient)
}
