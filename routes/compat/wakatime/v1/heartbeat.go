package v1

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
)

type HeartbeatHandler struct {
	userSrvc      services.IUserService
	heartbeatSrvc services.IHeartbeatService
}

func NewHeartbeatHandler(userService services.IUserService, heartbeatService services.IHeartbeatService) *HeartbeatHandler {
	return &HeartbeatHandler{
		userSrvc:      userService,
		heartbeatSrvc: heartbeatService,
	}
}

func (h *HeartbeatHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler,
	)
	r.Path("/compat/wakatime/v1/users/{user}/heartbeats").Methods(http.MethodGet).HandlerFunc(h.Get)
}

// @Summary Get heartbeats of user for specified date
// @ID get-heartbeats
// @Tags heartbeat
// @Param date query string true "Date"
// @Param user path string true "Username (or current)"
// @Security ApiKeyAuth
// @Success 200
// @Router /compat/wakatime/v1/users/{user}/heartbeats [get]
func (h *HeartbeatHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := routeutils.CheckEffectiveUser(w, r, h.userSrvc, "current")
	if err != nil {
		return // response was already sent by util function
	}

	params := r.URL.Query()
	dateParam := params.Get("date")
	date, err := time.Parse(conf.SimpleDateFormat, dateParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad date"))
		return
	}
	timezone := user.TZ()
	rangeFrom, rangeTo := utils.StartOfDay(date).In(timezone), utils.EndOfDay(date).In(timezone)

	heartbeats, err := h.heartbeatSrvc.GetAllWithin(rangeFrom, rangeTo, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(conf.ErrInternalServerError))
		conf.Log().Request(r).Error("failed to retrieve heartbeats - %v", err)
		return
	}
	utils.RespondJSON(w, r, http.StatusOK, heartbeats)
}
