package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
)

func GetRateLimit(ctx utils.RouteContext, authentication utils.AuthenticationContext, res http.ResponseWriter, _ *http.Request) {
	if authentication.BypassCache && ctx.Config.Endpoints.RateLimit {
		res.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(res, fmt.Sprintf(`{"rate_limit_remaining": %d, "rate_limit_reset": %d}`, utils.RateLimitRemaining, utils.RateLimitReset))
	} else {
		res.WriteHeader(http.StatusNotFound)
	}
}
