package auth

type Repository interface {
	IsValidUserNameAndPassword(user, pwd string) bool
}
