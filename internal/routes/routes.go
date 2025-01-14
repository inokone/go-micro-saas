package routes

import (
	"github.com/cskr/pubsub/v2"
	"github.com/gin-gonic/gin"
	"github.com/inokone/go-micro-saas/internal/auth"
	"github.com/inokone/go-micro-saas/internal/auth/account"
	"github.com/inokone/go-micro-saas/internal/auth/role"
	"github.com/inokone/go-micro-saas/internal/auth/user"
	"github.com/inokone/go-micro-saas/internal/common"
	"github.com/inokone/go-micro-saas/internal/history"
	"github.com/inokone/go-micro-saas/internal/mail"
)

// Storers is a struct to collect all `Storer` entities used by the application
type Storers struct {
	Users    user.Storer
	Roles    role.Storer
	Accounts account.Storer
	History  history.Storer
}

// InitPrivate is a function to initialize handler mapping for URLs protected with CORS
func InitPrivate(private *gin.RouterGroup, st Storers, c *common.AppConfig, ps *pubsub.PubSub[string, common.Event]) error {
	rc, err := common.NewRecaptchaValidator(c.Auth.RecaptchaProjectID, c.Auth.RecaptchaKey, c.PathFor(c.Auth.RecaptchaAppCreds))
	if err != nil {
		return err
	}

	var (
		mailer = mail.NewService(c.Mail, ps)
		m      = auth.NewJWTHandler(st.Users, c.Auth)
		a      = auth.NewHandler(st.Users, st.Accounts, m, rc)
		ac     = account.NewHandler(st.Users, st.Accounts, mailer, c.Auth, rc)
		u      = user.NewHandler(st.Users)
		r      = role.NewHandler(st.Roles)
		h      = history.NewHandler(st.History)
	)

	private.GET("healthcheck", common.Healthcheck)

	g := private.Group("/auth")
	{
		g.POST("/signin", a.Signin)
		g.GET("/signout", a.Signout)
	}

	g = private.Group("/account")
	{
		g.POST("/signup", ac.Signup)
		g.GET("/confirm", ac.Confirm)
		g.PUT("/resend", ac.ResendConfirmation)
		g.PUT("/recover", ac.Recover)
		g.PUT("/password/reset", ac.ResetPassword)
		g.PUT("/password/change", m.Validate, ac.ChangePassword)
		g.GET("/profile", m.Validate, u.Profile)
	}

	g = private.Group("/users")
	{
		g.GET("/", m.ValidateAdmin, u.List)
		g.PUT("/:id", m.Validate, u.Update)
		g.PATCH("/:id", m.ValidateAdmin, u.Patch)
		g.PUT("/:id/enabled", m.ValidateAdmin, u.SetEnabled)
		g.GET("/:id/history", m.Validate, h.List)
	}

	g = private.Group("/roles", m.ValidateAdmin)
	{
		g.GET("/", r.List)
		g.PUT("/:id", r.Update)
	}

	return nil
}

// InitPublic is a function to initialize handler mapping for URLs not protected with CORS
func InitPublic(public *gin.RouterGroup, st Storers, c *common.AppConfig) {
	m := auth.NewJWTHandler(st.Users, c.Auth)
	gt := auth.NewGoogleHandler(*c.Auth, st.Users, m)
	ft := auth.NewFacebookHandler(*c.Auth, st.Users, m)

	g := public.Group("/auth")
	{
		g.GET("/google/redirect", gt.Redirect)
		g.GET("/google", gt.Signin)
		g.GET("/facebook/redirect", ft.Redirect)
		g.GET("/facebook", ft.Signin)
	}
}
