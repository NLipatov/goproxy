package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api/crypto_cloud_billing_errors"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api/dto"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api/services"
	"goproxy/infrastructure/api/api-http/google_auth"
	dto2 "goproxy/infrastructure/dto"
	"net/http"
)

type CreateInvoiceHandler struct {
	billingService        services.BillingService
	authenticationService google_auth.GoogleAuthService
}

func NewCreateInvoiceHandler(billingService services.BillingService) CreateInvoiceHandler {
	return CreateInvoiceHandler{
		billingService: billingService,
	}
}

func (ih *CreateInvoiceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	email, emailErr := ih.getUserEmailFromCookieToken(r)
	if emailErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = ih.writeJSON(w, dto2.ApiResponse[dto.CreateInvoiceResponse]{
			Payload:      nil,
			ErrorCode:    401,
			ErrorMessage: "invalid token",
		})

		return
	}

	var requestDto dto.IssueInvoiceCommandDto
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&requestDto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body"))
		return
	}

	requestDto.Email = email

	issueInvoiceResult, issueInvoiceResultErr := ih.billingService.IssueInvoice(requestDto)
	if issueInvoiceResultErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = ih.writeJSON(w, dto2.ApiResponse[dto.CreateInvoiceResponse]{
			Payload:      nil,
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: issueInvoiceResultErr.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusOK)
	_ = ih.writeJSON(w, dto2.ApiResponse[dto.CreateInvoiceResponse]{
		Payload: &dto.CreateInvoiceResponse{
			PaymentLink: issueInvoiceResult.PaymentLinq,
		},
		ErrorCode:    0,
		ErrorMessage: "",
	})
}

func (ih *CreateInvoiceHandler) getUserEmailFromCookieToken(r *http.Request) (string, error) {
	token, tokenErr := ih.extractToken(r)
	if tokenErr != nil {
		return "", tokenErr
	}

	verifiedToken, verifiedTokenErr := ih.verifyToken(token)
	if verifiedTokenErr != nil {
		return "", verifiedTokenErr
	}

	email, emailErr := ih.extractEmail(verifiedToken)
	if emailErr != nil {
		return "", emailErr
	}

	return email, nil
}

func (ih *CreateInvoiceHandler) extractToken(r *http.Request) (string, error) {
	idToken, err := google_auth.GetIdTokenFromCookie(r)
	if err != nil {
		return "", crypto_cloud_billing_errors.NewTokenExtractionErr(err)
	}

	return idToken, nil
}

func (ih *CreateInvoiceHandler) verifyToken(token string) (*jwt.Token, error) {
	verifiedToken, err := google_auth.VerifyIDToken(token)
	if err != nil {
		return verifiedToken, crypto_cloud_billing_errors.
			NewTokenVerificationErr(err)
	}

	return verifiedToken, nil
}

func (ih *CreateInvoiceHandler) extractEmail(token *jwt.Token) (string, error) {
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

func (ih *CreateInvoiceHandler) writeJSON(w http.ResponseWriter, response dto2.ApiResponse[dto.CreateInvoiceResponse]) error {
	responseBytes, responseBytesErr := json.Marshal(response)
	if responseBytesErr != nil {
		return responseBytesErr
	}

	_, err := w.Write(responseBytes)
	if err != nil {
		return err
	}

	return nil
}
