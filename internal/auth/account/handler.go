package account

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/inokone/go-micro-saas/internal/auth/role"
	"github.com/inokone/go-micro-saas/internal/auth/user"
	"github.com/inokone/go-micro-saas/internal/common"
	"github.com/inokone/go-micro-saas/internal/mail"
)

const (
	aDay     time.Duration = time.Hour * 24
	twoWeeks time.Duration = aDay * 14
)

var statusBadRequest common.StatusMessage = common.StatusMessage{Message: "Invalid user data provided!"}

// Handler is a struct for web handles related to authentication and authorization.
type Handler struct {
	users    user.Storer
	accounts Storer
	sender   *mail.Service
	config   *common.AuthConfig
	captcha  *common.RecaptchaValidator
}

// NewHandler creates a new `Handler`, based on the user persistence and the authentication configuration parameters.
func NewHandler(users user.Storer, accounts Storer, sender *mail.Service, config *common.AuthConfig, captcha *common.RecaptchaValidator) *Handler {
	return &Handler{
		users:    users,
		accounts: accounts,
		sender:   sender,
		config:   config,
		captcha:  captcha,
	}
}

// Signup is a method of `Handler`. Signs the user up for the application with username/password credentials.
// @Summary User signup endpoint
// @Schemes
// @Description Signs the user up for the application
// @Accept json
// @Produce json
// @Param data body user.SignupRequest true "User data provided for the signup"
// @Success 201 {object} user.Profile
// @Failure 400 {object} common.StatusMessage
// @Failure 500 {object} common.StatusMessage
// @Router /account/signup [post]
func (h *Handler) Signup(g *gin.Context) {
	var s user.SignupRequest
	if err := g.ShouldBindJSON(&s); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.ValidationMessage(err))
		return
	}

	if err := h.captcha.Verify(s.Captcha); err != nil {
		log.WithError(err).Error("Failed to verify captcha.")
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Captcha verification failed!"})
		return
	}

	usr, err := user.NewUser(s.Email, s.Password, s.FirstName, s.LastName)
	if err != nil {
		log.WithError(err).Error("Could not create new user")
		g.AbortWithStatusJSON(http.StatusInternalServerError, common.StatusMessage{Message: "Could not create user."})
		return
	}
	usr.RoleID = role.RoleCustomerUser
	if err = h.users.Store(usr); err != nil {
		log.WithError(err).Error("Could not store user")
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{
			Message: "User with this email already exist.",
		})
		return
	}

	err = h.confirmMail(usr)
	if err != nil {
		log.WithError(err).Error("Could not send e-mail confirmation")
		g.AbortWithStatusJSON(http.StatusInternalServerError, common.StatusMessage{
			Message: "Could not create user.",
		})
		return
	}

	g.JSON(http.StatusCreated, usr.AsProfile())
}

func (h *Handler) confirmMail(usr *user.User) error {
	state := Account{
		UserID:            usr.ID,
		CreatedAt:         time.Now(),
		ConfirmationToken: uuid.New().String(),
		ConfirmationTTL:   time.Now().Add(twoWeeks),
	}
	if err := h.accounts.Store(&state); err != nil {
		return err
	}
	url := h.config.FrontendRoot + "/confirm?token=" + state.ConfirmationToken
	return h.sender.EmailConfirmation(usr.Email, url)
}

// ResendConfirmation is a method of `Handler`. Resends email confirmation for an email address.
// @Summary Resends email confirmation endpoint
// @Schemes
// @Description Resends email confirmation for an email address.
// @Accept json
// @Produce json
// @Param data body account.ConfirmationResend true "The email to send the confirmation to"
// @Success 200 {object} common.StatusMessage
// @Failure 400 {object} common.StatusMessage
// @Failure 500 {object} common.StatusMessage
// @Router /account/resend [put]
func (h *Handler) ResendConfirmation(g *gin.Context) {
	var (
		s   ConfirmationResend
		err error
		usr *user.User
	)

	if err = g.ShouldBindJSON(&s); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, "Invalid confirmation resend data.")
		return
	}

	usr, err = h.users.ByEmail(s.Email)
	if err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, "Invalid confirmation resend e-mail.")
		return
	}

	err = h.resendMail(usr)
	if err != nil {
		log.WithError(err).Error("Could not send e-mail confirmation")
		g.AbortWithStatusJSON(http.StatusInternalServerError, common.StatusMessage{
			Message: "Could not send confirmation email.",
		})
		return
	}

	g.JSON(http.StatusOK, common.StatusMessage{
		Message: "Confirmation sent!",
	})
}

func (h *Handler) resendMail(usr *user.User) error {
	s, err := h.accounts.ByUser(usr.ID)
	if err != nil {
		return err
	}
	s.ConfirmationToken = uuid.New().String()
	s.ConfirmationTTL = time.Now().Add(twoWeeks)
	if err := h.accounts.Update(s); err != nil {
		return err
	}
	url := h.config.FrontendRoot + "/confirm?token=" + s.ConfirmationToken
	return h.sender.EmailConfirmation(usr.Email, url)
}

// Confirm is a method of `Handler`. Confirms the email of the user for the hash provided as URL parameter.
// @Summary Email confirmation endpoint
// @Schemes
// @Description Confirms the email address of the user
// @Accept json
// @Produce json
// @Param   token    query     string  true  "Token for the email confirmation"  Format(uuid)
// @Success 200 {object} common.StatusMessage
// @Failure 400 {object} common.StatusMessage
// @Failure 500 {object} common.StatusMessage
// @Router /account/confirm [get]
func (h *Handler) Confirm(g *gin.Context) {
	var (
		token   string
		account *Account
		err     error
		usr     *user.User
	)
	token = g.Query("token")
	account, err = h.accounts.ByConfirmToken(token)
	if err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Invalid token!"})
		return
	}
	if account.ConfirmationTTL.IsZero() || account.ConfirmationTTL.Before(time.Now()) {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Expired token, resend it please!"})
		return
	}
	account.ConfirmationTTL = time.Now()
	account.ConfirmationToken = ""
	account.Confirmed = true
	if err = h.accounts.Update(account); err != nil {
		log.WithError(err).Error("Failed to update account.")
		g.AbortWithStatusJSON(http.StatusInternalServerError, statusBadRequest)
		return
	}

	usr, err = h.users.ByID(account.UserID)
	if err != nil {
		log.WithError(err).Error("Failed to collect user.")
		g.AbortWithStatusJSON(http.StatusInternalServerError, statusBadRequest)
		return
	}

	usr.Status = user.Confirmed
	if err = h.users.Update(usr); err != nil {
		log.WithError(err).Error("Failed to update user.")
		g.AbortWithStatusJSON(http.StatusInternalServerError, statusBadRequest)
		return
	}

	g.JSON(http.StatusOK, common.StatusMessage{
		Message: "E-mail is confirmed!",
	})
}

// Recover initiates a password reset - sends an email to a user
// @Summary Recover account endpoint
// @Schemes
// @Description Send a password reset email to a user
// @Accept json
// @Produce json
// @Param data body account.Recovery true "The email to send the account recovery to"
// @Success 202 {object} common.StatusMessage
// @Failure 400 {object} common.StatusMessage
// @Failure 500 {object} common.StatusMessage
// @Router /account/recover [put]
func (h *Handler) Recover(g *gin.Context) {
	var (
		s   Recovery
		err error
		usr *user.User
	)

	if err = g.ShouldBindJSON(&s); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{
			Message: "Invalid recovery data.",
		})
		return
	}

	usr, err = h.users.ByEmail(s.Email)
	if err == nil {
		err = h.recoverMail(usr)
		if err != nil {
			log.WithError(err).Error("Could not send recovery e-mail")
		}
	}

	g.JSON(http.StatusAccepted, common.StatusMessage{
		Message: "Recover request accepted!",
	})
}

func (h *Handler) recoverMail(usr *user.User) error {
	s, err := h.accounts.ByUser(usr.ID)
	if err != nil {
		return err
	}
	s.RecoveryToken = uuid.New().String()
	s.RecoveryTTL = time.Now().Add(aDay)
	if err := h.accounts.Update(s); err != nil {
		return err
	}
	url := h.config.FrontendRoot + "/password/reset?token=" + s.RecoveryToken
	return h.sender.PasswordReset(usr.Email, url)
}

// ResetPassword resets the password of the logged in user - not implemented yet
// @Summary Reset password endpoint
// @Schemes
// @Description Resets the password of the logged in user
// @Accept json
// @Produce json
// @Param data body account.PasswordReset true "The token and new password to reset the current set password"
// @Success 200 {object} common.StatusMessage
// @Failure 400 {object} common.StatusMessage
// @Failure 500 {object} common.StatusMessage
// @Router /account/password/reset [put]
func (h *Handler) ResetPassword(g *gin.Context) {
	var (
		reset PasswordReset
		state *Account
		err   error
		usr   *user.User
	)

	if err = g.ShouldBindJSON(&reset); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{
			Message: "Invalid recovery data.",
		})
		return
	}

	state, err = h.accounts.ByRecoveryToken(reset.Token)
	if err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Invalid token!"})
		return
	}
	if state.RecoveryTTL.IsZero() || state.RecoveryTTL.Before(time.Now()) {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Expired token, please restart the recovery!"})
		return
	}
	state.RecoveryToken = ""
	state.LastRecovery = time.Now()
	if err = h.accounts.Update(state); err != nil {
		log.WithError(err).Error("Failed to update account.")
		g.AbortWithStatusJSON(http.StatusInternalServerError, statusBadRequest)
		return
	}

	usr, err = h.users.ByID(state.UserID)
	if err != nil {
		log.WithError(err).Error("Failed to collect user.")
		g.AbortWithStatusJSON(http.StatusInternalServerError, statusBadRequest)
		return
	}

	if err = usr.SetPassword(reset.Password); err != nil {
		log.WithError(err).Error("Failed to set password for the user.")
		g.AbortWithStatusJSON(http.StatusInternalServerError, statusBadRequest)
		return
	}

	if err = h.users.Update(usr); err != nil {
		log.WithError(err).Error("Failed to update user.")
		g.AbortWithStatusJSON(http.StatusInternalServerError, statusBadRequest)
		return
	}

	g.JSON(http.StatusOK, common.StatusMessage{Message: "Password updated!"})
}

// ChangePassword resets the password of the logged in user - not implemented yet
// @Summary Reset password endpoint
// @Schemes
// @Description Resets the password of the logged in user
// @Accept json
// @Produce json
// @Param data body account.PasswordChange true "The new and old passwords, required to update the password"
// @Success 200 {object} common.StatusMessage
// @Failure 400 {object} common.StatusMessage
// @Failure 500 {object} common.StatusMessage
// @Router /account/password/change [put]
func (h *Handler) ChangePassword(g *gin.Context) {
	var (
		chg PasswordChange
		err error
		usr *user.User
	)

	u, _ := g.Get("user")
	usr = u.(*user.User)

	if err = g.ShouldBindJSON(&chg); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Invalid change data."})
		return
	}

	if !usr.VerifyPassword(chg.Old) {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Incorrect old password."})
		return
	}

	if err = usr.SetPassword(chg.New); err != nil {
		log.WithError(err).Error("Failed to set password for the user.")
		g.AbortWithStatusJSON(http.StatusInternalServerError, statusBadRequest)
		return
	}

	if err = h.users.Update(usr); err != nil {
		log.WithError(err).Error("Failed to update user.")
		g.AbortWithStatusJSON(http.StatusInternalServerError, statusBadRequest)
		return
	}

	g.JSON(http.StatusOK, common.StatusMessage{Message: "Password updated!"})
}
