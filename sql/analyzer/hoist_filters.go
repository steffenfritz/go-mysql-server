package analyzer

import (
	"strings"

	"github.com/dolthub/go-mysql-server/sql/fixidx"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
)

// hoistOutOfScopeFilters pulls filters upwards into the parent scope
// to decorrelate subqueries for further optimizations.
//
// select * from xy where exists (select * from uv where x = 1)
// =>
// select * from xy where x = 1 and exists (select * from uv)
func hoistOutOfScopeFilters(ctx *sql.Context, a *Analyzer, n sql.Node, scope *plan.Scope, sel RuleSelector) (sql.Node, transform.TreeIdentity, error) {
	switch n.(type) {
	case *plan.TriggerBeginEndBlock:
		return n, transform.SameTree, nil
	default:
	}
	ret, same, filters, err := recurseSubqueryForOuterFilters(n, a, scope)
	if len(filters) != 0 {
		// todo empty table fold filters before here
		return n, transform.SameTree, nil
		//return n, transform.SameTree, fmt.Errorf("rule 'hoistOutOfScopeFilters' tried to hoist filters above root node")
	}
	return ret, same, err
}

// recurseSubqueryForOuterFilters recursively hoists filters that belong
// to an outer scope (maybe higher than the parent). We do a DFS for hoisting
// subquery filters. We do a BFS to extract hoistable filters from subquery
// expressions before checking the normalized subquery and its hoisted
// filters for further hoisting.
func recurseSubqueryForOuterFilters(n sql.Node, a *Analyzer, scope *plan.Scope) (sql.Node, transform.TreeIdentity, []sql.Expression, error) {
	var hoistFilters []sql.Expression
	lowestAllowedIdx := len(scope.Schema())
	var inScope TableAliases
	ret, same, err := transform.Node(n, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		sq, _ := n.(*plan.SubqueryAlias)
		if sq != nil {
			subScope := scope.NewScope(sq)
			newQ, same, hoisted, err := recurseSubqueryForOuterFilters(sq.Child, a, subScope)
			if err != nil {
				return n, transform.SameTree, err
			}
			if same {
				return n, transform.SameTree, nil
			}
			if len(hoisted) > 0 {
				hoistFilters = append(hoistFilters, hoisted...)
			}
			return sq.WithChild(newQ), transform.NewTree, nil
		}
		f, _ := n.(*plan.Filter)
		if f == nil {
			return n, transform.SameTree, nil
		}

		var keepFilters []sql.Expression
		allSame := transform.SameTree
		queue := expression.SplitConjunction(f.Expression)
		for len(queue) > 0 {
			e := queue[0]
			queue = queue[1:]

			var not bool
			if n, ok := e.(*expression.Not); ok {
				not = true
				e = n.Child
			}

			// (1) normalize subquery expressions
			// (1a) recurse downwards
			// (1b) add hoisted to queue
			// (1c) standardize subquery expression for hoisting
			var sq *plan.Subquery
			switch e := e.(type) {
			case *plan.InSubquery:
				sq, _ = e.Right.(*plan.Subquery)
			case *plan.ExistsSubquery:
				sq = e.Query
			default:
			}
			if sq != nil {
				children := e.Children()
				subScope := scope.NewScopeFromSubqueryExpression(n)
				newQ, same, hoisted, err := recurseSubqueryForOuterFilters(sq.Query, a, subScope)
				if err != nil {
					return n, transform.SameTree, err
				}
				allSame = allSame && same
				newSq := sq.WithQuery(newQ)
				children[len(children)-1] = newSq
				e, _ = e.WithChildren(children...)

				if len(hoisted) > 0 {
					newScopeFilters, _, err := fixidx.FixFieldIndexesOnExpressions(scope, a.LogFn(), n.Schema(), hoisted...)
					if err != nil {
						return n, transform.SameTree, err
					}
					if not {
						// hoisted are tied to parent NOT, more elegant simplification
						// required to expose individual expressions for further hoisting
						e = expression.JoinAnd(e, expression.JoinAnd(hoisted...))
					} else {
						queue = append(queue, newScopeFilters...)
					}
				}
			}

			if not {
				e = expression.NewNot(e)
			}

			if lowestAllowedIdx == 0 {
				// cannot hoist filters above root scope
				keepFilters = append(keepFilters, e)
				continue
			}

			// (2) evaluate if expression hoistable
			if inScope == nil {
				var err error
				inScope, err = getTableAliases(n, nil)
				if err != nil {
					return n, transform.SameTree, err
				}
			}
			foundRef, foundAlias := exprRefsTableSet(e, inScope)

			// (3) bucket filter into parent or current scope
			if !foundRef && !foundAlias {
				// belongs in outer scope
				hoistFilters = append(hoistFilters, e)
			} else {
				keepFilters = append(keepFilters, e)
			}
		}

		if len(hoistFilters) > 0 {
			allSame = transform.NewTree
		}
		if allSame {
			return n, transform.SameTree, nil
		}

		if len(keepFilters) == 0 {
			return f.Child, transform.NewTree, nil
		}
		ret := plan.NewFilter(expression.JoinAnd(keepFilters...), f.Child)
		return ret, transform.NewTree, nil
	})
	return ret, same, hoistFilters, err
}

// exprRefsTableSet returns |foundRef| if the expression directly
// references a table in the table set, and |foundAlias| if the expression
// references an alias that is scope-ambiguous.
func exprRefsTableSet(e sql.Expression, tables TableAliases) (foundRef, foundAlias bool) {
	transform.InspectExpr(e, func(e sql.Expression) bool {
		switch e := e.(type) {
		case *expression.GetField:
			tName := strings.ToLower(e.Table())
			if tName == "" {
				foundAlias = true
			} else if _, ok := tables[tName]; ok {
				foundRef = true
			}
		case *plan.Subquery:
			transform.InspectExpressions(e.Query, func(e sql.Expression) bool {
				subqRef, subqAlias := exprRefsTableSet(e, tables)
				foundRef = foundRef || subqRef
				foundAlias = foundAlias || subqAlias
				return foundRef && foundAlias
			})
		}
		return foundRef && foundAlias
	})
	return
}
