package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/inokone/go-micro-saas/internal/common"
)

// Handler is a struct for web handles related to application users.
type Handler struct {
	users Storer
}

// NewHandler creates a new `Handler`, based on the user persistence.
func NewHandler(users Storer) *Handler {
	return &Handler{
		users: users,
	}
}

// Profile is a method of `Handler`. Retrieves profile data of the user based on the JWT token in the request.
// @Summary Get user profile endpoint
// @Schemes
// @Description Gets the current logged in user
// @Accept json
// @Produce json
// @Success 200 {object} Profile
// @Failure 403 {object} common.StatusMessage
// @Router /users/profile [get]
func (h *Handler) Profile(g *gin.Context) {
	u, _ := g.Get("user")
	usr := u.(*User)
	g.JSON(http.StatusOK, usr.AsProfile())
}

// List lists the users of the application.
// @Summary List users endpoint
// @Schemes
// @Description Lists the users of the application.
// @Accept json
// @Produce json
// @Success 200 {object} common.StatusMessage
// @Failure 400 {object} common.StatusMessage
// @Router /users [get]
func (h *Handler) List(g *gin.Context) {
	var (
		users []User
		err   error
		res   []AdminView
	)
	users, err = h.users.List()
	if err != nil {
		log.WithError(err).Error("Failed to list users")
		g.AbortWithStatusJSON(http.StatusInternalServerError, common.StatusMessage{
			Message: "Unknown error, please contact administrator!",
		})
		return
	}

	res = make([]AdminView, 0)

	for _, usr := range users {
		res = append(res, usr.AsAdminView())
	}
	g.JSON(http.StatusOK, res)
}

// Patch updates settings (e.g. role) for a user.
// @Summary User update endpoint
// @Schemes
// @Description Updates the target user
// @Accept json
// @Produce json
// @Param id path int true "ID of the user information to patch"
// @Success 200 {object} common.StatusMessage
// @Failure 400 {object} common.StatusMessage
// @Router /users/:id [patch]
func (h *Handler) Patch(g *gin.Context) {
	var (
		in  Patch
		err error
	)
	if err = g.ShouldBindJSON(&in); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Malformed user data"})
		return
	}
	if err = h.users.Patch(in); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Invalid user parameters provided!"})
		return
	}
	g.JSON(http.StatusOK, common.StatusMessage{
		Message: "User patched!",
	})
}

// SetEnabled enables/disables a user for login.
// @Summary User enable/disable endpoint
// @Schemes
// @Description Updates the target user
// @Accept json
// @Produce json
// @Param id path int true "ID of the user information to patch"
// @Param data body user.SetEnabled true "Whether the user is enabled to log in and upload photos"
// @Success 200 {object} common.StatusMessage
// @Failure 400 {object} common.StatusMessage
// @Router /users/:id/enabled [put]
func (h *Handler) SetEnabled(g *gin.Context) {
	var (
		in  SetEnabled
		id  uuid.UUID
		err error
	)
	if err = g.ShouldBindJSON(&in); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Malformed user data"})
		return
	}
	id, err = uuid.Parse(in.ID)
	if err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Invalid user ID provided!"})
		return
	}
	if err = h.users.SetEnabled(id, in.Enabled); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Invalid parameters provided!"})
		return
	}
	g.JSON(http.StatusOK, common.StatusMessage{
		Message: "User updated!",
	})
}

// Update details (firstname and lastname) for a user.
// @Summary User update endpoint
// @Schemes
// @Description Updates the target user
// @Accept json
// @Produce json
// @Param id path int true "ID of the user information to patch"
// @Param data body user.Profile true "The new version of the user information to use for update"
// @Success 200 {object} common.StatusMessage
// @Failure 400 {object} common.StatusMessage
// @Router /users/:id [put]
func (h *Handler) Update(g *gin.Context) {
	var (
		in  Profile
		err error
	)

	u, _ := g.Get("user")
	usr := u.(*User)

	if err = g.ShouldBindJSON(&in); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Malformed user data"})
		return
	}

	usr.FirstName = in.FirstName
	usr.LastName = in.LastName

	if err = h.users.Update(usr); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, common.StatusMessage{Message: "Invalid user parameters provided!"})
		return
	}
	g.JSON(http.StatusOK, common.StatusMessage{
		Message: "User patched!",
	})
}
