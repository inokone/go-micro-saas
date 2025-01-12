package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/inokone/go-micro-saas/internal/auth/account"
	"github.com/inokone/go-micro-saas/internal/auth/user"
	"github.com/inokone/go-micro-saas/internal/common"
)

const (
	jwtTokenKey string = "Authorization"
)

var (
	statusInvalidCredentials = common.StatusMessage{Message: "User does not exist or password does not match!"}
	statusBadRequest         = common.StatusMessage{Message: "Invalid user data provided!"}
)

// Handler is a struct for web handles related to authentication and authorization.
type Handler struct {
	users   user.Storer
	auths   account.Storer
	jwt     *JWTHandler
	captcha *common.RecaptchaValidator
	service *Service
}

// NewHandler creates a new `Handler`, based on the user persistence.
func NewHandler(users user.Storer, auths account.Storer, jwt *JWTHandler, captcha *common.RecaptchaValidator) *Handler {
	return &Handler{
		users:   users,
		auths:   auths,
		jwt:     jwt,
		captcha: captcha,
		service: NewService(users, auths, jwt),
	}
}

// Signin is a method of `Handler`. Authenticates the user to the application, sets a JWT token on success in the cookies.
// @Summary User sign in endpoint
// @Schemes
// @Description Logs in the user, sets up the JWT authorization
// @Accept json
// @Produce json
// @Param data body user.Credentials true "Credentials provided for signing in"
// @Success 200 {object} common.StatusMessage
// @Failure 400 {object} common.StatusMessage
// @Failure 403 {object} common.StatusMessage
// @Failure 500 {object} common.StatusMessage
// @Router /account/signin [post]
func (h *Handler) Signin(g *gin.Context) {
	var (
		s   user.Credentials
		err error
		usr *user.User
	)

	err = g.ShouldBindJSON(&s)
	if err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, statusBadRequest)
		return
	}

	if err := h.captcha.Verify(s.Captcha); err != nil {
		log.WithError(err).Error("Failed to verify captcha.")
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Captcha verification failed!"})
		return
	}

	usr, err = h.users.ByEmail(s.Email)
	if err != nil {
		log.WithError(err).Error("Failed to collect user.")
		g.AbortWithStatusJSON(http.StatusBadRequest, statusInvalidCredentials)
		return
	}

	if !usr.Enabled {
		g.AbortWithStatusJSON(http.StatusUnauthorized, common.StatusMessage{Message: "Your account has been deactivated. Please contact our administrators!"})
		return
	}

	if usr.Source != "credentials" {
		g.AbortWithStatusJSON(http.StatusUnauthorized, common.StatusMessage{Message: "Your account can not be used with credentials!"})
		return
	}

	if err = h.service.ValidateCredentials(usr, s.Password); err != nil {
		switch e := err.(type) {
		case InvalidCredentials:
			g.AbortWithStatusJSON(http.StatusBadRequest, statusInvalidCredentials)
		case LockedUser:
			g.AbortWithStatusJSON(http.StatusForbidden, common.StatusMessage{
				Message: fmt.Sprintf("You have been locked out for failed credentials. You have to wait %v more seconds.", e.seconds),
			})
		default:
			g.AbortWithStatusJSON(http.StatusInternalServerError, common.StatusMessage{Message: "Unknown error"})
		}
		return
	}

	h.jwt.Issue(g, usr.ID.String())

	g.JSON(http.StatusOK, common.StatusMessage{
		Message: "Logged in!",
	})
}

// Logout is a method of `Handler`. Clears the JWT token from the cookies thus logging out the current user.
// @Summary Logout endpoint
// @Schemes
// @Description Logs out of the application, deletes the JWT token uased for authorization
// @Accept json
// @Produce json
// @Success 200 {object} common.StatusMessage
// @Router /account/signout [get]
func (h *Handler) Signout(g *gin.Context) {
	g.SetCookie(jwtTokenKey, "", 0, "", "", true, true)
	g.JSON(http.StatusOK, common.StatusMessage{Message: "Logged out successfully! See you!"})
}
