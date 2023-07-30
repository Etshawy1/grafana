package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/models/roletype"
	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"
	"github.com/grafana/grafana/pkg/services/join_requester"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/services/user"
	"github.com/grafana/grafana/pkg/web"
)

// swagger:route GET /org/invites org_invites getPendingOrgInvites
//
// Get pending invites.
//
// Responses:
// 200: getPendingOrgInvitesResponse
// 401: unauthorisedError
// 403: forbiddenError
// 500: internalServerError
func (hs *HTTPServer) GetOrgJoinRequests(c *contextmodel.ReqContext) response.Response {
	query := join_requester.SearchJoinRequestQuery{OrgID: c.OrgID}

	queryResult, err := hs.joinRequestService.Search(c.Req.Context(), &query)
	if err != nil {
		return response.Error(500, "Failed to get join requests from db", err)
	}

	return response.JSON(http.StatusOK, queryResult)
}

// swagger:route DELETE /org/joinRequest/{join_request_id}/reject org_join_requests rejectJoinRequest
//
// Reject join request.
//
// Responses:
// 200: okResponse
// 401: unauthorisedError
// 403: forbiddenError
// 404: notFoundError
// 500: internalServerError
func (hs *HTTPServer) RejectJoinRequest(c *contextmodel.ReqContext) response.Response {
	joinRequestId, err := strconv.ParseInt(web.Params(c.Req)[":id"], 10, 64)
	if err != nil {
		return response.Error(500, "Failed to convert join request id to int", err)
	}
	_, err = hs.joinRequestService.Delete(c.Req.Context(), joinRequestId)
	if err != nil {
		return response.Error(500, "Failed to delete join request", err)
	}

	return response.Success("Join Request rejected")
}

// swagger:route DELETE /org/joinRequest/{join_request_id}/approve org_join_requests approveJoinRequest
//
// Approve join request.
//
// Responses:
// 200: okResponse
// 401: unauthorisedError
// 403: forbiddenError
// 404: notFoundError
// 500: internalServerError
func (hs *HTTPServer) ApproveJoinRequest(c *contextmodel.ReqContext) response.Response {
	joinRequestId, err := strconv.ParseInt(web.Params(c.Req)[":id"], 10, 64)
	if err != nil {
		return response.Error(500, "Failed to convert join request id to int", err)
	}
	joinRequest, err := hs.joinRequestService.GetByID(c.Req.Context(), joinRequestId)
	if err != nil {
		return response.Error(500, "Failed to get join request from database", err)
	}

	usr, err := hs.userService.GetByEmail(c.Req.Context(), &user.GetUserByEmailQuery{Email: joinRequest.Email})

	orgUser := org.OrgUser{
		OrgID:   joinRequest.OrgID,
		UserID:  usr.ID,
		Role:    roletype.RoleType(joinRequest.Role),
		Created: time.Now(),
		Updated: time.Now(),
	}

	_, err = hs.orgService.InsertOrgUser(c.Req.Context(), &orgUser)

	if err != nil {
		return response.Error(500, "Failed to create user from join request", err)
	}

	_, err = hs.joinRequestService.Delete(c.Req.Context(), joinRequestId)
	if err != nil {
		return response.Error(500, "Failed to delete join request", err)
	}

	return response.Success("user added to the organization")
}
