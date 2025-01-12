package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	goauth2 "google.golang.org/api/oauth2/v2"

	"github.com/inokone/go-micro-saas/internal/auth/role"
	"github.com/inokone/go-micro-saas/internal/auth/user"
	"github.com/inokone/go-micro-saas/internal/common"
)

// GoogleSource is the source we use for `user.User`s registered from Google OAuth
const GoogleSource = "Google"

// state should be regenerated per auth request
var (
	GoogleState = "raw.ninja.google.random.csrf.string-c9551e5b-c326-4610-98dd-b1f78f76e25c"
)

// GoogleHandler is a handler for endpoints of Google OAuth based autentication
type GoogleHandler struct {
	key         string
	secret      string
	users       user.Storer
	jwt         *JWTHandler
	successURL  string
	redirectURL string
}

// NewGoogleHandler is a function creating an instance of `GoogleHandler`
func NewGoogleHandler(c common.AuthConfig, users user.Storer, jwt *JWTHandler) *GoogleHandler {
	return &GoogleHandler{
		key:         c.GoogleKey,
		secret:      c.GoogleSecret,
		users:       users,
		jwt:         jwt,
		successURL:  c.FrontendRoot + "/dashboard",
		redirectURL: c.BackendRoot + "/api/public/v1/auth/google/redirect",
	}
}

// Signin endpoint
// @Summary Signin is the authentication endpoint. Starts Google authentication process.
// @Schemes
// @Description Starts Google authentication process.
// @Accept json
// @Produce json
// @Router /auth/google [get]
func (h *GoogleHandler) Signin(g *gin.Context) {
	url := h.getConfig().AuthCodeURL(GoogleState)
	g.Redirect(http.StatusTemporaryRedirect, url)
}

// Redirect endpoint
// @Summary Redirect is the authentication callback endpoint. Authenticates/Registers users, sets up JWT token.
// @Schemes
// @Description Called by Google Auth when we have a result of the authentication process
// @Accept json
// @Produce text/html
// @Router /auth/google/redirect [get]
func (h *GoogleHandler) Redirect(g *gin.Context) {
	email, err := h.authenticateCode(g)
	if err != nil {
		g.AbortWithStatusJSON(http.StatusInternalServerError, common.StatusMessage{Message: err.Error()})
		return
	}

	usr, err := h.users.ByEmail(email)
	if err != nil {
		log.Debug("Populating new user from Google.")
		userData := &user.User{
			// Name:  google_user.Name,
			// Photo: google_user.Picture,
			Email:    email,
			PassHash: "",
			Source:   GoogleSource,
			RoleID:   role.RoleCustomerUser,
			Status:   user.Confirmed,
			Enabled:  true,
		}
		if err = h.users.Store(userData); err != nil {
			log.WithError(err).Error("Can not store Google user.")
			g.AbortWithStatusJSON(http.StatusInternalServerError, common.StatusMessage{Message: "Something went wrong. Please contact our administrators!"})
			return
		}
		usr, err = h.users.ByEmail(email)
		if err != nil {
			log.WithError(err).Error("Can not load Google user.")
			g.AbortWithStatusJSON(http.StatusInternalServerError, common.StatusMessage{Message: "Something went wrong. Please contact our administrators!"})
			return
		}
	}
	if !usr.Enabled {
		g.AbortWithStatusJSON(http.StatusUnauthorized, common.StatusMessage{Message: "Your account has been deactivated. Please contact our administrators!"})
		return
	}
	if usr.Source != GoogleSource {
		g.AbortWithStatusJSON(http.StatusUnauthorized, common.StatusMessage{Message: "The provided email address is registered already with a different provider!"})
		return
	}
	h.jwt.Issue(g, usr.ID.String())
	g.Redirect(http.StatusTemporaryRedirect, h.successURL)
}

func (h *GoogleHandler) authenticateCode(g *gin.Context) (string, error) {
	state := g.Query("state")
	if state != GoogleState {
		log.Warn("invalid oauth state")
		return "", errors.New("invalid oauth state")
	}

	code := g.Query("code")
	token, err := h.getConfig().Exchange(context.Background(), code)
	if err != nil {
		log.WithError(err).Error("token exchange error")
		return "", fmt.Errorf("token exchange error: %s", err)
	}

	client := h.getConfig().Client(context.Background(), token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.WithError(err).Error("error getting userinfo")
		return "", fmt.Errorf("error getting userinfo: %s", err)
	}

	//nolint:staticcheck
	userinfoService, err := goauth2.New(client)
	if err != nil {
		log.WithError(err).Error("error creating userinfo")
		return "", fmt.Errorf("error creating userinfo service: %s", err)
	}

	userinfo, err := goauth2.NewUserinfoV2MeService(userinfoService).Get().Context(g).Do()
	if err != nil {
		log.WithError(err).Error("error getting userinfo")
		return "", fmt.Errorf("error getting userinfo: %s", err)
	}

	defer response.Body.Close()

	return userinfo.Email, nil
}

func (h *GoogleHandler) getConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.key,
		ClientSecret: h.secret,
		RedirectURL:  h.redirectURL,
		Scopes:       []string{goauth2.UserinfoEmailScope}, // goauth2.UserinfoProfileScope},
		Endpoint:     google.Endpoint,
	}
}
