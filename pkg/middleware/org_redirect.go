package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/grafana/grafana/pkg/services/contexthandler"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/services/provisioning/utils"
	"github.com/grafana/grafana/pkg/services/user"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/web"
)

// OrgRedirect changes org and redirects users if the
// querystring `orgId` doesn't match the active org.
func OrgRedirect(cfg *setting.Cfg, userSvc user.Service, orgSvc org.Service) web.Handler {
	return func(res http.ResponseWriter, req *http.Request, c *web.Context) {
		orgIdValue := req.URL.Query().Get("orgId")
		orgId, err := strconv.ParseInt(orgIdValue, 10, 64)

		if err != nil || orgId == 0 {
			return
		}

		ctx := contexthandler.FromContext(req.Context())
		if !ctx.IsSignedIn {
			return
		}

		if orgId == ctx.OrgID {
			return
		}

		cmd := user.SetUsingOrgCommand{UserID: ctx.UserID, OrgID: orgId}
		if err := userSvc.SetUsingOrg(ctx.Req.Context(), &cmd); err != nil {
			err = utils.CheckOrgExists(ctx.Req.Context(), orgSvc, orgId)

			if err != nil {
				if ctx.IsApiRequest() {
					ctx.JsonApiErr(404, fmt.Sprintf("Organization with id %d Not found", orgId), nil)
				} else {
					http.Error(ctx.Resp, fmt.Sprintf("Organization with id %d Not found", orgId), http.StatusNotFound)
				}
				return
			}
			newURL := fmt.Sprintf("%s%s/%s", cfg.AppURL, "org/join-request", orgIdValue)
			c.Redirect(newURL, 302)
			return
		}

		urlParams := c.Req.URL.Query()
		qs := urlParams.Encode()

		if urlParams.Has("kiosk") && urlParams.Get("kiosk") == "" {
			urlParams.Del("kiosk")
			qs = fmt.Sprintf("%s&kiosk", urlParams.Encode())
		}

		newURL := fmt.Sprintf("%s%s?%s", cfg.AppURL, strings.TrimPrefix(c.Req.URL.Path, "/"), qs)

		c.Redirect(newURL, 302)
	}
}
