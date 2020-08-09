package rack

import (
	"net"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mazzegi/log"
	"github.com/pkg/errors"
)

type HandleFunc func(ctx Context, w http.ResponseWriter, r *http.Request)

type Option func(e *Environment) error

func WithNotAuthorizedHandler(handle HandleFunc) Option {
	return func(e *Environment) error {
		e.notAuthorizedHandler = handle
		return nil
	}
}

func WithForbiddenHandler(handle HandleFunc) Option {
	return func(e *Environment) error {
		e.forbiddenHandler = handle
		return nil
	}
}

func WithNotFoundHandler(handle HandleFunc) Option {
	return func(e *Environment) error {
		e.router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := e.ensureSession(w, r)
			handle(newContext(session), w, r)
		})
		return nil
	}
}

type Environment struct {
	listener             net.Listener
	router               *mux.Router
	sessionStore         SessionStore
	notAuthorizedHandler HandleFunc
	forbiddenHandler     HandleFunc
	noAuth               bool
}

func NewEnvironment(bind string, sessionStore SessionStore, options ...Option) (*Environment, error) {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		return nil, errors.Wrapf(err, "listen to %s", bind)
	}
	router := mux.NewRouter().StrictSlash(true)
	router.Use(handlers.CORS())

	e := &Environment{
		listener:     l,
		router:       router,
		sessionStore: sessionStore,
	}

	for _, opt := range options {
		err := opt(e)
		if err != nil {
			return nil, err
		}
	}

	//Check environment for no-auth
	if os.Getenv("RACK_NO_AUTH") == "1" {
		e.noAuth = true
	} else {
		e.noAuth = false
	}

	return e, nil
}

func (e *Environment) Close() {
	e.listener.Close()
}

func (e *Environment) Run() {
	http.Serve(e.listener, e.router)
}

func (e *Environment) ensureSession(w http.ResponseWriter, r *http.Request) Session {
	session, ok := e.sessionStore.Find(r)
	if !ok {
		session = e.sessionStore.New(w)
		log.Infof("found no session for request. create new (%s)", session.String())
	}
	return session
}

//Files
func (e *Environment) ServeFiles(prefix string, dir string) {
	e.router.PathPrefix(prefix).Handler(http.StripPrefix(prefix, http.FileServer(http.Dir(dir))))
}

//Static Handler
func (e *Environment) HandleNotAuthorized(ctx Context, w http.ResponseWriter, r *http.Request) {
	if e.notAuthorizedHandler != nil {
		e.notAuthorizedHandler(ctx, w, r)
	} else {
		http.Error(w, "not authorized", http.StatusUnauthorized)
	}
}

func (e *Environment) HandleForbidden(ctx Context, w http.ResponseWriter, r *http.Request) {
	if e.forbiddenHandler != nil {
		e.forbiddenHandler(ctx, w, r)
	} else {
		http.Error(w, "forbidden", http.StatusForbidden)
	}
}

//

func (e *Environment) HandleGET(pattern string, handle HandleFunc) {
	e.router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		session := e.ensureSession(w, r)
		handle(newContext(session), w, r)
	}).Methods("GET", "OPTIONS")
}

func (e *Environment) HandlePOST(pattern string, handle HandleFunc) {
	e.router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		session := e.ensureSession(w, r)
		handle(newContext(session), w, r)
	}).Methods("POST", "OPTIONS")
}

func (e *Environment) HandleGETAuthorized(pattern string, handle HandleFunc) {
	e.router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		session := e.ensureSession(w, r)
		if !e.noAuth && !session.IsAuthorized() {
			e.HandleNotAuthorized(newContext(session), w, r)
			return
		}
		handle(newContext(session), w, r)
	}).Methods("GET", "OPTIONS")
}

func (e *Environment) HandlePOSTAuthorized(pattern string, handle HandleFunc) {
	e.router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		session := e.ensureSession(w, r)
		if !e.noAuth && !session.IsAuthorized() {
			e.HandleNotAuthorized(newContext(session), w, r)
			return
		}
		handle(newContext(session), w, r)
	}).Methods("POST", "OPTIONS")
}
