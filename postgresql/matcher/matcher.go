// Copyright 2021 Burak Sezer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package matcher

import (
	"github.com/buraksezer/dante/config"
	"github.com/buraksezer/dante/utils"
	pg_query "github.com/pganalyze/pg_query_go/v2"
	"github.com/valyala/fastjson"
)

var (
	pool              fastjson.ParserPool
	defaultSchemaName = []byte("public")
)

type Query struct {
	hierarchy map[string]map[string]struct{}
}

func (q *Query) add(schema, table string) {
	_, ok := q.hierarchy[schema]
	if !ok {
		q.hierarchy[schema] = make(map[string]struct{})
	}
	q.hierarchy[schema][table] = struct{}{}
}

func (q *Query) Match(c []*config.Cache, f func(t *config.Table) (bool, error)) (bool, error) {
	for _, cache := range c {
		s, ok := q.hierarchy[cache.Schema]
		if !ok {
			return false, nil
		}

		for i, table := range cache.Tables {
			if _, ok = s[table.Name]; !ok {
				continue
			}
			return f(cache.Tables[i])
		}
	}

	return false, nil
}

func discoveryHierarchy(q *Query, value *fastjson.Value) {
	if value.Type() == fastjson.TypeObject && value.Exists("RangeVar") {
		rangeVar := value.GetObject("RangeVar")
		schemaName := rangeVar.Get("schemaname").GetStringBytes()
		if schemaName == nil {
			schemaName = defaultSchemaName
		}
		schema := utils.ByteToString(schemaName)
		relname := rangeVar.Get("relname").GetStringBytes()
		table := utils.ByteToString(relname)
		q.add(schema, table)
		return
	}

	if value.Type() == fastjson.TypeObject {
		obj := value.GetObject()
		obj.Visit(func(key []byte, v *fastjson.Value) {
			discoveryHierarchy(q, v)
		})
	}
}

func Parse(query []byte) (*Query, error) {
	result, err := pg_query.ParseToJSON(utils.ByteToString(query))
	if err != nil {
		return nil, err
	}
	parser := pool.Get()
	defer pool.Put(parser)

	parsed, err := parser.Parse(result)
	if err != nil {
		return nil, err
	}

	obj, err := parsed.Object()
	if err != nil {
		return nil, err
	}
	stmts := obj.Get("stmts")
	values, err := stmts.Array()
	if err != nil {
		return nil, err
	}

	q := &Query{
		hierarchy: make(map[string]map[string]struct{}),
	}
	for _, value := range values {
		fromClause := value.Get("stmt").GetObject("SelectStmt").Get("fromClause").GetArray()
		for _, item := range fromClause {
			discoveryHierarchy(q, item)
		}
	}

	return q, nil
}
