package mockmgo

import (
	"sync"
)

import (
	"labix.org/v2/base/bson"
	. "labix.org/v2/base/log"
)

var ErrNotFound = errors.New("not found")

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

func (c *Collection) Find(query interface{}) (q *Query) {
	q = &Query{coll: c}
	q.op.query = query
	return
}

type Query struct {
	coll *Collection
	op   QueryOp
}

type QueryOp struct {
	query    interface{}
	skip     int32
	limit    int32
	selector interface{}
}

func (q *Query) One(result interface{}) (err error) {
	q.coll.RLock()
	defer q.coll.RUnlock()

	// TO DO...
	if data == nil {
		return ErrNotFound
	}
	if result != nil {
		err = bson.Unmarshal(data, result)
		if err == nil {
			Debugf("Query %p document unmarshaled: %#v", q, result)
		} else {
			Debugf("Query %p document unmarshaling failed: %#v", q, err)
			return err
		}
	}
}

func (q *Query) Iter() imgo.Iter {
	q.coll.RLock()
	coll := q.coll
	op := q.op
	q.coll.RUnlock()

	iter := &Iter{
		coll: coll,
		op:   op,
	}
	return iter
}

type Iter struct {
	//m    sync.Mutex
	coll *Collection
	//gotReply sync.Cond
	//session        *Session
	//server         *mongoServer
	docData queue
	err     error
	op      QueryOp
	//op             getMoreOp
	//prefetch       float64
	//limit int32
	//docsToReceive  int
	//docsBeforeMore int
	//	timeout        time.Duration
	//timedout       bool
}
