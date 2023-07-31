package join_requesterimpl

import (
	"context"

	"github.com/grafana/grafana/pkg/infra/db"
	"github.com/grafana/grafana/pkg/infra/localcache"
	"github.com/grafana/grafana/pkg/services/join_requester"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/services/quota"
	"github.com/grafana/grafana/pkg/services/supportbundles"
	"github.com/grafana/grafana/pkg/services/team"
	"github.com/grafana/grafana/pkg/setting"
)

type Service struct {
	store        store
	orgService   org.Service
	teamService  team.Service
	cacheService *localcache.CacheService
	cfg          *setting.Cfg
}

func ProvideService(
	db db.DB,
	orgService org.Service,
	cfg *setting.Cfg,
	teamService team.Service,
	cacheService *localcache.CacheService,
	quotaService quota.Service,
	bundleRegistry supportbundles.Service,
) (join_requester.Service, error) {
	store := ProvideStore(db, cfg)
	s := &Service{
		store:        &store,
		orgService:   orgService,
		cfg:          cfg,
		teamService:  teamService,
		cacheService: cacheService,
	}

	return s, nil
}

func (s *Service) Create(ctx context.Context, cmd *join_requester.CreateJoinRequestCommand) (*join_requester.JoinRequester, error) {

	// create user
	join_requester := &join_requester.JoinRequester{
		Email:         cmd.Email,
		OrgID:         cmd.OrgID,
		Role:          string(cmd.Role),
		Justification: cmd.Justification,
		Created:       timeNow(),
		Updated:       timeNow(),
	}

	_, err := s.store.Insert(ctx, join_requester)
	if err != nil {
		return nil, err
	}

	return join_requester, nil
}

func (s *Service) Delete(ctx context.Context, joinRequesterId int64) (*join_requester.JoinRequester, error) {
	// delete from all the stores
	return s.store.Delete(ctx, joinRequesterId)
}

func (s *Service) GetByID(ctx context.Context, joinRequesterId int64) (*join_requester.JoinRequester, error) {
	joinRequest, err := s.store.GetByID(ctx, joinRequesterId)
	if err != nil {
		return nil, err
	}
	return joinRequest, nil
}

func (s *Service) Search(ctx context.Context, query *join_requester.SearchJoinRequestQuery) ([]*join_requester.JoinRequester, error) {
	return s.store.Search(ctx, query)
}
