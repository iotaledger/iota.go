package mongo

import (
	"context"
	"github.com/iotaledger/iota.go/account/store"
	"github.com/iotaledger/iota.go/trinary"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"strconv"
	"time"
)

// ContextProviderFunc which generates a new context.
type ContextProviderFunc func() context.Context

// Config defines settings which are used in conjunction with MongoDB.
type Config struct {
	// The name of the database inside MongoDB.
	DBName string
	// The name of the collection in which to store accounts.
	CollName string
	// The context provider which gets called up on each MongoDB call
	// in order to determine the timeout/cancellation.
	ContextProvider ContextProviderFunc
}

const (
	DefaultDBName   = "iota_account"
	DefaultCollName = "accounts"

	DepositAddressesKey = "deposit_addresses"
	KeyIndexKey         = "key_index"
	PendingTransfersKey = "pending_transfers"
)

func defaultConfig(cnf *Config) *Config {
	if cnf == nil {
		return &Config{
			DBName:          DefaultDBName,
			CollName:        DefaultCollName,
			ContextProvider: defaultCtxProvider,
		}
	}
	if cnf.DBName == "" {
		cnf.DBName = DefaultDBName
	}
	if cnf.CollName == "" {
		cnf.CollName = DefaultCollName
	}
	if cnf.ContextProvider == nil {
		cnf.ContextProvider = defaultCtxProvider
	}
	return cnf
}

func defaultMongoDBConf() []*options.ClientOptions {
	return []*options.ClientOptions{
		{
			WriteConcern: writeconcern.New(writeconcern.J(true), writeconcern.WMajority(), writeconcern.WTimeout(5*time.Second)),
			ReadConcern:  readconcern.Majority(),
		},
	}
}

func defaultCtxProvider() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return ctx
}

// NewMongoStore creates a new MongoDB store. If no MongoDB client options are defined,
// the client will be configured with a majority read/write concern, 5 sec. write concern timeout and journal write acknowledgement.
func NewMongoStore(uri string, cnf *Config, opts ...*options.ClientOptions) (*MongoStore, error) {
	if len(opts) == 0 {
		opts = defaultMongoDBConf()
	}
	opts = append(opts, options.Client().ApplyURI(uri))
	client, err := mongo.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	mongoStore := &MongoStore{client: client, cnf: defaultConfig(cnf)}
	if err := mongoStore.init(); err != nil {
		return nil, err
	}
	return mongoStore, nil
}

type MongoStore struct {
	client *mongo.Client
	coll   *mongo.Collection
	cnf    *Config
}

func (ms *MongoStore) init() error {
	if err := ms.client.Connect(ms.cnf.ContextProvider()); err != nil {
		return err
	}
	if err := ms.client.Ping(ms.cnf.ContextProvider(), nil); err != nil {
		return err
	}
	ms.coll = ms.client.Database(ms.cnf.DBName).Collection(ms.cnf.CollName)
	return nil
}

func newaccountstate() *accountstate {
	return &accountstate{
		DepositAddresses: make(map[string]*store.StoredDepositAddress),
		PendingTransfers: make(map[string]*store.PendingTransfer),
	}
}

// account state with deposit addresses map adjusted to use string keys as
// the marshaller of the MongoDB lib can't marshal uint64 map keys.
type accountstate struct {
	ID               string                                 `bson:"_id"`
	KeyIndex         uint64                                 `json:"key_index" bson:"key_index"`
	DepositAddresses map[string]*store.StoredDepositAddress `json:"deposit_addresses" bson:"deposit_addresses"`
	PendingTransfers map[string]*store.PendingTransfer      `json:"pending_transfers" bson:"pending_transfers"`
}

// FIXME: remove once MongoDB driver knows how to unmarshal uint64 map keys
func (ac *accountstate) AccountState() *store.AccountState {
	state := &store.AccountState{}
	state.PendingTransfers = ac.PendingTransfers
	state.KeyIndex = ac.KeyIndex
	state.DepositAddresses = map[uint64]*store.StoredDepositAddress{}
	for key, val := range ac.DepositAddresses {
		keyNum, _ := strconv.ParseUint(key, 10, 64)
		state.DepositAddresses[keyNum] = val
	}
	return state
}

func (ms *MongoStore) LoadAccount(id string) (*store.AccountState, error) {
	state := newaccountstate()
	cursor := ms.coll.FindOne(ms.cnf.ContextProvider(), bson.D{{"_id", id}})
	if err := cursor.Err(); err != nil {
		// return an empty account state object if there isn't
		// a previously stored account
		if err == mongo.ErrNoDocuments {
			// we build our own account object to allocate the obj inside the db.
			state.ID = id
			_, err := ms.coll.InsertOne(ms.cnf.ContextProvider(), state)
			if err != nil {
				return nil, err
			}
			return store.NewAccountState(), nil
		}
		return nil, err
	}
	if err := cursor.Decode(state); err != nil {
		return nil, err
	}
	return state.AccountState(), nil
}

func (ms *MongoStore) RemoveAccount(id string) error {
	delRes, err := ms.coll.DeleteOne(ms.cnf.ContextProvider(), bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	if delRes.DeletedCount == 0 {
		return store.ErrAccountNotFound
	}
	return nil
}

func (ms *MongoStore) ImportAccount(state store.ExportedAccountState) error {
	stateToImport := newaccountstate()
	stateToImport.KeyIndex = state.KeyIndex
	stateToImport.PendingTransfers = state.PendingTransfers
	stateToImport.ID = state.ID
	for index, depositAddress := range state.DepositAddresses {
		stateToImport.DepositAddresses[strconv.Itoa(int(index))] = depositAddress
	}
	t := true
	opts := &options.ReplaceOptions{Upsert: &t}
	_, err := ms.coll.ReplaceOne(ms.cnf.ContextProvider(), bson.D{{"_id", state.ID}}, stateToImport, opts)
	if err != nil {
		return err
	}
	return nil
}

func (ms *MongoStore) ExportAccount(id string) (*store.ExportedAccountState, error) {
	state := newaccountstate()
	cursor := ms.coll.FindOne(ms.cnf.ContextProvider(), bson.D{{"_id", id}})
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	if err := cursor.Decode(state); err != nil {
		return nil, err
	}
	return &store.ExportedAccountState{ID: id, Date: time.Now(), AccountState: *state.AccountState()}, nil
}

type keyindex struct {
	KeyIndex uint64 `bson:"key_index"`
}

func (ms *MongoStore) ReadIndex(id string) (uint64, error) {
	opts := &options.FindOneOptions{
		Projection: bson.D{
			{"_id", 0},
			{KeyIndexKey, 1},
		},
	}
	res := ms.coll.FindOne(ms.cnf.ContextProvider(), bson.D{{"_id", id}}, opts)
	if res.Err() != nil {
		return 0, res.Err()
	}
	result := &keyindex{}
	if err := res.Decode(result); err != nil {
		return 0, err
	}
	return result.KeyIndex, nil
}

func (ms *MongoStore) WriteIndex(id string, index uint64) error {
	mutation := bson.D{{"$set", bson.D{{KeyIndexKey, index}}}}
	_, err := ms.coll.UpdateOne(ms.cnf.ContextProvider(), bson.D{{"_id", id}}, mutation)
	if err != nil {
		return err
	}
	return nil
}

func (ms *MongoStore) AddDepositAddress(id string, index uint64, depositAddress *store.StoredDepositAddress) error {
	indexStr := strconv.FormatUint(index, 10)
	mutation := bson.D{{"$set", bson.D{{DepositAddressesKey + "." + indexStr, depositAddress}}}}
	_, err := ms.coll.UpdateOne(ms.cnf.ContextProvider(), bson.D{{"_id", id}}, mutation)
	if err != nil {
		return err
	}
	return nil
}

func (ms *MongoStore) RemoveDepositAddress(id string, index uint64) error {
	indexStr := strconv.FormatUint(index, 10)
	mutation := bson.D{{"$unset", bson.D{{DepositAddressesKey + "." + indexStr, ""}}}}
	_, err := ms.coll.UpdateOne(ms.cnf.ContextProvider(), bson.D{{"_id", id}}, mutation)
	if err != nil {
		return err
	}
	return nil
}

type depositaddresses struct {
	DepositAddresses map[string]*store.StoredDepositAddress `bson:"deposit_addresses"`
}

// FIXME: remove once MongoDB driver knows how to unmarshal uint64 map keys
func (dr *depositaddresses) convert() map[uint64]*store.StoredDepositAddress {
	m := make(map[uint64]*store.StoredDepositAddress, len(dr.DepositAddresses))
	for key, val := range dr.DepositAddresses {
		keyNum, _ := strconv.ParseUint(key, 10, 64)
		m[keyNum] = val
	}
	return m
}

func (ms *MongoStore) GetDepositAddresses(id string) (map[uint64]*store.StoredDepositAddress, error) {
	opts := &options.FindOneOptions{
		Projection: bson.D{
			{"_id", 0},
			{DepositAddressesKey, 1},
		},
	}
	res := ms.coll.FindOne(ms.cnf.ContextProvider(), bson.D{{"_id", id}}, opts)
	if res.Err() != nil {
		return nil, res.Err()
	}
	result := &depositaddresses{}
	if err := res.Decode(result); err != nil {
		return nil, err
	}
	return result.convert(), nil
}

func (ms *MongoStore) AddPendingTransfer(id string, originTailTxHash trinary.Hash, bundleTrytes []trinary.Trytes, indices ...uint64) error {
	pendingTransfer := store.TrytesToPendingTransfer(bundleTrytes)
	pendingTransfer.Tails = append(pendingTransfer.Tails, originTailTxHash)
	mutation := bson.D{
		{"$set", bson.D{{PendingTransfersKey + "." + originTailTxHash, pendingTransfer}}},
	}
	if len(indices) > 0 {
		unsetMap := bson.M{}
		for _, index := range indices {
			indexStr := strconv.FormatUint(index, 10)
			unsetMap[DepositAddressesKey+"."+indexStr] = ""
		}
		mutation = append(mutation, bson.E{"$unset", unsetMap})
	}
	_, err := ms.coll.UpdateOne(ms.cnf.ContextProvider(), bson.D{{"_id", id}}, mutation)
	if err != nil {
		return err
	}
	return nil
}

func (ms *MongoStore) RemovePendingTransfer(id string, originTailTxHash trinary.Hash) error {
	mutation := bson.D{{"$unset", bson.D{{PendingTransfersKey + "." + originTailTxHash, ""}}}}
	_, err := ms.coll.UpdateOne(ms.cnf.ContextProvider(), bson.D{{"_id", id}}, mutation)
	if err != nil {
		return err
	}
	return nil
}

func (ms *MongoStore) AddTailHash(id string, originTailTxHash trinary.Hash, newTailTxHash trinary.Hash) error {
	mutation := bson.D{
		{"$addToSet", bson.D{{PendingTransfersKey + "." + originTailTxHash + ".tails", newTailTxHash}}},
	}
	_, err := ms.coll.UpdateOne(ms.cnf.ContextProvider(), bson.D{
		{"_id", id},
		{PendingTransfersKey + "." + originTailTxHash, bson.D{{"$exists", true}}},
	}, mutation)
	if err != nil {
		return err
	}
	return nil
}

type pendingtransfers struct {
	PendingTransfers map[string]*store.PendingTransfer `bson:"pending_transfers"`
}

func (ms *MongoStore) GetPendingTransfers(id string) (map[string]*store.PendingTransfer, error) {
	opts := &options.FindOneOptions{
		Projection: bson.D{
			{"_id", 0},
			{PendingTransfersKey, 1},
		},
	}
	res := ms.coll.FindOne(ms.cnf.ContextProvider(), bson.D{{"_id", id}}, opts)
	if res.Err() != nil {
		return nil, res.Err()
	}
	result := &pendingtransfers{}
	if err := res.Decode(result); err != nil {
		return nil, err
	}
	return result.PendingTransfers, nil
}
