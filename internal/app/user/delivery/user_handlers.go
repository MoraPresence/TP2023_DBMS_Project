package delivery

import (
	"TP2023_DBMS_Project/internal/app/forum/Repository"
	"TP2023_DBMS_Project/internal/app/user"
	"TP2023_DBMS_Project/internal/app/user/models"
	"TP2023_DBMS_Project/internal/pkg/GetValue"
	"TP2023_DBMS_Project/internal/pkg/resp"
	"encoding/json"
	"fmt"
	"github.com/fasthttp/router"
	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
)

type userHandler struct {
	userRepo  user.Repository
	forumRepo Repository.Repository
}

func New(r *router.Router, ur user.Repository, fr Repository.Repository) {
	handler := userHandler{
		userRepo:  ur,
		forumRepo: fr,
	}

	r.POST("/api/user/{nick}/create", handler.AddUser)
	r.GET("/api/user/{nick}/profile", handler.GetUser)
	r.POST("/api/user/{nick}/profile", handler.UpdateUser)

	r.GET("/api/forum/{slug}/users", handler.GetByForumSlug)
}

func (ur *userHandler) AddUser(ctx *fasthttp.RequestCtx) {
	nick, ok := ctx.UserValue("nick").(string)
	if !ok {
		resp.Send(400, "((((", ctx)
		return
	}

	added := models.UserInfo{Nick: nick}

	err := json.Unmarshal(ctx.PostBody(), &added)
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	err = ur.userRepo.AddUser(added)

	if err != nil {
		users, err := ur.userRepo.GetUserByNickAndEmail(added.Nick, added.Email)
		if err != nil {
			resp.SendUnexpError(err.Error(), ctx)
		}
		resp.Send(409, users, ctx)
		return
	}

	resp.Send(201, added, ctx)
	return
}

func (ur *userHandler) GetUser(ctx *fasthttp.RequestCtx) {
	nick, found := ctx.UserValue("nick").(string)
	if !found {
		resp.Send(400, "((((", ctx)
		return
	}

	user, err := ur.userRepo.GetUserByNick(nick)
	if err != nil {
		err := resp.HttpErr{
			Message: fmt.Sprintf("Can't find user by nick: %s", nick),
		}
		resp.Send(404, err, ctx)
		return
	}

	resp.SendOk(user, ctx)
	return
}

func (ur *userHandler) UpdateUser(ctx *fasthttp.RequestCtx) {
	nick, found := ctx.UserValue("nick").(string)
	if !found {
		resp.Send(400, "((((", ctx)
		return
	}

	updated := models.UserInfo{Nick: nick}

	err := json.Unmarshal(ctx.PostBody(), &updated)
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	userDB, err := ur.userRepo.UpdateUser(updated)
	if pgerr, ok := err.(pgx.PgError); ok {
		switch pgerr.Code {
		case "23505":
			err := resp.HttpErr{
				Message: fmt.Sprintf("This email is already registered by user: %s", updated.Email),
			}
			resp.Send(409, err, ctx)
			return
		}
	}
	if err != nil {
		err := resp.HttpErr{
			Message: fmt.Sprintf("Can't find user by nick: %s", updated.Nick),
		}
		resp.Send(404, err, ctx)
		return
	}

	resp.SendOk(userDB, ctx)
	return
}

func (ur *userHandler) GetByForumSlug(ctx *fasthttp.RequestCtx) {
	slug, found := ctx.UserValue("slug").(string)
	if !found {
		resp.Send(400, "((((", ctx)
		return
	}

	lim, err := GetValue.GetInt(ctx, "limit")
	if err != nil {
		resp.Send(400, err, ctx)
		return
	}

	from := string(ctx.QueryArgs().Peek("since"))

	desc, err := GetValue.GetBool(ctx, "desc")
	if err != nil {
		resp.Send(400, err, ctx)
		return
	}

	users, err := ur.userRepo.GetUsersByForumSlug(slug, lim, from, desc)
	if err != nil {
		resp.Send(404, err, ctx)
		return
	}

	if users == nil {
		_, err = ur.forumRepo.ForumBySlug(slug)
		if err != nil {
			resp.Send(404, err, ctx)
			return
		}
		resp.SendOk([]models.UserInfo{}, ctx)
		return
	}

	resp.SendOk(users, ctx)
	return
}
