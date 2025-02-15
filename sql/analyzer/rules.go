// Copyright 2020-2021 Dolthub, Inc.
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

package analyzer

// OnceBeforeDefault contains the rules to be applied just once before the
// DefaultRules.
var OnceBeforeDefault = []Rule{
	{applyDefaultSelectLimitId, applyDefaultSelectLimit},
	{replaceCountStarId, replaceCountStar},
	{validateOffsetAndLimitId, validateLimitAndOffset},
	{validateCreateTableId, validateCreateTable},
	{validateExprSemId, validateExprSem},
	{validateCreateProcedureId, validateCreateProcedure},
	{resolveDropConstraintId, resolveDropConstraint},
	{resolveAlterColumnId, resolveAlterColumn},
	{validateDropTablesId, validateDropTables},
	{resolveCreateSelectId, resolveCreateSelect},
	{validateDropConstraintId, validateDropConstraint},
	{resolveUnionsId, resolveUnions},
	{resolveDescribeQueryId, resolveDescribeQuery}, //TODO
	{validateCreateTriggerId, validateCreateTrigger},
	{validateColumnDefaultsId, validateColumnDefaults},
	{validateReadOnlyDatabaseId, validateReadOnlyDatabase},
	{validateReadOnlyTransactionId, validateReadOnlyTransaction},
	{validateDatabaseSetId, validateDatabaseSet},
	{validateDeleteFromId, validateDeleteFrom},
	{validatePrivilegesId, validatePrivileges}, // Ensure that checking privileges happens after db, table  & table function resolution
	{evalFilterId, simplifyFilters},            //TODO inline?
	{hoistOutOfScopeFiltersId, hoistOutOfScopeFilters},
}

// DefaultRules to apply when analyzing nodes.
var DefaultRules = []Rule{
	{validateStarExpressionsId, validateStarExpressions}, //TODO
	{flattenTableAliasesId, flattenTableAliases},         //TODO
	{pushdownSubqueryAliasFiltersId, pushdownSubqueryAliasFilters},
	{pruneTablesId, pruneTables},
	{fixupAuxiliaryExprsId, fixupAuxiliaryExprs},
	{validateCheckConstraintId, validateCheckConstraints},
	{transformJoinApplyId, transformJoinApply},
	{resolveSubqueriesId, resolveSubqueries},
	{replaceCrossJoinsId, replaceCrossJoins},
	{moveJoinCondsToFilterId, moveJoinConditionsToFilter}, // depends on indexes being correct
}

var OnceAfterDefault = []Rule{
	{hoistSelectExistsId, hoistSelectExists},
	{finalizeUnionsId, finalizeUnions},
	{loadTriggersId, loadTriggers},
	{loadEventsId, loadEvents},
	{processTruncateId, processTruncate},
	{stripTableNameInDefaultsId, stripTableNamesFromColumnDefaults},
	{pushFiltersId, pushFilters},
	{optimizeJoinsId, optimizeJoins},
	{generateIndexScansId, generateIndexScans},
	{finalizeSubqueriesId, finalizeSubqueries},
	{subqueryIndexesId, applyIndexesFromOuterScope},
	{replaceSortPkId, replacePkSort},
	{setJoinScopeLenId, setJoinScopeLen},
	{eraseProjectionId, eraseProjection},
	{insertTopNId, insertTopNNodes},
	{applyHashInId, applyHashIn},
	{resolveInsertRowsId, resolveInsertRows},
	{applyTriggersId, applyTriggers},
	{applyProceduresId, applyProcedures},
	{assignRoutinesId, assignRoutines},
	{modifyUpdateExprsForJoinId, modifyUpdateExpressionsForJoin},
	{applyRowUpdateAccumulatorsId, applyUpdateAccumulators},
	{wrapWithRollbackId, wrapWritesWithRollback},
	{applyFKsId, applyForeignKeys},
}

// DefaultValidationRules to apply while analyzing nodes.
var DefaultValidationRules = []Rule{
	{validateResolvedId, validateIsResolved},
	{validateOrderById, validateOrderBy},
	{validateGroupById, validateGroupBy},
	{validateSchemaSourceId, validateSchemaSource},
	{validateIndexCreationId, validateIndexCreation},
	{validateOperandsId, validateOperands},
	{validateIntervalUsageId, validateIntervalUsage},
	{validateSubqueryColumnsId, validateSubqueryColumns},
	{validateUnionSchemasMatchId, validateUnionSchemasMatch},
	{validateAggregationsId, validateAggregations},
}

var OnceAfterAll = []Rule{
	{inlineSubqueryAliasRefsId, inlineSubqueryAliasRefs},
	{cacheSubqueryResultsId, cacheSubqueryResults},
	{cacheSubqueryAliasesInJoinsId, cacheSubqueryAliasesInJoins},
	{AutocommitId, addAutocommitNode},
	{TrackProcessId, trackProcess},
	{parallelizeId, parallelize},
	{clearWarningsId, clearWarnings},
}
