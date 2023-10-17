package handler

import (
	"net/http"

	"github.com/Dyleme/Notifier/internal/authorization/handler/api"
	"github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/pkg/http/requests"
	"github.com/Dyleme/Notifier/pkg/http/responses"
)

type AuthHandler struct {
	serv *service.AuthService
}

func New(serv *service.AuthService) *AuthHandler {
	return &AuthHandler{serv: serv}
}

func (ah AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginInput api.LoginBody
	err := requests.Bind(r, &loginInput)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err)

		return
	}

	accessKey, err := ah.serv.AuthUser(r.Context(), mapValidateUser(loginInput))
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	tokens := mapTokens(accessKey)
	responses.JSON(w, http.StatusOK, tokens)
}

func mapTokens(accessKey string) api.Tokens {
	return api.Tokens{
		AccessToken:  &accessKey,
		RefreshToken: nil,
	}
}

func mapValidateUser(lb api.LoginBody) service.ValidateUserInput {
	return service.ValidateUserInput{
		AuthName: lb.LoginString,
		Password: lb.Password,
	}
}

func (ah AuthHandler) Registration(w http.ResponseWriter, r *http.Request) {
	var regInput api.RegistrationBody
	err := requests.Bind(r, &regInput)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err)

		return
	}

	accessKey, err := ah.serv.CreateUser(r.Context(), mapCreateUserInput(regInput))
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	tokens := mapTokens(accessKey)
	responses.JSON(w, http.StatusOK, tokens)
}

func mapCreateUserInput(reg api.RegistrationBody) service.CreateUserInput {
	return service.CreateUserInput{
		Email:    string(reg.Email),
		Password: reg.Password,
		TGID:     nil,
	}
}
