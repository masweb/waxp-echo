package filter

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

type Operator string

const (
	OpEq     Operator = "eq"
	OpNeq    Operator = "neq"
	OpLike   Operator = "like"
	OpGt     Operator = "gt"
	OpGte    Operator = "gte"
	OpLt     Operator = "lt"
	OpLte    Operator = "lte"
	OpIn     Operator = "in"
	OpNot    Operator = "not"
	OpIsNull Operator = "isnull"
)

var allOperators = []string{
	"isnull", "like", "gte", "lte", "gt", "lt", "neq", "not", "in", "eq",
}

type Filter struct {
	Column   string
	Operator Operator
	Value    string
}

type Result struct {
	WhereClause string
	Args        []any
}

type Builder struct {
	allowed map[string]string
	filters []Filter
}

func NewBuilder(allowedColumns map[string]string) *Builder {
	return &Builder{allowed: allowedColumns}
}

func (b *Builder) Parse(queryParams url.Values) error {
	keys := make([]string, 0, len(queryParams))
	for k := range queryParams {
		if strings.HasPrefix(k, "filter[") && strings.HasSuffix(k, "]") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, key := range keys {
		inner := key[7 : len(key)-1]
		values := queryParams[key]
		if len(values) == 0 {
			continue
		}

		column, op := parseKey(inner)

		colName, ok := b.allowed[column]
		if !ok {
			continue
		}

		b.filters = append(b.filters, Filter{
			Column:   colName,
			Operator: op,
			Value:    values[0],
		})
	}
	return nil
}

func parseKey(key string) (column string, op Operator) {
	for _, o := range allOperators {
		suffix := "_" + o
		if strings.HasSuffix(key, suffix) && len(key) > len(suffix) {
			return key[:len(key)-len(suffix)], Operator(o)
		}
	}
	return key, OpEq
}

func (b *Builder) Build(cursor *int64) Result {
	var clauses []string
	var args []any
	idx := 0

	if cursor != nil {
		idx++
		clauses = append(clauses, fmt.Sprintf("id > $%d", idx))
		args = append(args, *cursor)
	}

	for _, f := range b.filters {
		switch f.Operator {
		case OpIsNull:
			if strings.EqualFold(f.Value, "true") {
				clauses = append(clauses, fmt.Sprintf("%s IS NULL", f.Column))
			} else {
				clauses = append(clauses, fmt.Sprintf("%s IS NOT NULL", f.Column))
			}

		case OpLike:
			idx++
			clauses = append(clauses, fmt.Sprintf("%s ILIKE $%d", f.Column, idx))
			args = append(args, "%"+f.Value+"%")

		case OpEq:
			idx++
			clauses = append(clauses, fmt.Sprintf("%s = $%d", f.Column, idx))
			args = append(args, f.Value)

		case OpNeq, OpNot:
			idx++
			clauses = append(clauses, fmt.Sprintf("%s != $%d", f.Column, idx))
			args = append(args, f.Value)

		case OpGt:
			idx++
			clauses = append(clauses, fmt.Sprintf("%s > $%d", f.Column, idx))
			args = append(args, f.Value)

		case OpGte:
			idx++
			clauses = append(clauses, fmt.Sprintf("%s >= $%d", f.Column, idx))
			args = append(args, f.Value)

		case OpLt:
			idx++
			clauses = append(clauses, fmt.Sprintf("%s < $%d", f.Column, idx))
			args = append(args, f.Value)

		case OpLte:
			idx++
			clauses = append(clauses, fmt.Sprintf("%s <= $%d", f.Column, idx))
			args = append(args, f.Value)

		case OpIn:
			vals := strings.Split(f.Value, ",")
			placeholders := make([]string, len(vals))
			for i, v := range vals {
				idx++
				placeholders[i] = fmt.Sprintf("$%d", idx)
				args = append(args, strings.TrimSpace(v))
			}
			clauses = append(clauses, fmt.Sprintf("%s IN (%s)", f.Column, strings.Join(placeholders, ", ")))
		}
	}

	if len(clauses) == 0 {
		return Result{Args: args}
	}

	return Result{
		WhereClause: " WHERE " + strings.Join(clauses, " AND "),
		Args:        args,
	}
}
