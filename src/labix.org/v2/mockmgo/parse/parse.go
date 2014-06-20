package parse

import (
	"reflect"
	"strings"
)

import (
	"labix.org/v2/base/bson"
	. "labix.org/v2/error"
)

func Match(data, query interface{}) (match bool, err error) {
	if query == nil {
		return true, nil
	}

	queryFields, err := parseQuery(query)
	if err != nil {
		return
	}

	match = compareWithQuerys(reflect.ValueOf(data), queryFields)
	return
}

func parseQuery(query interface{}) (queryFields QueryFields, err error) {
	queryM, ok := query.(bson.M)
	if !ok {
		err = UnknownQuery
		return
	}

	queryFields = parseBsonM("", queryM, false)
	return
}

func parseBsonM(field string, queryM bson.M, elematch bool) (queryFields QueryFields) {
	queryFields.Relation = AND
	queryFields.ElemMatch = elematch
	queryFields.Key = field
	queryLen := len(queryM)
	for k, v := range queryM {
		if k == "$or" {
			k = ""
			queryFields.Relation = OR
		}

		if k == "$elemMatch" {
			k = ""
			queryFields.ElemMatch = true
		}

		op := EQ
		if op_, ok := OperateM[k]; ok {
			op = op_
			k = ""
		}

		if k == "" {
			k = field
			queryFields.Key = ""
		}

		if vM, ok := v.(bson.M); ok {
			queryFields_ := parseBsonM(k, vM, queryFields.ElemMatch)
			if queryLen == 1 {
				queryFields = queryFields_
			} else {
				if len(queryFields_.Fields) == 1 {
					queryFields.Fields = append(queryFields.Fields, queryFields_.Fields[0])
				} else {
					queryFields.Fields = append(queryFields.Fields, queryFields_)
				}
			}
		} else if vsM, ok := v.([]bson.M); ok {
			for _, vM := range vsM {
				queryFields_ := parseBsonM(k, vM, queryFields.ElemMatch)
				if len(queryFields_.Fields) == 1 {
					queryFields.Fields = append(queryFields.Fields, queryFields_.Fields[0])
				} else {
					queryFields.Fields = append(queryFields.Fields, queryFields_)
				}
			}
		} else {
			queryFields.Fields = append(queryFields.Fields, QueryField{k, v, op})
		}
	}
	return
}

func GetStructValueByFlag(k string, data reflect.Value) (result reflect.Value, ok bool) {

	key, left := k, ""
	index := strings.Index(k, ".")
	if index != -1 {
		key = k[:index]
		left = k[index+1:]
	}

	type_ := data.Type()
	for i := 0; i < data.NumField(); i++ {
		field := type_.Field(i)
		tag := field.Tag.Get("bson")
		if tag == "" {
			tag = field.Tag.Get("json")
		}
		v := data.Field(i)
		if key == tag {
			if left != "" {
				return GetStructValueByFlag(left, v)
			}
			return v, true
		}
	}

	return reflect.ValueOf(nil), false
}

type QueryField struct {
	Field string
	Value interface{}
	Op    OP
}

type QueryFields struct {
	Fields    []interface{}
	Relation  RELATION
	ElemMatch bool
	Key       string
}

type RELATION int

const (
	AND RELATION = 1
	OR  RELATION = 2
)

var OperateM = map[string]OP{
	"$ne":     NE,
	"$gt":     GT,
	"$gte":    GTE,
	"$lt":     LT,
	"$lte":    LTE,
	"$all":    EQ,
	"$in":     IN,
	"$nin":    NIN,
	"$exists": EXISTS,
}

type OP int

var (
	EQ     OP = 0x00000001
	NE     OP = 0x00000002
	GT     OP = 0x00000004
	GTE    OP = 0x00000008
	LT     OP = 0x00000010
	LTE    OP = 0x00000020
	IN     OP = 0x00000040
	NIN    OP = 0x00000080
	EXISTS OP = 0x00000100
)

func compareElemMatch(data reflect.Value, queryFields QueryFields) bool {
	if data.Kind() != reflect.Array && data.Kind() != reflect.Slice {
		return false
	}

	for i := 0; i < data.Len(); i++ {
		if compareWithQuerys(data.Index(i), queryFields) {
			return true
		}
	}

	return false
}

func compareWithQuerys(data reflect.Value, queryFields QueryFields) bool {
	if queryFields.Key != "" {
		data, _ = GetStructValueByFlag(queryFields.Key, data)
		queryFields.Key = ""
	}

	if queryFields.ElemMatch {
		queryFields.ElemMatch = false
		return compareElemMatch(data, queryFields)
	}

	relation := queryFields.Relation
	match := true

	for _, field := range queryFields.Fields {
		if queryField, ok := field.(QueryField); ok {
			actualData, _ := GetStructValueByFlag(queryField.Field, data)
			match = compare(actualData, reflect.ValueOf(queryField.Value), queryField.Op)
		} else if queryFields_, ok := field.(QueryFields); ok {
			match = compareWithQuerys(data, queryFields_)
		} else {
			match = false
		}
		if (match && relation == OR) || (!match && relation == AND) {
			return match
		}
	}

	return match
}

func compare(data1, data2 reflect.Value, op OP) bool {
	switch op {
	case EXISTS:
		if data2.Kind() != reflect.Bool {
			return false
		}
		return data2.Bool() != (data1 == reflect.ValueOf(nil))
	case IN:
		if data2.Kind() != reflect.Array && data2.Kind() != reflect.Slice {
			return false
		}
		for i := 0; i < data2.Len(); i++ {
			if compare(data1, data2.Index(i), EQ) {
				return true
			}
		}
		return false
	case NIN:
		if data2.Kind() != reflect.Array && data2.Kind() != reflect.Slice {
			return false
		}
		if data1 == reflect.ValueOf(nil) {
			return true
		}
		for i := 0; i < data2.Len(); i++ {
			if compare(data1, data2.Index(i), EQ) {
				return false
			}
		}
		return true
	default:

		if data1.Kind() != data2.Kind() {
			return op == NE
		}

		switch data1.Kind() {
		case reflect.Ptr, reflect.Interface:
			return compare(data1.Elem(), data2.Elem(), op)
		case reflect.Struct:
			if data1.NumField() != data2.NumField() {
				return op == NE
			}
			result := true
			for i := 0; i < data1.NumField(); i++ {
				result = compare(data1.Field(i), data2.Field(i), op)
				if (result && op == NE) || (!result && op != NE) {
					return result
				}
			}
			return result
		case reflect.Map:
			result := true
			keys := data1.MapKeys()
			for _, key := range keys {
				result = compare(data1.MapIndex(key), data2.MapIndex(key), op)
				if (result && op == NE) || (!result && op != NE) {
					return result
				}
			}
			return result
		case reflect.Array, reflect.Slice:
			if data1.Len() != data2.Len() {
				return op == NE
			}
			result := true
			for i := 0; i < data1.Len(); i++ {
				result = compare(data1.Index(i), data2.Index(i), op)
				if (result && op == NE) || (!result && op != NE) {
					return result
				}
			}
			return result
		case reflect.Invalid:
			return op == EQ
		default:
			return compareSimple(data1, data2, op)
		}
	}

	return false
}

func compareSimple(data1, data2 reflect.Value, op OP) bool {
	switch data1.Kind() {
	case reflect.String:
		d1 := data1.String()
		d2 := data2.String()

		switch op {
		case GT:
			return d1 > d2
		case GTE:
			return d1 >= d2
		case LT:
			return d1 < d2
		case LTE:
			return d1 <= d2
		case EQ:
			return d1 == d2
		case NE:
			return d1 != d2
		}
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		d1 := data1.Int()
		d2 := data2.Int()

		switch op {
		case GT:
			return d1 > d2
		case GTE:
			return d1 >= d2
		case LT:
			return d1 < d2
		case LTE:
			return d1 <= d2
		case EQ:
			return d1 == d2
		case NE:
			return d1 != d2
		}
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uintptr, reflect.Uint16, reflect.Uint8:
		d1 := data1.Uint()
		d2 := data2.Uint()
		switch op {
		case GT:
			return d1 > d2
		case GTE:
			return d1 >= d2
		case LT:
			return d1 < d2
		case LTE:
			return d1 <= d2
		case EQ:
			return d1 == d2
		case NE:
			return d1 != d2
		}
	case reflect.Bool:
		d1 := data1.Bool()
		d2 := data2.Bool()
		switch op {
		case GT, GTE, LT, LTE:
			return false
		case EQ:
			return d1 == d2
		case NE:
			return d1 != d2
		}
	case reflect.Float64, reflect.Float32:
		d1 := data1.Float()
		d2 := data2.Float()
		switch op {
		case GT:
			return d1 > d2
		case GTE:
			return d1 >= d2
		case LT:
			return d1 < d2
		case LTE:
			return d1 <= d2
		case EQ:
			return d1 == d2
		case NE:
			return d1 != d2
		}
	}
	return false
}
