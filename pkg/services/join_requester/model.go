package join_requester

import (
	"time"

	"github.com/grafana/grafana/pkg/services/org"
)

type JoinRequester struct {
	ID            int64     `xorm:"pk autoincr 'id'" json:"id"`
	OrgID         int64     `xorm:"org_id" json:"orgId"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	Justification string    `json:"justification"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated,omitempty"`
}

type CreateJoinRequestCommand struct {
	OrgID         int64
	Email         string
	Role          org.RoleType `json:"role" binding:"Required"`
	Justification string       `json:"justification" binding:"Required"`
}

type SearchJoinRequestQuery struct {
	OrgID         int64 `xorm:"org_id"`
	Email         string
	Role          string `json:"role"`
	Justification string `json:"justification"`
}

type GetUserByLoginQuery struct {
	LoginOrEmail string
}

type GetUserByEmailQuery struct {
	Email string
}

// implement Conversion interface to define custom field mapping (xorm feature)
type AuthModuleConversion []string

func (auth *AuthModuleConversion) FromDB(data []byte) error {
	auth_module := string(data)
	*auth = []string{auth_module}
	return nil
}

// Just a stub, we don't want to write to database
func (auth *AuthModuleConversion) ToDB() ([]byte, error) {
	return []byte{}, nil
}

type Filter interface {
	WhereCondition() *WhereCondition
	InCondition() *InCondition
	JoinCondition() *JoinCondition
}

type WhereCondition struct {
	Condition string
	Params    interface{}
}

type InCondition struct {
	Condition string
	Params    interface{}
}

type JoinCondition struct {
	Operator string
	Table    string
	Params   string
}

type SearchJoinRequesterFilter interface {
	GetFilter(filterName string, params []string) Filter
	GetFilterList() map[string]FilterHandler
}

type FilterHandler func(params []string) (Filter, error)

const (
	QuotaTargetSrv string = "join_requester"
	QuotaTarget    string = "join_requester"
)
