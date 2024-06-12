package usersrv

import "github.com/olexsmir/onasty/internal/store/psql/userepo"

type UserServicer interface {
	SignUp() error
}

type UserSrv struct {
	store userepo.UserStorer
}

func New(store userepo.UserStorer) UserServicer {
	return &UserSrv{
		store: store,
	}
}

// type SignUp
func (s *UserSrv) SignUp() error {
	return nil
}
