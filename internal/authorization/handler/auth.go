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

func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body api.LoginJSONBody
	err := requests.Bind(r, &body)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err)

		return
	}

	accessKey, err := ah.serv.AuthUser(r.Context(),
		service.ValidateUserInput{
			AuthName: body.LoginString,
			Password: body.Password,
		},
	)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	tokens := api.Tokens{
		AccessToken:  &accessKey,
		RefreshToken: nil,
	}
	responses.JSON(w, http.StatusOK, tokens)
}

func (ah *AuthHandler) StartBindingToTG(w http.ResponseWriter, r *http.Request) {
	var body api.StartBindingToTGJSONBody
	err := requests.Bind(r, &body)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err)

		return
	}

	err = ah.serv.StartUserBinding(r.Context(), service.StartUserBindingInput{
		TGNickname: body.TgNickname,
		Password:   body.Password,
	})
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.Status(w, http.StatusOK)
}

func (ah *AuthHandler) BindToTG(w http.ResponseWriter, r *http.Request) {
	var body api.BindToTGJSONBody
	err := requests.Bind(r, &body)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err)

		return
	}

	err = ah.serv.BindUser(r.Context(), service.BindUserInput{
		Code:       body.Code,
		TGNickname: body.TgNickname,
	})
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.Status(w, http.StatusOK)
}
