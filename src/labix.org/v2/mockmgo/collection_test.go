package mockmgo

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"labix.org/v2/base/bson"
	. "labix.org/v2/base/log"
)

type cLogger struct{}

func (c *cLogger) Output(calldepth int, s string) error {
	t := time.Now()
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}

	mondule := "UNKNOWN"
	name := "???"
	pos := strings.LastIndex(file, "/")
	if pos != -1 {
		name = file[pos+1:]
		pos1 := strings.LastIndex(file[:pos], "/src/")
		if pos1 != -1 {
			mondule = file[pos1+5 : pos]
		}
	}

	fmt.Printf("%s [%s] %s:%d: %s \n", t.Format("2006/01/02 15:04:05"), mondule, name, line, s)
	return nil
}

func testFindOne(t *testing.T) {
	type Inner struct {
		F int     `bson:"f"`
		G float64 `bson:"g"`
	}

	type Test struct {
		Id bson.ObjectId `bson:"_id"`
		A  int           `bson:"a"`
		B  string        `bson:"b"`
		D  []int         `bson:"d"`
		E  []Inner       `bson:"e"`
		H  float32       `bson:"h"`
	}

	data := Test{
		Id: bson.NewObjectId(),
		A:  10,
		B:  "good",
		D:  []int{1, 2, 3},
		E:  []Inner{Inner{1, float64(2.0)}, Inner{5, float64(1.0)}},
		H:  1.01,
	}

	dataM := Data{
		data.Id.Hex(): data,
	}

	c := NewCollection("test1", dataM)
	defer func() {
		c = nil
	}()

	var result Test
	err := c.Find(bson.M{"_id": data.Id}).One(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual(data, result) {
		t.Fatal("expected:", data, ", actual:", result)
	}

	err = c.Find(bson.M{"h": float32(1.01)}).One(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual(data, result) {
		t.Fatal("expected:", data, ", actual:", result)
	}
}

func testFindAll(t *testing.T) {
	type Test struct {
		Id bson.ObjectId `bson:"_id"`
		A  int           `bson:"a"`
	}

	data := Test{
		Id: bson.NewObjectId(),
		A:  10,
	}

	data2 := Test{
		Id: bson.NewObjectId(),
		A:  20,
	}

	dataM := Data{
		data.Id.Hex():  data,
		data2.Id.Hex(): data2,
	}

	c := NewCollection("test2", dataM)
	defer func() {
		c = nil
	}()

	var result []Test
	err := c.Find(bson.M{"a": 10}).All(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual([]Test{data}, result) {
		t.Fatal("expected:", []Test{data}, ", actual:", result)
	}

	err = c.Find(bson.M{"a": bson.M{"$gte": 10}}).All(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual([]Test{data, data2}, result) {
		t.Fatal("expected:", []Test{data, data2}, ", actual:", result)
	}

	// test iterate
	result = []Test{}
	iter := c.Find(nil).Iter()
	var elem Test
	for iter.Next(&elem) {
		result = append(result, elem)
	}
	if !reflect.DeepEqual([]Test{data, data2}, result) {
		t.Fatal("expected:", []Test{data, data2}, ", actual:", result)
	}

	// negative test
	err = c.Find(bson.M{"a": bson.M{"$gte": 10}}).All(result)
	if err == nil {
		t.Fatal("should failed but succeed!")
	}

	// count
	n, err := c.Find(bson.M{"a": bson.M{"$gte": 10}}).Count()
	if err != nil || n != 2 {
		t.Fatal("count failed:", err)
	}

	n, err = c.Find(bson.M{"a": bson.M{"$gte": 15}}).Count()
	if err != nil || n != 1 {
		t.Fatal("count failed:", err)
	}

	n, err = c.Find(bson.M{"a": bson.M{"$gt": 20}}).Count()
	if err != nil || n != 0 {
		t.Fatal("count failed:", err)
	}

	n, err = c.Find(bson.M{"a": bson.M{"$lt": 10}}).Count()
	if err != nil || n != 0 {
		t.Fatal("count failed:", err)
	}

	n, err = c.Find(bson.M{"_id": data.Id}).Count()
	if err != nil || n != 1 {
		t.Fatal("find failed:", err)
	}

	n, err = c.Find(bson.M{"_id": bson.NewObjectId()}).Count()
	if err != nil || n != 0 {
		t.Fatal("find failed:", err)
	}

	// test offset, limit
	err = c.Find(bson.M{"a": bson.M{"$gte": 10}}).Skip(1).All(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual([]Test{data2}, result) {
		t.Fatal("expected:", []Test{data2}, ", actual:", result)
	}

	err = c.Find(bson.M{"a": bson.M{"$gte": 10}}).Skip(2).All(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual([]Test{}, result) {
		t.Fatal("expected:", []Test{}, ", actual:", result)
	}

	err = c.Find(bson.M{"a": bson.M{"$gte": 10}}).Skip(-1).All(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual([]Test{data, data2}, result) {
		t.Fatal("expected:", []Test{data, data2}, ", actual:", result)
	}

	err = c.Find(bson.M{"a": bson.M{"$gte": 10}}).Skip(0).Limit(1).All(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual([]Test{data}, result) {
		t.Fatal("expected:", []Test{data}, ", actual:", result)
	}

	err = c.Find(bson.M{"a": bson.M{"$gte": 10}}).Skip(1).Limit(1).All(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual([]Test{data2}, result) {
		t.Fatal("expected:", []Test{data2}, ", actual:", result)
	}

	err = c.Find(bson.M{"a": bson.M{"$gte": 10}}).Limit(0).All(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual([]Test{data, data2}, result) {
		t.Fatal("expected:", []Test{data, data2}, ", actual:", result)
	}

	// test sort
	err = c.Find(bson.M{"a": bson.M{"$gte": 10}}).Sort("a").All(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual([]Test{data, data2}, result) {
		t.Fatal("expected:", []Test{data, data2}, ", actual:", result)
	}

	err = c.Find(bson.M{"a": bson.M{"$gte": 10}}).Sort("-a").All(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual([]Test{data2, data}, result) {
		t.Fatal("expected:", []Test{data2, data}, ", actual:", result)
	}

	err = c.Find(nil).Sort("-a").Skip(1).All(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual([]Test{data}, result) {
		t.Fatal("expected:", []Test{data}, ", actual:", result)
	}

	err = c.Find(nil).Sort("-a").Skip(0).Limit(1).All(&result)
	if err != nil {
		t.Fatal("find failed:", err)
	}

	if !reflect.DeepEqual([]Test{data2}, result) {
		t.Fatal("expected:", []Test{data2}, ", actual:", result)
	}
}

func TestFind(t *testing.T) {
	SetDebug(true)
	SetLogger(new(cLogger))

	testFindOne(t)
	testFindAll(t)
}
