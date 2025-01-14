package history

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/inokone/go-micro-saas/internal/auth/user"
	"github.com/inokone/go-micro-saas/internal/common"
)

const historySize = 25

// Handler is a struct for web handles related to user history.
type Handler struct {
	history Storer
}

// NewHandler creates a new `Handler`, based on the user history persistence.
func NewHandler(history Storer) *Handler {
	return &Handler{
		history: history,
	}
}

// List is a method of `Handler`. Lists all history events for a user.
// @Summary List history events endpoint
// @Schemes
// @Description Lists all history events for a user
// @Accept json
// @Produce json
// @Success 200 {array} common.Event
// @Failure 400 {object} common.StatusMessage
// @Failure 404 {object} common.StatusMessage
// @Failure 500 {object} common.StatusMessage
// @Router /users/:id/history [get]
func (h *Handler) List(g *gin.Context) {
	usr, err := currentUser(g)
	if err != nil {
		g.AbortWithStatusJSON(http.StatusUnauthorized, common.StatusMessage{Message: "Not authorized!"})
		return
	}

	events, err := h.history.List(usr.ID, historySize)
	if err != nil {
		log.WithError(err).Error("Could not get history events, unknown error")
		g.AbortWithStatusJSON(http.StatusNotFound, common.StatusMessage{Message: "User history not found!"})
		return
	}

	g.JSON(http.StatusOK, events)
}

func currentUser(g *gin.Context) (*user.User, error) {
	u, ok := g.Get("user")
	if !ok {
		return nil, errors.New("user could not be extracted from session")
	}
	usr := u.(*user.User)
	return usr, nil
}
