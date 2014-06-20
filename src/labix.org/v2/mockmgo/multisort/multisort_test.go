package multisort

import (
	"reflect"
	"testing"
)

import (
	"labix.org/v2/base/bson"
)

type Test struct {
	A int     `bson:"a"`
	B float32 `bson:"b"`
	C string  `bson:"c"`
}

type Test2 struct {
	A string `bson:"a"`
	B int    `bson:"b"`
	C uint   `bson:"c"`
}

type Test3 struct {
	A Test2 `bson:"a"`
	B int   `bson:"b"`
	C uint  `bson:"c"`
}

func TestMultiSort(t *testing.T) {
	data := []interface{}{Test{1, 10.0, "c"}, Test{1, 6.0, "b"}, Test{1, 6.0, "a"}, Test{2, 11.1, "a"}}
	result, err := MultiSort(data, bson.D{bson.DocElem{"a", 1}})
	if err != nil {
		t.Fatal("got err:", err)
	}

	if !reflect.DeepEqual(data, result) {
		t.Fatalf("expected:%#v, actual:%#v", data, result)
	}

	result, err = MultiSort(data, bson.D{bson.DocElem{"a", 1}, bson.DocElem{"b", 1}})
	if err != nil {
		t.Fatal("got err:", err)
	}

	expected := []interface{}{Test{1, 6.0, "b"}, Test{1, 6.0, "a"}, Test{1, 10.0, "c"}, Test{2, 11.1, "a"}}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected:%#v, actual:%#v", expected, result)
	}

	result, err = MultiSort(data, bson.D{bson.DocElem{"a", 1}, bson.DocElem{"b", 1}, bson.DocElem{"c", 1}})
	if err != nil {
		t.Fatal("got err:", err)
	}

	expected = []interface{}{Test{1, 6.0, "a"}, Test{1, 6.0, "b"}, Test{1, 10.0, "c"}, Test{2, 11.1, "a"}}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected:%#v, actual:%#v", expected, result)
	}

	result, err = MultiSort(data, bson.D{bson.DocElem{"a", -1}, bson.DocElem{"b", 1}, bson.DocElem{"c", 1}})
	if err != nil {
		t.Fatal("got err:", err)
	}

	expected = []interface{}{Test{2, 11.1, "a"}, Test{1, 6.0, "a"}, Test{1, 6.0, "b"}, Test{1, 10.0, "c"}}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected:%#v, actual:%#v", expected, result)
	}

	result, err = MultiSort(data, bson.D{bson.DocElem{"a", -1}, bson.DocElem{"b", -1}, bson.DocElem{"c", 1}})
	if err != nil {
		t.Fatal("got err:", err)
	}

	expected = []interface{}{Test{2, 11.1, "a"}, Test{1, 10.0, "c"}, Test{1, 6.0, "a"}, Test{1, 6.0, "b"}}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected:%#v, actual:%#v", expected, result)
	}

	result, err = MultiSort(data, bson.D{bson.DocElem{"a", -1}, bson.DocElem{"b", -1}, bson.DocElem{"c", -1}})
	if err != nil {
		t.Fatal("got err:", err)
	}

	expected = []interface{}{Test{2, 11.1, "a"}, Test{1, 10.0, "c"}, Test{1, 6.0, "b"}, Test{1, 6.0, "a"}}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected:%#v, actual:%#v", expected, result)
	}

	data = []interface{}{Test3{Test2{"haha", 3, 1}, 3, 4}, Test3{Test2{"haha", 1, 2}, 3, 4}}
	result, err = MultiSort(data, bson.D{bson.DocElem{"a", 1}})
	if err != nil {
		t.Fatal("got err", err)
	}
	expected = []interface{}{Test3{Test2{"haha", 1, 2}, 3, 4}, Test3{Test2{"haha", 3, 1}, 3, 4}}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected:%#v, actual:%#v", expected, result)
	}

	data = []interface{}{Test3{Test2{"haha", 3, 1}, 3, 4}, Test3{Test2{"haha", 1, 2}, 3, 4}}
	result, err = MultiSort(data, bson.D{bson.DocElem{"a", -1}})
	if err != nil {
		t.Fatal("got err", err)
	}
	expected = []interface{}{Test3{Test2{"haha", 3, 1}, 3, 4}, Test3{Test2{"haha", 1, 2}, 3, 4}}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected:%#v, actual:%#v", expected, result)
	}

	// negative test
	datadiff := []interface{}{Test{1, 10.0, "c"}, Test2{"haha", 1, 2}}
	_, err = MultiSort(datadiff, bson.D{bson.DocElem{"a", 1}})
	if err == nil {
		t.Fatal("should got err")
	}

}
