package delivery

import (
	"TP2023_DBMS_Project/internal/app/forum/Model"
	"TP2023_DBMS_Project/internal/app/forum/Repository"
	"TP2023_DBMS_Project/internal/app/user"
	"TP2023_DBMS_Project/internal/pkg/GetValue"
	"TP2023_DBMS_Project/internal/pkg/resp"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/fasthttp/router"
	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
)

type forumHandler struct {
	forumRepo Repository.Repository
	userRepo  user.Repository
}

func New(r *router.Router, fr Repository.Repository, ur user.Repository) {
	handler := forumHandler{forumRepo: fr, userRepo: ur}

	r.POST("/api/forum/create", handler.CreateForum)
	r.GET("/api/forum/{slug}/details", handler.GetForum)
	r.POST("/api/forum/{slug}/create", handler.CreateThread)

	r.GET("/api/forum/{slug}/threads", handler.GetThreads)

	r.GET("/api/thread/{slug_or_id}/details", handler.GetThreadBySlug)

	r.POST("/api/thread/{slug_or_id}/details", handler.UpdateThreadBySlugOrID)

	r.POST("/api/thread/{slug_or_id}/create", handler.CreatePostBySlug)
	r.GET("/api/thread/{slug_or_id}/posts", handler.GetPostBySlug)

	r.GET("/api/post/{id:[0-9]+}/details", handler.GetPostByID)
	r.POST("/api/post/{id:[0-9]+}/details", handler.UpdatePost)

	r.POST("/api/thread/{id:[0-9]+}/vote", handler.CreateVoteByID)
	r.POST("/api/thread/{slug}/vote", handler.CreateVoteBySlug)

	r.GET("/api/service/status", handler.Stat)
	r.POST("/api/service/clear", handler.Clear)
}

func (f *forumHandler) CreateForum(ctx *fasthttp.RequestCtx) {
	var forumGot Model.ForumProj
	err := json.Unmarshal(ctx.PostBody(), &forumGot)
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	forumUpdated, err := f.forumRepo.CreateForum(forumGot)
	pgerr, ok := err.(pgx.PgError)
	if ok {
		switch pgerr.Code {
		case "23505":
			f, err := f.forumRepo.ForumBySlug(forumGot.Slug)
			if err != nil {
				resp.SendUnexpError(err.Error(), ctx)
				return
			}
			resp.Send(409, f, ctx)
			return
		case "23503":
			err := resp.HttpErr{
				Message: fmt.Sprintf("Can't find user with nickname: %s", forumGot.User),
			}
			resp.Send(404, err, ctx)
			return
		}

	}
	if err == pgx.ErrNoRows {
		err := resp.HttpErr{
			Message: fmt.Sprintf("Can't find user with nickname: %s", forumGot.User),
		}
		resp.Send(404, err, ctx)
		return
	}

	if err != nil {
		resp.Send(400, err.Error(), ctx)
		return
	}

	resp.Send(201, forumUpdated, ctx)
}

func (f *forumHandler) GetForum(ctx *fasthttp.RequestCtx) {
	slug, ok := ctx.UserValue("slug").(string)
	if !ok {
		resp.Send(400, "((((", ctx)
		return
	}

	forumGot, err := f.forumRepo.ForumBySlug(slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			err := resp.HttpErr{
				Message: fmt.Sprintf("Can't find forum with slug: %s", slug),
			}
			resp.Send(404, err, ctx)
			return
		}
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	resp.SendOk(forumGot, ctx)
	return
}

func (f *forumHandler) CreateThread(ctx *fasthttp.RequestCtx) {
	forumGot, got := ctx.UserValue("slug").(string)
	if !got {
		resp.Send(400, "((((", ctx)
		return
	}

	threadGot := Model.ThreadForum{Forum: forumGot}

	err := json.Unmarshal(ctx.PostBody(), &threadGot)
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	threadNew, err := f.forumRepo.CreateThread(threadGot)
	pgerr, ok := err.(pgx.PgError)
	if ok && pgerr.Code == "23505" {
		threadOld, err := f.forumRepo.ThreadBySlug(threadGot.Slug.String)
		if err != nil {
			resp.SendUnexpError(err.Error(), ctx)
			return
		}
		resp.Send(409, threadOld, ctx)
		return
	}

	if err != nil {
		respBad := resp.HttpErr{Message: err.Error()}
		resp.Send(404, respBad, ctx)
		return
	}
	resp.Send(201, threadNew, ctx)
}

func (f *forumHandler) GetThreads(ctx *fasthttp.RequestCtx) {
	forumGot, ok := ctx.UserValue("slug").(string)
	if !ok {
		resp.Send(400, "((((", ctx)
		return
	}

	lim, err := GetValue.GetInt(ctx, "limit")
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	from := string(ctx.QueryArgs().Peek("since"))

	desc, err := GetValue.GetBool(ctx, "desc")
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}
	threadsGot, err := f.forumRepo.GetThreadsWithParams(forumGot, lim, from, desc)
	if err == pgx.ErrNoRows || len(threadsGot) == 0 {
		exist, err := f.forumRepo.IsThread(forumGot)
		if err != nil {
			resp.SendUnexpError(err.Error(), ctx)
			return
		}
		if exist {
			respBad := make([]Model.ThreadForum, 0, 0)
			resp.SendOk(respBad, ctx)
			return
		}
		respBad := resp.HttpErr{
			Message: fmt.Sprintf("Can't find forum by slug: %s", forumGot),
		}
		resp.Send(404, respBad, ctx)
		return
	}

	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}
	resp.SendOk(threadsGot, ctx)
	return
}

func (f *forumHandler) createPost(ctx *fasthttp.RequestCtx, id int) {
	var posts []Model.PostThread
	err := json.Unmarshal(ctx.PostBody(), &posts)
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}
	if len(posts) == 0 {
		resp.Send(201, posts, ctx)
		return
	}
	creator := posts[0].Author
	posts, err = f.forumRepo.CreatePosts(posts, id)
	if len(posts) == 0 {
		err = pgx.ErrNoRows
	}
	if err != nil {
		if pgerr, ok := err.(pgx.PgError); ok {
			switch pgerr.Code {
			case "00409":
				resp.Send(409, map[int]int{}, ctx)
				return
			}
		}

		if err == pgx.ErrNoRows {
			_, err = f.forumRepo.ThreadByID(id)
			if err == pgx.ErrNoRows {
				resp.Send(404, map[int]int{}, ctx)
				return
			}
			_, err = f.userRepo.GetUserByNick(creator)
			if err == pgx.ErrNoRows {
				resp.Send(404, map[int]int{}, ctx)
				return
			}
			resp.Send(409, map[int]int{}, ctx)
			return
		}

		respBad := map[string]string{
			"message": err.Error(),
		}
		resp.Send(404, respBad, ctx)
		return
	}

	resp.Send(201, posts, ctx)
	return
}

func (f *forumHandler) CreatePostBySlug(ctx *fasthttp.RequestCtx) {
	smth, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		resp.Send(400, "((((", ctx)
		return
	}
	var id int
	id, err := strconv.Atoi(smth)
	if err == nil {
		_, err = f.forumRepo.ThreadByID(id)
		if err != nil {
			pespBad := resp.HttpErr{
				Message: fmt.Sprintf(err.Error()),
			}
			resp.Send(404, pespBad, ctx)

			return
		}
	} else {
		id, err = f.forumRepo.ThreadIDBySlug(smth)
		if err != nil {
			pespBad := resp.HttpErr{
				Message: fmt.Sprintf(err.Error()),
			}
			resp.Send(404, pespBad, ctx)
			return
		}
	}

	f.createPost(ctx, id)

}

func (f *forumHandler) CreateVoteBySlug(ctx *fasthttp.RequestCtx) {
	slug, ok := ctx.UserValue("slug").(string)
	if !ok {
		resp.Send(400, "((((", ctx)
		return
	}

	var voteNew Model.VotePost
	err := json.Unmarshal(ctx.PostBody(), &voteNew)
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}
	threadID, _ := f.forumRepo.ThreadIDBySlug(slug)
	voteNew.IdThread = int64(threadID)
	err = f.forumRepo.CreateVote(voteNew)
	if err != nil {
		pgerr, ok := err.(pgx.PgError)
		if !ok {
			pespBad := resp.HttpErr{
				Message: fmt.Sprintf(err.Error()),
			}
			resp.Send(404, pespBad, ctx)
			return
		}
		if pgerr.Code != "23505" {
			pespBad := resp.HttpErr{
				Message: fmt.Sprintf(err.Error()),
			}
			resp.Send(404, pespBad, ctx)
			return
		}

	}

	threadGot, err := f.forumRepo.ThreadBySlug(slug)
	if err != nil {
		pespBad := resp.HttpErr{
			Message: fmt.Sprintf(err.Error()),
		}
		resp.Send(404, pespBad, ctx)
		return
	}

	resp.SendOk(threadGot, ctx)
}

func (f *forumHandler) CreateVoteByID(ctx *fasthttp.RequestCtx) {
	str, ok := ctx.UserValue("id").(string)
	if !ok {
		resp.Send(400, "((((", ctx)
		return
	}

	value, err := strconv.Atoi(str)
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	var voteNew Model.VotePost
	err = json.Unmarshal(ctx.PostBody(), &voteNew)
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}
	voteNew.IdThread = int64(value)

	err = f.forumRepo.CreateVote(voteNew)
	if err != nil {
		pgerr, ok := err.(pgx.PgError)
		if !ok {
			pespBad := resp.HttpErr{
				Message: fmt.Sprintf(err.Error()),
			}
			resp.Send(404, pespBad, ctx)
			return
		}
		if pgerr.Code != "23505" {
			pespBad := resp.HttpErr{
				Message: fmt.Sprintf(err.Error()),
			}
			resp.Send(404, pespBad, ctx)
			return
		} else {
			err = f.forumRepo.UpdateVote(voteNew)
			if err != nil {
				pespBad := resp.HttpErr{
					Message: fmt.Sprintf(err.Error()),
				}
				resp.Send(404, pespBad, ctx)
				return
			}
		}
	}
	threadGot, err := f.forumRepo.ThreadByID(value)
	if err != nil {
		pespBad := resp.HttpErr{
			Message: fmt.Sprintf(err.Error()),
		}
		resp.Send(404, pespBad, ctx)
		return
	}

	resp.SendOk(threadGot, ctx)
}

func (f *forumHandler) GetThreadBySlug(ctx *fasthttp.RequestCtx) {
	slug, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		resp.Send(400, "((((", ctx)
		return
	}

	id, err := strconv.Atoi(slug)
	if err != nil {
		id, err = f.forumRepo.ThreadIDBySlug(slug)
		if err != nil {
			pespBad := resp.HttpErr{
				Message: fmt.Sprintf(err.Error()),
			}
			resp.Send(404, pespBad, ctx)
			return
		}
	}

	forumGot, err := f.forumRepo.ThreadByID(id)
	if err != nil {
		pespBad := resp.HttpErr{
			Message: fmt.Sprintf(err.Error()),
		}
		resp.Send(404, pespBad, ctx)
		return
	}

	resp.SendOk(forumGot, ctx)
	return
}

func (f *forumHandler) UpdateThreadBySlugOrID(ctx *fasthttp.RequestCtx) {
	smth, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		resp.Send(400, "((((", ctx)
		return
	}
	var threadGot Model.ThreadForum
	if id, err := strconv.Atoi(smth); err == nil {
		threadGot.Id = int32(id)
	} else {
		threadGot.Slug = Model.NullString{
			NullString: sql.NullString{Valid: true, String: smth},
		}
	}

	err := json.Unmarshal(ctx.PostBody(), &threadGot)
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	threadNew, err := f.forumRepo.UpdateThread(threadGot)
	if err != nil {
		resp.Send(404, err, ctx)
		return
	}

	resp.SendOk(threadNew, ctx)
	return
}

func (f *forumHandler) GetPostBySlug(ctx *fasthttp.RequestCtx) {
	smth, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		resp.Send(400, "((((", ctx)
		return
	}

	lim, err := GetValue.GetInt(ctx, "limit")
	if lim == 4 {
		fmt.Println("hui")
	}
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	from, err := GetValue.GetInt(ctx, "since")
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	sortType := string(ctx.QueryArgs().Peek("sort"))
	if sortType == "" {
		sortType = "flat"
	}

	desc, err := GetValue.GetBool(ctx, "desc")
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}
	var threadGot Model.ThreadForum
	if id, err := strconv.Atoi(smth); err == nil {
		threadGot.Id = int32(id)
	}
	slug := sql.NullString{String: smth, Valid: true}
	slugJs := Model.NullString{NullString: slug}
	threadGot.Slug = slugJs

	postsGot, err := f.forumRepo.GetPostsWithParams(threadGot, lim, from, sortType, desc)
	if err != nil {
		pespBad := resp.HttpErr{
			Message: fmt.Sprintf(err.Error()),
		}
		resp.Send(404, pespBad, ctx)
		return
	}

	if postsGot == nil {
		if threadGot.Id != 0 {
			_, err := f.forumRepo.ThreadByID(int(threadGot.Id))
			if err == pgx.ErrNoRows {
				respBad := resp.HttpErr{Message: err.Error()}
				resp.Send(404, respBad, ctx)
				return
			}
		}
		resp.SendOk([]int{}, ctx)
		return
	}

	resp.SendOk(postsGot, ctx)
	return
}

func (f *forumHandler) GetPostByID(ctx *fasthttp.RequestCtx) {
	str, ok := ctx.UserValue("id").(string)
	if !ok {
		resp.Send(400, "((((", ctx)
		return
	}

	id, err := strconv.Atoi(str)
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	related := string(ctx.QueryArgs().Peek("related"))

	postGot, err := f.forumRepo.GetPostRelated(id, strings.Split(related, ","))
	if err != nil {
		respBad := resp.HttpErr{Message: err.Error()}
		resp.Send(404, respBad, ctx)
		return
	}

	resp.SendOk(postGot, ctx)
	return
}

func (f *forumHandler) UpdatePost(ctx *fasthttp.RequestCtx) {
	str, ok := ctx.UserValue("id").(string)
	if !ok {
		resp.Send(400, "((((", ctx)
		return
	}

	id, err := strconv.Atoi(str)
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	postGot := Model.PostThread{
		Id: int64(id),
	}

	err = json.Unmarshal(ctx.PostBody(), &postGot)
	if err != nil {
		resp.SendUnexpError(err.Error(), ctx)
		return
	}

	postGot, err = f.forumRepo.UpdatePost(postGot)
	if err != nil {
		respBad := resp.HttpErr{Message: err.Error()}
		resp.Send(404, respBad, ctx)
		return
	}

	resp.SendOk(postGot, ctx)
	return
}

func (f *forumHandler) Stat(ctx *fasthttp.RequestCtx) {
	stat, err := f.forumRepo.Status()
	if err != nil {
		resp.Send(404, err.Error(), ctx)
		return
	}
	resp.SendOk(stat, ctx)
	return
}

func (f *forumHandler) Clear(ctx *fasthttp.RequestCtx) {
	err := f.forumRepo.Clear()
	if err != nil {
		resp.Send(404, err.Error(), ctx)
		return
	}
	resp.SendOk("", ctx)
	return
}
