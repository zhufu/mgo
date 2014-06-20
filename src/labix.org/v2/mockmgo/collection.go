package mockmgo

import (
	"reflect"
	"sync"
	//"time"
	//"math"
)

import (
	"labix.org/v2/base/bson"
	. "labix.org/v2/base/log"
	. "labix.org/v2/base/queue"
	. "labix.org/v2/error"
	"labix.org/v2/imgo"
	"labix.org/v2/mockmgo/parse"
)

type Data map[string]interface{} // _id: json string

type Collection struct {
	name string
	data Data
	sync.RWMutex
}

func NewCollection(name string, data Data) (c *Collection) {
	c = &Collection{name: name, data: data}
	if data == nil {
		c.data = make(Data, 0)
	}
	return
}

func (c *Collection) Find(query interface{}) imgo.Query {
	q := &Query{coll: c}
	q.op.query = query
	return q
}

func (c *Collection) FindId(id interface{}) (q imgo.Query) {
	return c.Find(bson.D{{"_id", id}})
}

func (c *Collection) findOne(query interface{}) (result interface{}, err error) {
	Debugf("query:%#v", query)
	queryM, ok := query.(bson.M)
	if !ok {
		return nil, UnknownQuery
	}

	Debugf("query:%#v", queryM)
	if objId, ok := queryM["_id"]; ok {
		Debug("good")
		id, ok := objId.(bson.ObjectId)
		if !ok {
			return nil, ErrNotFound
		}
		Debug("11111")
		result, ok := c.data[id.Hex()]
		if !ok {
			return nil, ErrNotFound
		}
		Debug("22222")
		return result, nil
	}

	// no index, iter all data
	for _, data := range c.data {
		if ok, _ := parse.Match(data, query); ok {
			return data, nil
		}
	}
	return nil, ErrNotFound
}

func (c *Collection) count(query interface{}) (n int, err error) {
	queryM, ok := query.(bson.M)
	if !ok {
		return 0, UnknownQuery
	}

	if objId, ok := queryM["_id"]; ok {
		id, ok := objId.(bson.ObjectId)
		if !ok {
			return
		}
		if _, ok := c.data[id.Hex()]; !ok {
			return
		}
		n = 1
		return
	}

	n = 0
	for _, data := range c.data {
		if ok, _ := parse.Match(data, query); ok {
			n++
		}
	}
	return
}

type Query struct {
	coll *Collection
	op   QueryOp
}

type QueryOp struct {
	query   interface{}
	OrderBy bson.D
	skip    int32
	limit   int32
}

func (q *Query) One(result interface{}) (err error) {
	q.coll.RLock()
	defer q.coll.RUnlock()

	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr {
		return ErrType
	}

	data, err := q.coll.findOne(q.op.query)
	if err != nil {
		return
	}

	resultv.Elem().Set(reflect.ValueOf(data))

	return
}

func (q *Query) All(result interface{}) (err error) {
	return q.Iter().All(result)
}

func (q *Query) Iter() imgo.Iter {
	iter := &Iter{
		coll: q.coll,
		op:   q.op,
	}
	return iter
}

func (q *Query) Count() (n int, err error) {
	q.coll.RLock()
	defer q.coll.RUnlock()

	return q.coll.count(q.op.query)
}

func (q *Query) Skip(n int) imgo.Query {
	q.op.skip = int32(n)
	return q
}

func (q *Query) Limit(n int) imgo.Query {
	q.op.limit = int32(n)
	return q
}

func (q *Query) Sort(fields ...string) imgo.Query {
	var order bson.D
	for _, field := range fields {
		n := 1
		if field != "" {
			switch field[0] {
			case '+':
				field = field[1:]
			case '-':
				n = -1
				field = field[1:]
			}
		}
		if field == "" {
			panic("Sort: empty field name")
		}
		order = append(order, bson.DocElem{field, n})
	}
	q.op.OrderBy = order
	return q
}

type Iter struct {
	coll    *Collection
	err     error
	op      QueryOp
	docData Queue
	got     bool
}

func (iter *Iter) Err() error {
	return iter.err
}

func (iter *Iter) Close() error {
	return iter.err
}

func (iter *Iter) Timeout() bool {
	return true
}

func (iter *Iter) getAll() {
	iter.coll.RLock()
	defer iter.coll.RUnlock()

	for _, data := range iter.coll.data {
		if ok, _ := parse.Match(data, iter.op.query); !ok {
			continue
		}
		iter.docData.Push(data)
	}

	if len(iter.op.OrderBy) > 0 {
		_ = iter.docData.Sort(iter.op.OrderBy)
	}

	iter.docData.Skip(int(iter.op.skip))
	iter.docData.Limit(int(iter.op.limit))

	iter.got = true
	return
}

func (iter *Iter) Next(result interface{}) bool {
	if iter.err == nil && !iter.got {
		iter.getAll()
	}

	resultv := reflect.ValueOf(result)
	if docData := iter.docData.Pop(); docData != nil {
		resultv.Elem().Set(reflect.ValueOf(docData))
		return true
	}
	return false
}

func (iter *Iter) All(result interface{}) error {
	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr || resultv.Elem().Kind() != reflect.Slice {
		return ErrTypeAll
	}
	slicev := resultv.Elem()
	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()
	i := 0
	for {
		if slicev.Len() == i {
			elemp := reflect.New(elemt)
			if !iter.Next(elemp.Interface()) {
				break
			}
			slicev = reflect.Append(slicev, elemp.Elem())
			slicev = slicev.Slice(0, slicev.Cap())
		} else {
			if !iter.Next(slicev.Index(i).Addr().Interface()) {
				break
			}
		}
		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))
	return iter.Close()
}
