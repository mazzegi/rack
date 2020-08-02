package rack

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mazzegi/uuidv4"
)

type Session interface {
	IsAuthorized() bool
	Authorize(user string)
	Unauthorize()
	User() string
	String() string
}

type SessionStore interface {
	Find(r *http.Request) (Session, bool)
	New(w http.ResponseWriter) Session
}

type session struct {
	sync.RWMutex
	id         string
	expiresOn  time.Time
	authorized bool
	user       string
}

func (s *session) IsAuthorized() bool {
	s.RLock()
	defer s.RUnlock()
	return s.authorized
}

func (s *session) Authorize(user string) {
	s.Lock()
	defer s.Unlock()
	s.authorized = true
	s.user = user
}

func (s *session) Unauthorize() {
	s.Lock()
	defer s.Unlock()
	s.authorized = false
}

func (s *session) User() string {
	return s.user
}

func (s *session) ID() string {
	return s.id
}

func (s *session) String() string {
	s.RLock()
	defer s.RUnlock()
	return fmt.Sprintf("id:(%s) logged-on:(%t) as (%s) expires-on:(%s)", s.id, s.authorized, s.user, s.expiresOn.Format(time.RFC3339))
}

//InMemorySessionStore
type InMemorySessionStore struct {
	sync.RWMutex
	cookiesName     string
	cookiesExpireIn time.Duration
	cookiesPath     string
	sessions        map[string]*session
}

func NewInMemorySessionStore(cookieName string, cookiesExpireIn time.Duration, cookiesPath string) *InMemorySessionStore {
	return &InMemorySessionStore{
		cookiesName:     cookieName,
		cookiesExpireIn: cookiesExpireIn,
		cookiesPath:     cookiesPath,
		sessions:        map[string]*session{},
	}
}

func (store *InMemorySessionStore) Find(r *http.Request) (Session, bool) {
	store.RLock()
	defer store.RUnlock()
	co, err := r.Cookie(store.cookiesName)
	if err != nil {
		return nil, false
	}
	s, contains := store.sessions[co.Value]
	return s, contains
}

func (store *InMemorySessionStore) New(w http.ResponseWriter) Session {
	store.Lock()
	defer store.Unlock()
	s := &session{
		id:         uuidv4.MustMake(),
		expiresOn:  time.Now().UTC().Add(store.cookiesExpireIn),
		authorized: false,
	}
	http.SetCookie(w, &http.Cookie{
		Name:    store.cookiesName,
		Value:   s.id,
		Expires: s.expiresOn,
		Path:    store.cookiesPath,
	})
	store.sessions[s.id] = s
	return s
}
