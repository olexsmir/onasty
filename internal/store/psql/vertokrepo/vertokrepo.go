package vertokrepo

import "github.com/olexsmir/onasty/internal/store/psqlutil"

type VerificationTokenStorer interface{}

var _ VerificationTokenStorer = (*VerificationTokenRepo)(nil)

type VerificationTokenRepo struct {
	db *psqlutil.DB
}

func New(db *psqlutil.DB) *VerificationTokenRepo {
	return &VerificationTokenRepo{
		db: db,
	}
}
