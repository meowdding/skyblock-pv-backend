package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/internal"
)

func GetRateLimit(ctx internal.RouteContext, authentication internal.AuthenticationContext, res http.ResponseWriter, _ *http.Request) {
	if authentication.BypassCache && ctx.Config.Endpoints.RateLimit {
		res.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(res, fmt.Sprintf(`{"rate_limit_remaining": %d, "rate_limit_reset": %d}`, internal.RateLimitRemaining, internal.RateLimitReset))
	} else {
		res.WriteHeader(http.StatusNotFound)
	}
}
