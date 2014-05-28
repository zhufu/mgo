package imgo

type Collection interface {
	FindId(interface{}) Query
	Find(interface{}) Query
	//Pipe(pipeline interface{}) *Pipe
	Insert(...interface{}) error
	Update(selector interface{}, change interface{}) error
	UpdateId(id interface{}, change interface{}) error
	//UpdateAll(selector interface{}, change interface{}) (*mgo.ChangeInfo, error)
	//Upsert(selector interface{}, change interface{}) (*mgo.ChangeInfo, error)
	//UpsertId(id interface{}, change interface{}) (*mgo.ChangeInfo, error)
	Remove(selector interface{}) error
	RemoveId(id interface{}) error
	//RemoveAll(selector interface{}) (ChangeInfo, error)
	//DropCollection() error
	//Create(info *CollectionInfo) error
	//Count() (n int, err error)
	//With(s *Session) Collection
	//EnsureIndexKey(key ...string) error
	//EnsureIndex(index Index) error
	//DropIndex(key ...string) error
	//Indexes() (indexes []Index, err error)
}

type Query interface {
	Batch(n int) Query
	Prefetch(p float64) Query
	Skip(n int) Query
	Limit(n int) Query
	Select(selector interface{}) Query
	Sort(fields ...string) Query
	//Explain(result interface{}) error
	//Hint(indexKey ...string) Query
	//Snapshot() Query
	//LogReplay() Query
	One(result interface{}) (err error)
	Iter() Iter
	//Tail(timeout time.Duration) *Iter
	All(result interface{}) error
	//For(result interface{}, f func() error) error
	Count() (n int, err error)
	Distinct(key string, result interface{}) error
	//MapReduce(job *MapReduce, result interface{}) (info *MapReduceInfo, err error)
	//Apply(change Change, result interface{}) (info *ChangeInfo, err error)
}

type Iter interface {
	Err() error
	Close() error
	Timeout() bool
	Next(result interface{}) bool
	All(result interface{}) error
	For(result interface{}, f func() error) (err error)
}
