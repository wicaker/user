package repository

import (
	"fmt"
	"strings"
)

func filterRecordsQuery(criterias map[string]interface{}, orderBy *map[string]string) (query string, args []interface{}) {
	var (
		orderByCommand string
		orderByKeyword string
		argsLen        uint16 = 1
		whereCondition string
	)

	for i, criteria := range criterias {
		if criteria != nil {
			args = append(args, criteria)
			whereCondition = whereCondition + fmt.Sprintf(" AND %s=$%d", i, argsLen)
			argsLen++
		} else {
			whereCondition = whereCondition + fmt.Sprintf(" AND %s is NULL", i)
		}
	}

	if orderBy != nil {
		orderByCommand = " ORDER BY"
		for i, ob := range *orderBy {
			orderByKeyword = orderByKeyword + fmt.Sprintf(" %s %s,", i, ob)
		}
		orderByKeyword = strings.TrimSuffix(orderByKeyword, ",")
	}

	query = query + whereCondition + orderByCommand + orderByKeyword
	return query, args
}
