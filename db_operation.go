package main

import (
	"context"
	"fmt"
	"strings"
)

func (db *pgProfileDB) createOperation(ctx context.Context, req operationReq) (int, error) {

	var ID int

	createString, numString, createQueryArgs := prepareOperationCreateQuery(req)

	if len(createQueryArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO operation_trace (%s) VALUES (%s) RETURNING id`, createString, numString),
			createQueryArgs...).Scan(&ID); err != nil {
			return 0, fmt.Errorf("problem creating operation_trace: %w", err)
		}

		return ID, nil
	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

func prepareOperationCreateQuery(req operationReq) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.Input != nil {
		createStrings = append(createStrings, "input")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Input)
	}

	if req.Output != nil {
		createStrings = append(createStrings, "output")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Output)
	}

	if req.Revert != nil {
		createStrings = append(createStrings, "revert")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Revert)
	}

	if req.RevertOutput != nil {
		createStrings = append(createStrings, "revert_output")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RevertOutput)
	}

	if req.Status != nil {
		createStrings = append(createStrings, "status")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Status)
	}

	if req.Type != nil {
		createStrings = append(createStrings, "type")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Type)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}
