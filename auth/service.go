package auth

import (
	"net/http"

	"github.com/mazzegi/log"
	"github.com/mazzegi/rack"
)

type Service struct {
	*rack.Service
	repo Repository
}

func NewService(svc *rack.Service, r Repository) *Service {
	s := &Service{
		Service: svc,
		repo:    r,
	}
	return s
}

func (s *Service) Activate() {
	s.HandlePOST(s.Resolve("login"), s.handlePOSTLogin)
	s.HandlePOSTAuthorized(s.Resolve("logout"), s.handlePOSTLogout)
}

func (s *Service) Deactivate() {

}

func (s *Service) handlePOSTLogin(ctx rack.Context, w http.ResponseWriter, r *http.Request) {
	user, pwd := r.URL.Query().Get("user"), r.URL.Query().Get("password")
	if !s.repo.IsValidUserNameAndPassword(user, pwd) {
		log.Infof("auth: login failed (%s|%s): %s", user, pwd, ctx.Session())
		s.HandleNotAuthorized(ctx, w, r)
		return
	}
	ctx.Session().Authorize(user)
	log.Infof("auth: login sucess: %s", ctx.Session())
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Service) handlePOSTLogout(ctx rack.Context, w http.ResponseWriter, r *http.Request) {
	ctx.Session().Unauthorize()
	log.Infof("auth: logout: %s", ctx.Session())
	http.Redirect(w, r, s.Resolve("/login"), http.StatusFound)
}
