package api

import (
	"crypto/subtle"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/lejianwen/rustdesk-api/v2/global"
	request "github.com/lejianwen/rustdesk-api/v2/http/request/api"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

type Server struct {
}

type verifyTokenResult struct {
	Ok       bool   `json:"ok"`
	UserId   uint   `json:"user_id"`
	Username string `json:"username"`
	Reason   string `json:"reason"`
}

// VerifyToken
// @Tags server
// @Summary verify a client access token for hbbs
// @Description Called by hbbs before accepting a PunchHoleRequest when MUST_LOGIN=api.
// @Accept json
// @Produce json
// @Param body body request.VerifyTokenForm true "token"
// @Success 200 {object} response.Response
// @Router /server/verify-token [post]
func (s *Server) VerifyToken(c *gin.Context) {
	secret := global.Config.Rustdesk.VerifySecret
	if secret == "" {
		// viper does not apply AutomaticEnv to keys missing from config.yaml,
		// so read the env var directly: the deployed conf may predate this key.
		secret = os.Getenv("RUSTDESK_API_RUSTDESK_VERIFY_SECRET")
	}
	if secret == "" {
		response.Error(c, "verify-token secret is not configured")
		return
	}
	given := c.GetHeader("X-Rdtest-Secret")
	if subtle.ConstantTimeCompare([]byte(secret), []byte(given)) != 1 {
		c.JSON(http.StatusForbidden, response.ErrorResponse{Error: "forbidden"})
		return
	}
	f := &request.VerifyTokenForm{}
	if err := c.ShouldBindBodyWith(f, binding.JSON); err != nil {
		response.Error(c, "invalid body: "+err.Error())
		return
	}
	ip := f.Ip
	if ip == "" {
		ip = c.ClientIP()
	}
	res := verifyClientToken(f.Token)
	// Smoke tests set X-Rdtest-NoAudit so they do not pollute the audit trail.
	if c.GetHeader("X-Rdtest-NoAudit") != "1" {
		_ = service.AllService.AuditService.CreateAuditLogin(&model.AuditLogin{
			Guid:     f.Guid,
			Uuid:     f.Uuid,
			Ip:       ip,
			UserId:   res.UserId,
			Username: res.Username,
			Result:   resultLabel(res),
			Reason:   res.Reason,
		})
	}
	response.Success(c, res)
}

func resultLabel(res verifyTokenResult) string {
	if res.Ok {
		return "ok"
	}
	return res.Reason
}

// verifyClientToken accepts the opaque access token stored by /api/login.
// It never trusts the JWT alone: the token row must still exist, be unexpired
// and belong to an enabled user, so logout and disabling take effect at once.
func verifyClientToken(token string) verifyTokenResult {
	if token == "" {
		return verifyTokenResult{Reason: "token is empty"}
	}
	u, ut := service.AllService.UserService.InfoByAccessToken(token)
	if ut.Id == 0 {
		return verifyTokenResult{Reason: "token not found"}
	}
	if ut.ExpiredAt < time.Now().Unix() {
		return verifyTokenResult{Reason: "token expired"}
	}
	if u.Id == 0 {
		return verifyTokenResult{Reason: "user not found"}
	}
	if !service.AllService.UserService.CheckUserEnable(u) {
		return verifyTokenResult{UserId: u.Id, Username: u.Username, Reason: "user disabled"}
	}
	return verifyTokenResult{Ok: true, UserId: u.Id, Username: u.Username, Reason: "ok"}
}
