package Repository

import (
	"TP2023_DBMS_Project/internal/app/forum/Model"
)

type Repository interface {
	CreateForum(forum Model.ForumProj) (Model.ForumProj, error)
	ForumBySlug(slug string) (Model.ForumProj, error)

	CreateThread(thread Model.ThreadForum) (Model.ThreadForum, error)
	UpdateThread(newThread Model.ThreadForum) (Model.ThreadForum, error)
	GetThreadsWithParams(slug string, lim int, from string, desc bool) ([]Model.ThreadForum, error)
	IsThread(slug string) (bool, error)
	ThreadBySlug(slug string) (Model.ThreadForum, error)
	ThreadByID(id int) (Model.ThreadForum, error)
	ThreadIDBySlug(slug string) (int, error)
	ThreadSlugByID(id int) (string, error)

	CreatePosts(posts []Model.PostThread, threadID int) ([]Model.PostThread, error)
	GetPostsWithParams(postSlugOrId Model.ThreadForum, lim, from int, sort string, desc bool) ([]Model.PostThread, error)
	GetPostRelated(id int, related []string) (map[string]interface{}, error)
	UpdatePost(newPost Model.PostThread) (Model.PostThread, error)

	CreateVote(vote Model.VotePost) error
	UpdateVote(vote Model.VotePost) error

	Status() (map[string]int, error)
	Clear() error
}
