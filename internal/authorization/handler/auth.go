package handler

import (
	"net/http"

	"github.com/Dyleme/Notifier/internal/authorization/handler/api"
	"github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/service/handler/request"
	"github.com/Dyleme/Notifier/internal/service/handler/response"
)

type AuthHandler struct {
	serv *service.AuthService
}

func New(serv *service.AuthService) *AuthHandler {
	return &AuthHandler{serv: serv}
}

func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body api.LoginJSONBody
	err := request.Bind(r, &body)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	accessKey, err := ah.serv.AuthUser(r.Context(),
		service.ValidateUserInput{
			AuthName: body.LoginString,
			Password: body.Password,
		},
	)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	tokens := api.Tokens{
		AccessToken:  &accessKey,
		RefreshToken: nil,
	}

	response.JSON(ctx, w, http.StatusOK, tokens)
}

func (ah *AuthHandler) StartBindingToTG(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body api.StartBindingToTGJSONBody
	err := request.Bind(r, &body)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	err = ah.serv.StartUserBinding(r.Context(), service.StartUserBindingInput{
		TGNickname: body.TgNickname,
		Password:   body.Password,
	})
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	response.Status(w, http.StatusOK)
}

func (ah *AuthHandler) BindToTG(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body api.BindToTGJSONBody
	err := request.Bind(r, &body)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	err = ah.serv.BindUser(r.Context(), service.BindUserInput{
		Code:       body.Code,
		TGNickname: body.TgNickname,
	})
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	response.Status(w, http.StatusOK)
}
