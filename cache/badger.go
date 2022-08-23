package cache

import (
	"github.com/dgraph-io/badger/v3"
	"time"
)

type BadgerCache struct {
	Conn   *badger.DB
	Prefix string
}

func (b BadgerCache) Has(key string) (bool, error) {
	_, err := b.Get(key)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (b BadgerCache) Get(key string) (interface{}, error) {
	var fromCache []byte

	err := b.Conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			fromCache = append([]byte{}, val...)
			return nil
		})

		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	decoded, err := decode(string(fromCache))
	if err != nil {
		return nil, err
	}

	return decoded[key], nil
}

func (b BadgerCache) Set(key string, value interface{}, expires ...int) error {

	entry := Entry{}
	entry[key] = value
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	if len(expires) > 0 {
		err = b.Conn.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(key), encoded).WithTTL(time.Second * time.Duration(expires[0]))
			err = txn.SetEntry(e)
			return err
		})
	} else {
		err = b.Conn.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(key), encoded)
			err = txn.SetEntry(e)
			return err
		})
	}

	return nil
}

func (b BadgerCache) Forget(key string) error {
	return b.Conn.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (b BadgerCache) EmptyByMatch(key string) error {
	return b.emptyByMatch(key)
}

func (b BadgerCache) Empty() error {
	return b.emptyByMatch("")
}

func (b *BadgerCache) emptyByMatch(pattern string) error {

	deleteKeys := func(keysForDelete [][]byte) error {
		if err := b.Conn.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	collectSize := 100000
	err := b.Conn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0

		for it.Seek([]byte(pattern)); it.ValidForPrefix([]byte(pattern)); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					return err
				}
			}
		}

		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				return err
			}
		}
		return nil
	})

	return err

}
