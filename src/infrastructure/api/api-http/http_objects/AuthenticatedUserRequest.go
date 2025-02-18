package http_objects

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api/crypto_cloud_billing_errors"
	"goproxy/infrastructure/api/api-http/google_auth"
	"net/http"
)

type AuthenticatedUserBillingRequest struct {
	request *http.Request
}

func NewAuthenticatedUserBillingRequest(r *http.Request) AuthenticatedUserBillingRequest {
	return AuthenticatedUserBillingRequest{
		request: r,
	}
}

func (a *AuthenticatedUserBillingRequest) UserEmail() (string, error) {
	token, tokenErr := a.extractToken(a.request)
	if tokenErr != nil {
		return "", tokenErr
	}

	verifiedToken, verifiedTokenErr := a.verifyToken(token)
	if verifiedTokenErr != nil {
		return "", verifiedTokenErr
	}

	email, emailErr := a.extractEmail(verifiedToken)
	if emailErr != nil {
		return "", emailErr
	}

	return email, nil
}

func (a *AuthenticatedUserBillingRequest) extractToken(r *http.Request) (string, error) {
	idToken, err := google_auth.GetIdTokenFromCookie(r)
	if err != nil {
		return "", crypto_cloud_billing_errors.NewTokenExtractionErr(err)
	}

	return idToken, nil
}

func (a *AuthenticatedUserBillingRequest) verifyToken(token string) (*jwt.Token, error) {
	verifiedToken, err := google_auth.VerifyIDToken(token)
	if err != nil {
		return verifiedToken, crypto_cloud_billing_errors.
			NewTokenVerificationErr(err)
	}

	return verifiedToken, nil
}

func (a *AuthenticatedUserBillingRequest) extractEmail(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", crypto_cloud_billing_errors.
			NewTokenEmailExtractionErr(fmt.Errorf("could not extract email from claims"))
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return "", crypto_cloud_billing_errors.
			NewTokenEmailExtractionErr(fmt.Errorf("key 'email' was not found in claims"))
	}

	return email, nil
}
