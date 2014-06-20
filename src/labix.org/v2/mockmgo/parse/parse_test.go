package parse

import (
	"reflect"
	"testing"
)

import (
	"labix.org/v2/base/bson"
)

func TestParseBsonM(t *testing.T) {
	var queryM bson.M
	var result QueryFields

	queryM = bson.M{"A": "B"}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Relation: AND, Fields: []interface{}{QueryField{"A", "B", EQ}}}, result)

	queryM = bson.M{"A": "B", "C": 1, "D": float32(1.5)}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Relation: AND, Fields: []interface{}{QueryField{"A", "B", EQ}, QueryField{"C", 1, EQ}, QueryField{"D", float32(1.5), EQ}}}, result)

	queryM = bson.M{"A": bson.M{"$gt": 3}}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Relation: AND, Fields: []interface{}{QueryField{"A", 3, GT}}}, result)

	queryM = bson.M{"A.B": bson.M{"$ne": 5}}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Relation: AND, Fields: []interface{}{QueryField{"A.B", 5, NE}}}, result)

	queryM = bson.M{"$or": []bson.M{bson.M{"A": bson.M{"$lt": 5}}, bson.M{"B": 10}}}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Relation: OR, Fields: []interface{}{QueryField{"A", 5, LT}, QueryField{"B", 10, EQ}}}, result)

	queryM = bson.M{"$or": []bson.M{bson.M{"A": bson.M{"$lt": 5}}, bson.M{"B": bson.M{"$gte": 10}}}}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Relation: OR, Fields: []interface{}{QueryField{"A", 5, LT}, QueryField{"B", 10, GTE}}}, result)

	queryM = bson.M{"$or": []bson.M{bson.M{"A": bson.M{"$lt": 5}, "B": bson.M{"$lte": 20}}, bson.M{"B": bson.M{"$gte": 10}}}}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Relation: OR, Fields: []interface{}{QueryFields{Relation: AND, Fields: []interface{}{QueryField{"A", 5, LT}, QueryField{"B", 20, LTE}}}, QueryField{"B", 10, GTE}}}, result)

	queryM = bson.M{"A": bson.M{"$all": []int64{1, 2, 3}}}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Relation: AND, Fields: []interface{}{QueryField{"A", []int64{1, 2, 3}, EQ}}}, result)

	queryM = bson.M{"A": bson.M{"$in": []int64{1, 2, 3}}}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Relation: AND, Fields: []interface{}{QueryField{"A", []int64{1, 2, 3}, IN}}}, result)

	queryM = bson.M{"A": bson.M{"$nin": []int64{1, 2, 3}}}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Relation: AND, Fields: []interface{}{QueryField{"A", []int64{1, 2, 3}, NIN}}}, result)

	queryM = bson.M{"A": bson.M{"$exists": false}}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Relation: AND, Fields: []interface{}{QueryField{"A", false, EXISTS}}}, result)

	queryM = bson.M{"A": bson.M{"$elemMatch": bson.M{"type": "base"}}}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Key: "A", Relation: AND, ElemMatch: true, Fields: []interface{}{QueryField{"type", "base", EQ}}}, result)

	queryM = bson.M{"A": bson.M{"$elemMatch": bson.M{"effect_time": bson.M{"$lt": 1000}, "type": "base"}}}
	result = parseBsonM("", queryM, false)
	assertQueryFields(t, QueryFields{Key: "A", Relation: AND, ElemMatch: true, Fields: []interface{}{QueryField{"effect_time", 1000, LT}, QueryField{"type", "base", EQ}}}, result)
}

func TestMatch(t *testing.T) {
	var queryM bson.M

	type Inner struct {
		F int     `bson:"f"`
		G float64 `bson:"g"`
	}

	data := struct {
		A int     `bson:"a"`
		B string  `bson:"b"`
		D []int   `bson:"d"`
		E []Inner `bson:"e"`
		H float32 `bson:"h"`
	}{
		A: 10,
		B: "good",
		D: []int{1, 2, 3},
		E: []Inner{Inner{1, float64(2.0)}, Inner{5, float64(1.0)}},
		H: 1.01,
	}

	queryM = bson.M{"a": 10}
	assertMatch(t, data, queryM)

	queryM = bson.M{"h": float32(1.01)}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": 10, "b": "good"}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$gt": 9}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$gte": 10}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$lt": 11}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$lte": 10}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$ne": 11}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$ne": nil}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$gt": 9, "$lt": 11}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$gt": 9}, "b": "good"}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$gt": 9}, "b": bson.M{"$ne": "bad"}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"c": nil}
	assertMatch(t, data, queryM)

	queryM = bson.M{"c": bson.M{"$exists": false}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$exists": true}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$gt": 9}, "c": bson.M{"$exists": false}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"$or": []bson.M{bson.M{"a": bson.M{"$lt": 10}}, bson.M{"b": "good"}}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"$or": []bson.M{bson.M{"a": bson.M{"$lt": 10}}, bson.M{"b": bson.M{"$ne": "good"}}}}
	assertNotMatch(t, data, queryM)

	queryM = bson.M{"$or": []bson.M{bson.M{"a": bson.M{"$gt": 9}, "b": bson.M{"$ne": "good"}}, bson.M{"b": "good"}}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"d": bson.M{"$all": []int{1, 2, 3}}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"d": bson.M{"$all": []int{1, 2}}}
	assertNotMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$in": []int{9, 10, 11}}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$in": []int{9, 12, 11}}}
	assertNotMatch(t, data, queryM)

	queryM = bson.M{"a": bson.M{"$nin": []int{9, 12, 11}}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"c": bson.M{"$nin": []int{9, 12, 11}}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"e": bson.M{"$elemMatch": bson.M{"f": bson.M{"$lte": 1}, "g": float64(2.0)}}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"e": bson.M{"$elemMatch": bson.M{"f": bson.M{"$gt": 4}, "g": bson.M{"$lte": float64(1.0)}}}}
	assertMatch(t, data, queryM)

	queryM = bson.M{"e": bson.M{"$elemMatch": bson.M{"f": bson.M{"$lte": 1}, "g": bson.M{"$lte": float64(1.0)}}}}
	assertNotMatch(t, data, queryM)
}

func assertQueryFields(t *testing.T, expected, result QueryFields) {
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected:%#v, actual:%#v", expected, result)
	}
}

func assertMatch(t *testing.T, data interface{}, queryM bson.M) {
	match, _ := Match(data, queryM)
	if !match {
		t.Fatalf("should be match, data:%#v, query:%#v", data, queryM)
	}
}

func assertNotMatch(t *testing.T, data interface{}, queryM bson.M) {
	match, _ := Match(data, queryM)
	if match {
		t.Fatalf("should not match, data:%#v, query:%#v", data, queryM)
	}
}
