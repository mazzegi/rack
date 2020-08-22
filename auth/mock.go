package auth

type MockRepositoty struct {
}

func (r MockRepositoty) IsValidUserNameAndPassword(user, pwd string) bool {
	return true
}
