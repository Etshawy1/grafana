package caching

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana/pkg/infra/remotecache"
	"github.com/grafana/grafana/pkg/infra/usagestats"
	"github.com/grafana/grafana/pkg/services/secrets"
	"github.com/grafana/grafana/pkg/setting"
)

const (
	XCacheHeader   = "X-Cache"
	StatusHit      = "HIT"
	StatusMiss     = "MISS"
	StatusBypass   = "BYPASS"
	StatusError    = "ERROR"
	StatusDisabled = "DISABLED"
)

type CacheQueryResponseFn func(context.Context, *backend.QueryDataResponse)
type CacheResourceResponseFn func(context.Context, *backend.CallResourceResponse)

type CachedQueryDataResponse struct {
	// The cached data response associated with a query, or nil if no cached data is found
	Response *backend.QueryDataResponse
	// A function that should be used to cache a QueryDataResponse for a given query.
	// It can be set to nil by the method implementation (if there is an error, for example), so it should be checked before being called.
	UpdateCacheFn CacheQueryResponseFn
}

type CachedResourceDataResponse struct {
	// The cached response associated with a resource request, or nil if no cached data is found
	Response *backend.CallResourceResponse
	// A function that should be used to cache a CallResourceResponse for a given resource request.
	// It can be set to nil by the method implementation (if there is an error, for example), so it should be checked before being called.
	// Because plugins can send multiple responses asynchronously, the implementation should be able to handle multiple calls to this function for one request.
	UpdateCacheFn CacheResourceResponseFn
}

func ProvideCachingService(cfg *setting.Cfg,
	usageStats usagestats.Service, secretsService secrets.Service) *OSSCachingService {

	cacheService, err := remotecache.ProvideService(cfg, nil, usageStats, secretsService)
	if err != nil {
		backend.Logger.Error("Failed to initialize caching service", err)
	}

	backend.Logger.Info("Caching service initialized with connection", cfg.RemoteCacheOptions.ConnStr)

	return &OSSCachingService{
		cache: cacheService,
	}
}

type CachingService interface {
	// HandleQueryRequest uses a QueryDataRequest to check the cache for any existing results for that query.
	// If none are found, it should return false and a CachedQueryDataResponse with an UpdateCacheFn which can be used to update the results cache after the fact.
	// This function may populate any response headers (accessible through the context) with the cache status using the X-Cache header.
	HandleQueryRequest(context.Context, *backend.QueryDataRequest) (bool, CachedQueryDataResponse)
	// HandleResourceRequest uses a CallResourceRequest to check the cache for any existing results for that request. If none are found, it should return false.
	// This function may populate any response headers (accessible through the context) with the cache status using the X-Cache header.
	HandleResourceRequest(context.Context, *backend.CallResourceRequest) (bool, CachedResourceDataResponse)
}

// Implementation of interface - does nothing
type OSSCachingService struct {
	cache *remotecache.RemoteCache
}

func (s *OSSCachingService) HandleQueryRequest(ctx context.Context, req *backend.QueryDataRequest) (bool, CachedQueryDataResponse) {

	cacheKey := s.getCacheKeyFromRequest(req)
	queryCachingTTL := s.getQueryCachingTTLFromRequest(req)

	// delete from cache or cache disabled
	if queryCachingTTL == 0 {
		s.cache.DeleteWithPrefix(ctx, s.getPanelCacheKeyPrefixFromRequest(req))
		return false, CachedQueryDataResponse{
			Response:      nil,
			UpdateCacheFn: nil,
		}
	}

	// check cache
	hit, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		return false, CachedQueryDataResponse{
			Response:      nil,
			UpdateCacheFn: s.updateCacheFunction,
		}
	}

	// unmarshal cached response
	var cachedResponse *backend.QueryDataResponse

	gzReader, err := gzip.NewReader(bytes.NewReader(hit))
	if err != nil {
		backend.Logger.Error("Failed to gzip decode cached QueryDataResponse", err)
		return false, CachedQueryDataResponse{
			Response:      nil,
			UpdateCacheFn: s.updateCacheFunction,
		}
	}

	decompressedGzip, err := io.ReadAll(gzReader)
	gzReader.Close()
	if err != nil {
		backend.Logger.Error("Failed to gzip decode cached QueryDataResponse", err)
		return false, CachedQueryDataResponse{
			Response:      nil,
			UpdateCacheFn: s.updateCacheFunction,
		}
	}

	err = json.Unmarshal(decompressedGzip, &cachedResponse)
	if err != nil {
		backend.Logger.Error("Failed to json unmarshal cached QueryDataResponse", err)
		return false, CachedQueryDataResponse{
			Response:      nil,
			UpdateCacheFn: s.updateCacheFunction,
		}
	}

	// return cached response
	return true, CachedQueryDataResponse{
		Response:      cachedResponse,
		UpdateCacheFn: s.updateCacheFunction,
	}
}

func (s *OSSCachingService) updateCacheFunction(ctx context.Context, res *backend.QueryDataResponse) {
	encoded, err := json.Marshal(res)
	if err != nil {
		backend.Logger.Error("Failed to json encode QueryDataResponse to be cached", err)
		return
	}
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(encoded); err != nil {
		backend.Logger.Error("Failed to gzip encode QueryDataResponse to be cached", err)
		return
	}
	gz.Close()

	req := ctx.Value("req").(*backend.QueryDataRequest)
	cacheKey := s.getCacheKeyFromRequest(req)
	queryCachingTTL := s.getQueryCachingTTLFromRequest(req)
	if queryCachingTTL != 0 {
		backend.Logger.Info("Caching QueryDataResponse", "cacheKey: ", cacheKey, "queryCachingTTL in ms: ", queryCachingTTL, "fromTime: ", req.Queries[0].TimeRange.From.String(), "toTime: ", req.Queries[0].TimeRange.To.String())
		err := s.cache.Set(ctx, cacheKey, b.Bytes(), time.Duration(queryCachingTTL*1000000000))
		if err != nil {
			backend.Logger.Error("Failed to cache QueryDataResponse", err)
		}
	}
}

func (s *OSSCachingService) getCacheKeyFromRequest(req *backend.QueryDataRequest) string {
	var queryCachingTTL = s.getQueryCachingTTLFromRequest(req)
	var startBin = req.Queries[0].TimeRange.From.Truncate(time.Duration(queryCachingTTL * 1000000000))
	var period = req.Queries[0].TimeRange.To.Sub(req.Queries[0].TimeRange.From).Round(time.Minute).String()
	var dashboardVars = ""
	var dashboardVarSortedKeys []string
	for key, _ := range req.Headers {
		if strings.HasPrefix(key, "http_X-Dashboard-Var") {
			dashboardVarSortedKeys = append(dashboardVarSortedKeys, key)
		}
	}
	sort.Strings(dashboardVarSortedKeys)
	for _, key := range dashboardVarSortedKeys {
		dashboardVars += (key[len("http_X-Dashboard-Var")+1:] + ":" + req.Headers[key] + "_")
	}

	return s.getPanelCacheKeyPrefixFromRequest(req) + "_" + dashboardVars + period + "_" + startBin.String()
}

func (s *OSSCachingService) getPanelCacheKeyPrefixFromRequest(req *backend.QueryDataRequest) string {
	return req.Headers["http_X-Dashboard-Uid"] + "_" + req.Headers["http_X-Datasource-Uid"] + "_" + req.Headers["http_X-Grafana-Org-Id"] + "_" + req.Headers["http_X-Panel-Id"]
}

func (s *OSSCachingService) getQueryCachingTTLFromRequest(req *backend.QueryDataRequest) float64 {
	rawQueryProp := make(map[string]any)
	queryBytes, err := req.Queries[0].JSON.MarshalJSON()
	if err != nil {
		backend.Logger.Error("Failed to json marshal query JSON to get queryCachingTTL", err)
		return 0
	}

	err = json.Unmarshal(queryBytes, &rawQueryProp)
	if err != nil {
		backend.Logger.Error("Failed to json unmarshal query JSON to get queryCachingTTL", err)
		return 0
	}

	queryCachingTTL, ok := rawQueryProp["queryCachingTTL"]
	if ok && queryCachingTTL != nil {
		return queryCachingTTL.(float64)
	}
	return 0
}

func (s *OSSCachingService) HandleResourceRequest(ctx context.Context, req *backend.CallResourceRequest) (bool, CachedResourceDataResponse) {
	return false, CachedResourceDataResponse{}
}

var _ CachingService = &OSSCachingService{}
