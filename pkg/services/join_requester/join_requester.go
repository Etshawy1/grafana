package join_requester

import (
	"context"
)

type Service interface {
	Create(context.Context, *CreateJoinRequestCommand) (*JoinRequester, error)
	Delete(context.Context, int64) (*JoinRequester, error)
	GetByID(context.Context, int64) (*JoinRequester, error)
	Search(context.Context, *SearchJoinRequestQuery) ([]*JoinRequester, error)
}
