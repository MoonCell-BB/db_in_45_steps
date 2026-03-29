package storage

import "bytes"

type KV struct {
	log Log
	mem map[string][]byte
}

func (kv *KV) Open() error {
	if err := kv.log.Open(); err != nil {
		return err
	}

	kv.mem = make(map[string][]byte)
	for {
		ent := Entry{}
		if eof, err := kv.log.Read(&ent); err != nil {
			return err
		} else if eof {
			break
		}

		if ent.deleted {
			delete(kv.mem, string(ent.key))
		} else {
			kv.mem[string(ent.key)] = ent.val
		}
	}

	return nil
}

func (kv *KV) Close() error {
	return kv.log.Close()
}

func (kv *KV) Get(key []byte) (val []byte, ok bool, err error) {
	val, ok = kv.mem[string(key)]
	return
}

func (kv *KV) Set(key []byte, val []byte) (updated bool, err error) {
	prev, exist := kv.mem[string(key)]
	updated = !exist || !bytes.Equal(prev, val)

	if updated {
		ent := &Entry{key: key, val: val}
		if err = kv.log.Write(ent); err != nil {
			return false, err
		}

		kv.mem[string(key)] = val
	}

	return
}

func (kv *KV) Del(key []byte) (deleted bool, err error) {
	_, deleted = kv.mem[string(key)]

	if deleted {
		ent := &Entry{key: key, deleted: true}
		if err = kv.log.Write(ent); err != nil {
			return false, err
		}

		delete(kv.mem, string(key))
	}

	return
}
