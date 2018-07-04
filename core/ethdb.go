package core

import (
	ethdb "github.com/ethereum/go-ethereum/ethdb"
	dbm "github.com/tendermint/tendermint/libs/db"
)

// EthereumDB implements Ethereum's ethdb.Database and ethdb.Batch interfaces.
// It will be used to facilitate persistence of codeHash => code mappings.
type EthereumDB struct {
	codeDB dbm.DB
}

// Put implements Ethereum's ethdb.Putter interface. It wraps the database
// write operation supported by both batches and regular databases.
func (edb *EthereumDB) Put(key []byte, value []byte) error {
	edb.codeDB.Set(key, value)
	return nil
}

// Get implements Ethereum's ethdb.Database interface. It returns a value for a
// given key.
func (edb *EthereumDB) Get(key []byte) ([]byte, error) {
	return edb.codeDB.Get(key), nil
}

// Has implements Ethereum's ethdb.Database interface. It returns a boolean
// determining if the underlying database has the given key or not.
func (edb *EthereumDB) Has(key []byte) (bool, error) {
	return edb.codeDB.Has(key), nil
}

// Delete implements Ethereum's ethdb.Database interface. It removes a given
// key from the underlying database.
func (edb *EthereumDB) Delete(key []byte) error {
	edb.codeDB.Delete(key)
	return nil
}

// Close implements Ethereum's ethdb.Database interface. It closes the
// underlying database.
func (edb *EthereumDB) Close() {
	edb.codeDB.Close()
}

// NewBatch implements Ethereum's ethdb.Database interface. It returns a new
// Batch object used for batch database operations.
func (edb *EthereumDB) NewBatch() ethdb.Batch {
	return edb
}

// ValueSize implements Ethereum's ethdb.Database interface. It performs a
// no-op.
func (edb *EthereumDB) ValueSize() int {
	return 0
}

// Write implements Ethereum's ethdb.Database interface. It performs a no-op.
func (edb *EthereumDB) Write() error {
	return nil
}

// Reset implements Ethereum's ethdb.Database interface. It performs a no-op.
func (edb *EthereumDB) Reset() {
}
