package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"

	"github.com/inokone/go-micro-saas/internal/auth/role"
	"github.com/inokone/go-micro-saas/internal/auth/user"
	"github.com/inokone/go-micro-saas/internal/common"
)

// FacebookSource is the source we use for `user.User`s registered from Google OAuth
const FacebookSource = "Facebook"

// state should be regenerated per auth request
var (
	FacebookState = "raw.ninja_facebook_random_csrf_string-c9551e5b-c326-4610-98dd-b1f78f76e25c"
)

// FacebookHandler is a handler for endpoints of Facebook OAuth based autentication
type FacebookHandler struct {
	key         string
	secret      string
	users       user.Storer
	jwt         *JWTHandler
	successURL  string
	redirectURL string
}

// NewFacebookHandler is a function creating an instance of `FacebookHandler`
func NewFacebookHandler(c common.AuthConfig, users user.Storer, jwt *JWTHandler) *FacebookHandler {
	return &FacebookHandler{
		key:         c.FacebookKey,
		secret:      c.FacebookSecret,
		users:       users,
		jwt:         jwt,
		successURL:  c.FrontendRoot + "/dashboard",
		redirectURL: c.BackendRoot + "/api/public/v1/auth/facebook/redirect",
	}
}

// Signin endpoint
// @Summary Signin is the authentication endpoint. Starts Facebook authentication process.
// @Schemes
// @Description Starts Facebook authentication process.
// @Accept json
// @Produce json
// @Router /auth/facebook [get]
func (h *FacebookHandler) Signin(g *gin.Context) {
	url := h.getConfig().AuthCodeURL(FacebookState)
	g.Redirect(http.StatusTemporaryRedirect, url)
}

// Redirect endpoint
// @Summary Redirect is the authentication callback endpoint. Authenticates/Registers users, sets up JWT token.
// @Schemes
// @Description Called by Facebook Auth when we have a result of the authentication process
// @Accept json
// @Produce text/html
// @Router /auth/facebook/redirect [get]
func (h *FacebookHandler) Redirect(g *gin.Context) {
	email, err := h.authenticateCode(g)
	if err != nil {
		g.AbortWithStatusJSON(http.StatusInternalServerError, common.StatusMessage{Message: err.Error()})
		return
	}

	usr, err := h.users.ByEmail(email)
	if err != nil {
		log.Debug("Populating new user from Facebook.")
		userData := &user.User{
			Email:    email,
			PassHash: "",
			Source:   FacebookSource,
			RoleID:   role.RoleCustomerUser,
			Status:   user.Confirmed,
			Enabled:  true,
		}
		if err = h.users.Store(userData); err != nil {
			log.WithError(err).Error("Can not store Facebook user.")
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
	if usr.Source != FacebookSource {
		g.AbortWithStatusJSON(http.StatusUnauthorized, common.StatusMessage{Message: "The provided email address is registered already with a different provider!"})
		return
	}
	h.jwt.Issue(g, usr.ID.String())
	g.Redirect(http.StatusTemporaryRedirect, h.successURL)
}

func (h *FacebookHandler) authenticateCode(g *gin.Context) (string, error) {
	var u UserDetails
	state := g.Query("state")
	if state != FacebookState {
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
	response, err := client.Get("https://graph.facebook.com/me?fields=id,name,email&access_token=" + token.AccessToken)
	if err != nil {
		log.WithError(err).Error("error getting userinfo")
		return "", fmt.Errorf("error getting userinfo: %s", err)
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&u)
	if err != nil {
		log.WithError(err).Error("user details invalid")
		return "", fmt.Errorf("user details invalid: %s", err)
	}

	return u.Email, nil
}

func (h *FacebookHandler) getConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.key,
		ClientSecret: h.secret,
		RedirectURL:  h.redirectURL,
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     facebook.Endpoint,
	}
}

// UserDetails is the user details structure for Facebook userdetails API
type UserDetails struct {
	ID      string
	Name    string
	Email   string
	Picture string
}
