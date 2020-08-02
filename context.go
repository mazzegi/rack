package rack

type Context interface {
	Session() Session
}

type context struct {
	session Session
}

func newContext(s Session) *context {
	return &context{
		session: s,
	}
}

func (c *context) Session() Session {
	return c.session
}
