package fastlru

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

type TestCase struct {
	key        string
	val        interface{}
	input      []entry
	output     []interface{}
	err        error
	actionType string
}

type entry struct {
	key string
	val interface{}
}

type Person struct {
	Name string
	Age  int
}

var testCase []*TestCase

func init() {
	v1 := &Person{
		Name: "alex",
		Age:  18,
	}
	testCase = append(testCase,
		&TestCase{
			key:        "test1",
			output:     nil,
			err:        EmptyErr,
			actionType: "Get",
		},
		&TestCase{
			input:[]entry{
				{key: "test1",val:v1},
			},
			output: []interface{}{v1},
			err:        nil,
			actionType: "Set",
		},
		&TestCase{
			key:        "test1",
			val:        v1,
			output:     []interface{}{v1},
			err:        nil,
			actionType: "Get",
		},
		&TestCase{
			input:[]entry{
				{key: "test2",val:2},
				{key: "test3",val:3},
				{key: "test4",val:4},
			},
			output: []interface{}{4,3,2,v1},
			err:        nil,
			actionType: "Set",
		},
		&TestCase{
			key:        "test2",
			val:        2,
			output:     []interface{}{2,4,3,v1},
			err:        nil,
			actionType: "Get",
		},
		&TestCase{
			input:[]entry{
				{key: "test7",val:7},
				{key: "test6",val:6},
				{key: "test5",val:5},
				{key: "test3",val:33},
				{key: "test8",val:8},
				{key: "test9",val:9},
				{key: "test10",val:10},
			},
			output: []interface{}{10,9,8,33,5,6,7,2,4,v1},
			err:        nil,
			actionType: "Set",
		},
		&TestCase{
			input:[]entry{
				{key: "test5",val:55},
				{key: "test11",val:11},
			},
			output: []interface{}{11,55,10,9,8,33,6,7,2,4},
			err:        nil,
			actionType: "Set",
		},)
}

func TestLRU(t *testing.T) {
	ctx := context.Background()
	lruCache := NewLru(10)

	for i := range testCase {
		var result []interface{}
		if testCase[i].actionType == "Get" {
			val, err := lruCache.Get(ctx, testCase[i].key)
			assert.Equal(t, testCase[i].err, err)
			assert.Equal(t, testCase[i].val, val)
		} else if testCase[i].actionType == "Set" {
			for j := range testCase[i].input {
				input := testCase[i].input[j]
				err := lruCache.Set(ctx, input.key, input.val)
				assert.Equal(t, nil, err)
			}
		}
		result = lruCache.GetAllValue(ctx)
		if len(result) > 0 {
			for j := range result {
				assert.Equal(t, testCase[i].output[j], result[j])
			}
		}
	}
}

func BenchmarkLru(t *testing.B) {
	lru := NewLru(10)
	ctx := context.Background()
	for i:= 0; i<t.N; i++{
		randNum := rand.Intn(20)
		key := fmt.Sprintf("%d", randNum)
		if rand.Intn(2) == 1 {
			_ = lru.Set(ctx, key, randNum)
		}else{
			_, _ = lru.Get(ctx, key)
		}
	}
}