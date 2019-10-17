package grocksdb

// #include "rocksdb/c.h"
import "C"

// WriteBatchWI is a batching with index of Puts, Merges and Deletes to implement read-your-own-write.
// See also: https://rocksdb.org/blog/2015/02/27/write-batch-with-index.html
type WriteBatchWI struct {
	c *C.rocksdb_writebatch_wi_t
}

// NewWriteBatchWI create a WriteBatchWI object.
// - reserved_bytes: reserved bytes in underlying WriteBatch
// - overwrite_key: if true, overwrite the key in the index when inserting
//                the same key as previously, so iterator will never
//                show two entries with the same key.
func NewWriteBatchWI(reservedBytes uint, overwriteKeys bool) *WriteBatchWI {
	return NewNativeWriteBatchWI(C.rocksdb_writebatch_wi_create(C.size_t(reservedBytes), boolToChar(overwriteKeys)))
}

// NewNativeWriteBatchWI create a WriteBatchWI object.
func NewNativeWriteBatchWI(c *C.rocksdb_writebatch_wi_t) *WriteBatchWI {
	return &WriteBatchWI{c}
}

// Put queues a key-value pair.
func (wb *WriteBatchWI) Put(key, value []byte) {
	cKey := byteToChar(key)
	cValue := byteToChar(value)
	C.rocksdb_writebatch_wi_put(wb.c, cKey, C.size_t(len(key)), cValue, C.size_t(len(value)))
}

// PutCF queues a key-value pair in a column family.
func (wb *WriteBatchWI) PutCF(cf *ColumnFamilyHandle, key, value []byte) {
	cKey := byteToChar(key)
	cValue := byteToChar(value)
	C.rocksdb_writebatch_wi_put_cf(wb.c, cf.c, cKey, C.size_t(len(key)), cValue, C.size_t(len(value)))
}

// PutLogData appends a blob of arbitrary size to the records in this batch.
func (wb *WriteBatchWI) PutLogData(blob []byte) {
	cBlob := byteToChar(blob)
	C.rocksdb_writebatch_wi_put_log_data(wb.c, cBlob, C.size_t(len(blob)))
}

// Merge queues a merge of "value" with the existing value of "key".
func (wb *WriteBatchWI) Merge(key, value []byte) {
	cKey := byteToChar(key)
	cValue := byteToChar(value)
	C.rocksdb_writebatch_wi_merge(wb.c, cKey, C.size_t(len(key)), cValue, C.size_t(len(value)))
}

// MergeCF queues a merge of "value" with the existing value of "key" in a
// column family.
func (wb *WriteBatchWI) MergeCF(cf *ColumnFamilyHandle, key, value []byte) {
	cKey := byteToChar(key)
	cValue := byteToChar(value)
	C.rocksdb_writebatch_wi_merge_cf(wb.c, cf.c, cKey, C.size_t(len(key)), cValue, C.size_t(len(value)))
}

// Delete queues a deletion of the data at key.
func (wb *WriteBatchWI) Delete(key []byte) {
	cKey := byteToChar(key)
	C.rocksdb_writebatch_wi_delete(wb.c, cKey, C.size_t(len(key)))
}

// DeleteCF queues a deletion of the data at key in a column family.
func (wb *WriteBatchWI) DeleteCF(cf *ColumnFamilyHandle, key []byte) {
	cKey := byteToChar(key)
	C.rocksdb_writebatch_wi_delete_cf(wb.c, cf.c, cKey, C.size_t(len(key)))
}

// DeleteRange deletes keys that are between [startKey, endKey)
func (wb *WriteBatchWI) DeleteRange(startKey []byte, endKey []byte) {
	cStartKey := byteToChar(startKey)
	cEndKey := byteToChar(endKey)
	C.rocksdb_writebatch_wi_delete_range(wb.c, cStartKey, C.size_t(len(startKey)), cEndKey, C.size_t(len(endKey)))
}

// DeleteRangeCF deletes keys that are between [startKey, endKey) and
// belong to a given column family
func (wb *WriteBatchWI) DeleteRangeCF(cf *ColumnFamilyHandle, startKey []byte, endKey []byte) {
	cStartKey := byteToChar(startKey)
	cEndKey := byteToChar(endKey)
	C.rocksdb_writebatch_wi_delete_range_cf(wb.c, cf.c, cStartKey, C.size_t(len(startKey)), cEndKey, C.size_t(len(endKey)))
}

// Data returns the serialized version of this batch.
func (wb *WriteBatchWI) Data() []byte {
	var cSize C.size_t
	cValue := C.rocksdb_writebatch_wi_data(wb.c, &cSize)
	return charToByte(cValue, cSize)
}

// Count returns the number of updates in the batch.
func (wb *WriteBatchWI) Count() int {
	return int(C.rocksdb_writebatch_wi_count(wb.c))
}

// NewIterator returns a iterator to iterate over the records in the batch.
func (wb *WriteBatchWI) NewIterator() *WriteBatchIterator {
	data := wb.Data()
	if len(data) < 8+4 {
		return &WriteBatchIterator{}
	}
	return &WriteBatchIterator{data: data[12:]}
}

// SetSavePoint records the state of the batch for future calls to RollbackToSavePoint().
// May be called multiple times to set multiple save points.
func (wb *WriteBatchWI) SetSavePoint() {
	C.rocksdb_writebatch_wi_set_save_point(wb.c)
}

// RollbackToSavePoint removes all entries in this batch (Put, Merge, Delete, PutLogData) since the
// most recent call to SetSavePoint() and removes the most recent save point.
func (wb *WriteBatchWI) RollbackToSavePoint() (err error) {
	var cErr *C.char
	C.rocksdb_writebatch_wi_rollback_to_save_point(wb.c, &cErr)
	err = fromCError(cErr)
	return
}

// Get returns the data associated with the key from batch.
func (wb *WriteBatchWI) Get(opts *Options, key []byte) (slice *Slice, err error) {
	var (
		cErr    *C.char
		cValLen C.size_t
		cKey    = byteToChar(key)
	)

	cValue := C.rocksdb_writebatch_wi_get_from_batch(wb.c, opts.c, cKey, C.size_t(len(key)), &cValLen, &cErr)
	if err = fromCError(cErr); err == nil {
		slice = NewSlice(cValue, cValLen)
	}

	return
}

// GetWithCF returns the data associated with the key from batch.
// Key belongs to specific column family.
func (wb *WriteBatchWI) GetWithCF(opts *Options, cf *ColumnFamilyHandle, key []byte) (slice *Slice, err error) {
	var (
		cErr    *C.char
		cValLen C.size_t
		cKey    = byteToChar(key)
	)

	cValue := C.rocksdb_writebatch_wi_get_from_batch_cf(wb.c, opts.c, cf.c, cKey, C.size_t(len(key)), &cValLen, &cErr)
	if err = fromCError(cErr); err == nil {
		slice = NewSlice(cValue, cValLen)
	}

	return
}

// GetFromDB returns the data associated with the key from the database and write batch.
func (wb *WriteBatchWI) GetFromDB(db *DB, opts *ReadOptions, key []byte) (slice *Slice, err error) {
	var (
		cErr    *C.char
		cValLen C.size_t
		cKey    = byteToChar(key)
	)

	cValue := C.rocksdb_writebatch_wi_get_from_batch_and_db(wb.c, db.c, opts.c, cKey, C.size_t(len(key)), &cValLen, &cErr)
	if err = fromCError(cErr); err == nil {
		slice = NewSlice(cValue, cValLen)
	}

	return
}

// GetFromDBWithCF returns the data associated with the key from the database and write batch.
// Key belongs to specific column family.
func (wb *WriteBatchWI) GetFromDBWithCF(db *DB, opts *ReadOptions, cf *ColumnFamilyHandle, key []byte) (slice *Slice, err error) {
	var (
		cErr    *C.char
		cValLen C.size_t
		cKey    = byteToChar(key)
	)

	cValue := C.rocksdb_writebatch_wi_get_from_batch_and_db_cf(wb.c, db.c, opts.c, cf.c, cKey, C.size_t(len(key)), &cValLen, &cErr)
	if err = fromCError(cErr); err == nil {
		slice = NewSlice(cValue, cValLen)
	}

	return
}

// Clear removes all the enqueued Put and Deletes.
func (wb *WriteBatchWI) Clear() {
	C.rocksdb_writebatch_wi_clear(wb.c)
}

// Destroy deallocates the WriteBatch object.
func (wb *WriteBatchWI) Destroy() {
	C.rocksdb_writebatch_wi_destroy(wb.c)
	wb.c = nil
}
