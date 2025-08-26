package authsrv

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/jwtutil"
)

func (a *AuthSrv) ParseJWTToken(token string) (jwtutil.Payload, error) {
	return a.jwtTokenizer.Parse(token)
}

func (a AuthSrv) issueTokens(ctx context.Context, userID uuid.UUID) (dtos.Tokens, error) {
	toks, err := a.createTokens(userID)
	if err != nil {
		return dtos.Tokens{}, err
	}

	if err := a.sessionstore.Set(ctx, userID, toks.Refresh, time.Now().Add(a.refreshTokenTTL)); err != nil {
		return dtos.Tokens{}, err
	}

	return toks, nil
}

func (a AuthSrv) createTokens(userID uuid.UUID) (dtos.Tokens, error) {
	accessToken, err := a.jwtTokenizer.AccessToken(jwtutil.Payload{UserID: userID.String()})
	if err != nil {
		return dtos.Tokens{}, err
	}

	refreshToken, err := a.jwtTokenizer.RefreshToken()
	if err != nil {
		return dtos.Tokens{}, err
	}

	return dtos.Tokens{
		Access:  accessToken,
		Refresh: refreshToken,
	}, err
}
