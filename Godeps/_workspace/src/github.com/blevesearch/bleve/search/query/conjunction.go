//  Copyright (c) 2014 Couchbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package query

import (
	"encoding/json"

	"github.com/blevesearch/bleve/index"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/searcher"
)

type ConjunctionQuery struct {
	Conjuncts []Query `json:"conjuncts"`
	Boost     *Boost  `json:"boost,omitempty"`
}

// NewConjunctionQuery creates a new compound Query.
// Result documents must satisfy all of the queries.
func NewConjunctionQuery(conjuncts []Query) *ConjunctionQuery {
	return &ConjunctionQuery{
		Conjuncts: conjuncts,
	}
}

func (q *ConjunctionQuery) SetBoost(b float64) {
	boost := Boost(b)
	q.Boost = &boost
}

func (q *ConjunctionQuery) AddQuery(aq ...Query) {
	for _, aaq := range aq {
		q.Conjuncts = append(q.Conjuncts, aaq)
	}
}

func (q *ConjunctionQuery) Searcher(i index.IndexReader, m mapping.IndexMapping, explain bool) (search.Searcher, error) {
	ss := make([]search.Searcher, len(q.Conjuncts))
	for in, conjunct := range q.Conjuncts {
		var err error
		ss[in], err = conjunct.Searcher(i, m, explain)
		if err != nil {
			for _, searcher := range ss {
				if searcher != nil {
					_ = searcher.Close()
				}
			}
			return nil, err
		}
	}
	return searcher.NewConjunctionSearcher(i, ss, explain)
}

func (q *ConjunctionQuery) Validate() error {
	for _, q := range q.Conjuncts {
		if q, ok := q.(ValidatableQuery); ok {
			err := q.Validate()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (q *ConjunctionQuery) UnmarshalJSON(data []byte) error {
	tmp := struct {
		Conjuncts []json.RawMessage `json:"conjuncts"`
		Boost     *Boost            `json:"boost,omitempty"`
	}{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	q.Conjuncts = make([]Query, len(tmp.Conjuncts))
	for i, term := range tmp.Conjuncts {
		query, err := ParseQuery(term)
		if err != nil {
			return err
		}
		q.Conjuncts[i] = query
	}
	q.Boost = tmp.Boost
	return nil
}