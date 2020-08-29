package auth

type MockRepository struct {
}

func (r MockRepository) IsValidUserNameAndPassword(user, pwd string) bool {
	return true
}
