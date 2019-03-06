//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package api

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/readers"
)

func listMessagesEndpoint(svc readers.MessageRepository) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(listMessagesReq)

		fmt.Println("requesting messages:" + req.chanID)

		if err := req.validate(); err != nil {
			return nil, err
		}

		messages := svc.ReadAll(req.chanID, req.offset, req.limit)

		return listMessagesRes{Messages: messages}, nil
	}
}
