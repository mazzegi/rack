package rack

import "strings"

type Service struct {
	prefix string
	*Environment
}

func NewService(prefix string, env *Environment) *Service {
	return &Service{
		prefix:      prefix,
		Environment: env,
	}
}

func (s *Service) Resolve(pattern string) string {
	return strings.ReplaceAll(s.prefix+pattern, "//", "/")
}
