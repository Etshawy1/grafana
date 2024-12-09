package join_requesterimpl

import (
	"context"

	"github.com/grafana/grafana/pkg/infra/db"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/join_requester"
	"github.com/grafana/grafana/pkg/services/sqlstore/migrator"
	"github.com/grafana/grafana/pkg/services/user"
	"github.com/grafana/grafana/pkg/setting"
)

type store interface {
	Insert(context.Context, *join_requester.JoinRequester) (int64, error)
	GetByID(context.Context, int64) (*join_requester.JoinRequester, error)
	Search(context.Context, *join_requester.SearchJoinRequestQuery) ([]*join_requester.JoinRequester, error)
	Delete(context.Context, int64) (*join_requester.JoinRequester, error)
}

type sqlStore struct {
	db      db.DB
	dialect migrator.Dialect
	logger  log.Logger
	cfg     *setting.Cfg
}

func ProvideStore(db db.DB, cfg *setting.Cfg) sqlStore {
	return sqlStore{
		db:      db,
		dialect: db.GetDialect(),
		cfg:     cfg,
		logger:  log.New("join_requester.store"),
	}
}

func (ss *sqlStore) Insert(ctx context.Context, cmd *join_requester.JoinRequester) (int64, error) {
	var err error
	err = ss.db.WithTransactionalDbSession(ctx, func(sess *db.Session) error {

		if _, err = sess.Insert(cmd); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}
	return cmd.ID, nil
}

func (ss *sqlStore) GetByID(ctx context.Context, joinRequestId int64) (*join_requester.JoinRequester, error) {
	var joinRequest join_requester.JoinRequester

	err := ss.db.WithDbSession(ctx, func(sess *db.Session) error {
		has, err := sess.ID(&joinRequestId).
			Get(&joinRequest)

		if err != nil {
			return err
		} else if !has {
			return user.ErrUserNotFound
		}
		return nil
	})
	return &joinRequest, err
}

func (ss *sqlStore) Delete(ctx context.Context, joinRequestId int64) (*join_requester.JoinRequester, error) {
	joinRequest, err := ss.GetByID(ctx, joinRequestId)
	err = ss.db.WithDbSession(ctx, func(sess *db.Session) error {
		var rawSQL = "DELETE FROM " + ss.dialect.Quote("join_requester") + " WHERE id = ?"
		_, err := sess.Exec(rawSQL, joinRequestId)
		return err
	})
	if err != nil {
		return nil, err
	}
	return joinRequest, err
}

func (ss *sqlStore) Search(ctx context.Context, query *join_requester.SearchJoinRequestQuery) ([]*join_requester.JoinRequester, error) {
	queryResult := make([]*join_requester.JoinRequester, 0)
	err := ss.db.WithDbSession(ctx, func(dbSess *db.Session) error {
		rawSQL := `SELECT
	                jr.id             as id,
	                jr.org_id         as org_id,
	                jr.email          as email,
									jr.role           as role,
									jr.justification  as justification,
									jr.created				as created
	                FROM ` + ss.db.GetDialect().Quote("join_requester") + ` as jr`
		params := []interface{}{}

		if query.OrgID > 0 {
			rawSQL += ` WHERE jr.org_id=?`
			params = append(params, query.OrgID)
		}

		if query.Email != "" {
			if ss.cfg.CaseInsensitiveLogin {
				rawSQL += ` AND LOWER(jr.email)=LOWER(?)`
			} else {
				rawSQL += ` AND jr.email=?`
			}
			params = append(params, query.Email)
		}

		rawSQL += " ORDER BY jr.created desc"

		sess := dbSess.SQL(rawSQL, params...)
		err := sess.Find(&queryResult)
		return err
	})
	if err != nil {
		return nil, err
	}
	return queryResult, nil
}
