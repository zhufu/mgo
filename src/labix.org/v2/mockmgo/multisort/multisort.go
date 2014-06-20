package multisort

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
)

import (
	"labix.org/v2/base/bson"
)

import (
	"labix.org/v2/mockmgo/parse"
)

type OP int

const (
	EQ OP = 1
	LT OP = 2
	GT OP = 3
)

func MultiSort(dataArray []interface{}, sortBy bson.D) (result []interface{}, err error) {
	dataArray2D := [][]interface{}{dataArray}

	for _, sortByElem := range sortBy {
		dataArray2D, err = Sort(dataArray2D, sortByElem)
		fmt.Printf("data:%#v\n", dataArray2D)
		if err != nil {
			return
		}
	}

	result = make([]interface{}, len(dataArray))
	i := 0
	for _, dataArray_ := range dataArray2D {
		for _, data := range dataArray_ {
			result[i] = data
			i++
		}
	}
	return
}

func Sort(dataArray2D [][]interface{}, sortByElem bson.DocElem) (results [][]interface{}, err error) {
	key := sortByElem.Name
	increment := sortByElem.Value.(int) == 1

	results = make([][]interface{}, 0)
	for _, dataArray := range dataArray2D {
		es := &ElementSlice{dataArray, key, increment, nil}
		sort.Sort(es)
		if es.err != nil {
			err = es.err
			return
		}
		for _, split := range es.Split() {
			results = append(results, split)
		}
		es = nil // help gc
	}
	return
}

type ElementSlice struct {
	elements  []interface{}
	key       string
	increment bool
	err       error
}

func (p *ElementSlice) Len() int {
	return len(p.elements)
}

func (p *ElementSlice) Less(i, j int) bool {
	result, err := p.compare(i, j)
	if err != nil {
		p.err = err
		return p.increment
	}

	if p.increment {
		return result == LT
	}
	return result == GT
}

func (p *ElementSlice) compare(i, j int) (op OP, err error) {
	elementI := reflect.ValueOf(p.elements[i])
	elementJ := reflect.ValueOf(p.elements[j])
	dataI, ok1 := parse.GetStructValueByFlag(p.key, elementI)
	dataJ, ok2 := parse.GetStructValueByFlag(p.key, elementJ)
	if !ok1 || !ok2 {
		return LT, nil // 默认升序排列
	}

	return compare(dataI, dataJ)
}

func compare(dataI, dataJ reflect.Value) (op OP, err error) {
	if dataI.Kind() != dataJ.Kind() {
		err = errors.New("data kind different:" + dataI.Kind().String() + "," + dataJ.Kind().String())
		return
	}
	if dataI.Type() != dataJ.Type() {
		err = errors.New("data type different:" + dataI.Type().String() + "," + dataJ.Type().String())
		return
	}

	switch dataI.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		if dataI.Int() < dataJ.Int() {
			return LT, nil
		} else if dataI.Int() == dataJ.Int() {
			return EQ, nil
		}
		return GT, nil
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uintptr, reflect.Uint16, reflect.Uint8:
		if dataI.Uint() < dataJ.Uint() {
			return LT, nil
		} else if dataI.Uint() == dataJ.Uint() {
			return EQ, nil
		}
		return GT, nil
	case reflect.Float64, reflect.Float32:
		if dataI.Float() < dataJ.Float() {
			return LT, nil
		} else if dataI.Float() == dataJ.Float() {
			return EQ, nil
		}
		return GT, nil
	case reflect.String:
		if dataI.String() < dataJ.String() {
			return LT, nil
		} else if dataI.String() == dataJ.String() {
			return EQ, nil
		}
		return GT, nil
	case reflect.Ptr, reflect.Interface:
		return compare(dataI.Elem(), dataJ.Elem())
	case reflect.Struct:
		if dataI.NumField() != dataJ.NumField() { // should never happend
			err = fmt.Errorf("data number field different, %d:%d", dataI.NumField(), dataJ.NumField())
			return
		}
		for i := 0; i < dataI.NumField(); i++ {
			result, err := compare(dataI.Field(i), dataJ.Field(i))
			if result != EQ || err != nil {
				return result, err
			}
		}
		return EQ, nil
	case reflect.Map:
		keys := dataI.MapKeys()
		for _, key := range keys {
			result, err := compare(dataI.MapIndex(key), dataJ.MapIndex(key))
			if result != EQ || err != nil {
				return result, err
			}
		}
		return GT, nil
	case reflect.Array, reflect.Slice:
		for i := 0; i < dataI.Len() || i < dataJ.Len(); i++ {
			if i == dataI.Len() || i == dataJ.Len() {
				if dataI.Len() < dataJ.Len() {
					return LT, nil
				} else if dataI.Len() > dataJ.Len() {
					return GT, nil
				}
			}
			result, err := compare(dataI.Index(i), dataJ.Index(i))
			if result != EQ || err != nil {
				return result, err
			}

		}
		return EQ, nil
	default:
		err = errors.New("not support kind:" + dataI.Kind().String())
		return
	}

	return // nerver here
}

func (p *ElementSlice) Swap(i, j int) {
	p.elements[i], p.elements[j] = p.elements[j], p.elements[i]
}

func (p *ElementSlice) Split() (splits [][]interface{}) {
	length := len(p.elements)
	if length == 0 {
		return [][]interface{}{}
	}

	if length == 1 {
		return [][]interface{}{p.elements}
	}
	index := 0
	splits = make([][]interface{}, 0)
	for i := 0; i < length-1; i++ {
		j := i + 1
		if op, _ := p.compare(i, j); op == EQ {
			if j == length-1 {
				splits = append(splits, p.elements[index:length])
				break
			}
			continue
		}
		splits = append(splits, p.elements[index:j])
		if j == length-1 {
			splits = append(splits, p.elements[j:length])
			break
		}
		index = j
	}
	return
}
