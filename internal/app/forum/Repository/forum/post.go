package Postgres

import (
	f "TP2023_DBMS_Project/internal/app/forum/Model"
	"errors"
	"fmt"
	"github.com/go-openapi/strfmt"
	"strings"
	"time"
)

func (p *RepStruct) getPostsFlat(tid, lim, from int, desc bool) ([]f.PostThread, error) {

	query := `SELECT * FROM post WHERE thread=$1 `

	if desc {
		if from > 0 {
			query += fmt.Sprintf("AND id < %d ", from)
		}
		query += `ORDER BY id DESC `
	} else {
		if from > 0 {
			query += fmt.Sprintf("AND id > %d ", from)
		}
		query += `ORDER BY id `
	}
	query += `LIMIT NULLIF($2, 0)`
	var posts []f.PostThread

	row, err := p.Connection.Query(query, tid, lim)

	if err != nil {
		return posts, err
	}
	defer func() {
		if row != nil {
			row.Close()
		}
	}()

	for row.Next() {
		var post f.PostThread
		var created time.Time

		err = row.Scan(&post.Author, &created, &post.Forum, &post.Id, &post.IsEdited, &post.Message,
			&post.Parent, &post.Thread, &post.Path)

		if err != nil {
			return posts, err
		}
		post.Created = strfmt.DateTime(created.UTC()).String()
		posts = append(posts, post)

	}
	return posts, err
}

func (p *RepStruct) getPostsTree(tid, lim, from int, desc bool) ([]f.PostThread, error) {
	var query string
	q := ""
	if from != 0 {
		if desc {
			q = `AND PATH < `
		} else {
			q = `AND PATH > `
		}
		q += fmt.Sprintf(`(SELECT path FROM post WHERE id = %d)`, from)
	}
	if desc {
		query = fmt.Sprintf(
			`SELECT * FROM post WHERE thread=$1 %s ORDER BY path DESC, id DESC LIMIT NULLIF($2, 0);`, q)
	} else {
		query = fmt.Sprintf(
			`SELECT * FROM post WHERE thread=$1 %s ORDER BY path, id LIMIT NULLIF($2, 0);`, q)
	}
	var posts []f.PostThread
	row, err := p.Connection.Query(query, tid, lim)

	if err != nil {
		return posts, err
	}
	defer func() {
		if row != nil {
			row.Close()
		}
	}()

	for row.Next() {
		var post f.PostThread
		var created time.Time

		err = row.Scan(&post.Author, &created, &post.Forum, &post.Id, &post.IsEdited, &post.Message,
			&post.Parent, &post.Thread, &post.Path)

		if err != nil {
			return posts, err
		}
		post.Created = strfmt.DateTime(created.UTC()).String()
		posts = append(posts, post)

	}
	return posts, err
}

func (p *RepStruct) getPostsParentTree(tid, lim, from int, desc bool) ([]f.PostThread, error) {
	var query string
	q := ""
	if from != 0 {
		if desc {
			q = `AND PATH[1] < `
		} else {
			q = `AND PATH[1] > `
		}
		q += fmt.Sprintf(`(SELECT path[1] FROM post WHERE id = %d)`, from)
	}

	parent := fmt.Sprintf(
		`SELECT id FROM post WHERE thread = $1 AND parent IS NULL %s`, q)

	if desc {
		parent += `ORDER BY id DESC`
		if lim > 0 {
			parent += fmt.Sprintf(` LIMIT %d`, lim)
		}
		query = fmt.Sprintf(
			`SELECT * FROM post WHERE path[1] IN (%s) ORDER BY path[1] DESC, path, id;`, parent)
	} else {
		parent += `ORDER BY id`
		if lim > 0 {
			parent += fmt.Sprintf(` LIMIT %d`, lim)
		}
		query = fmt.Sprintf(
			`SELECT * FROM post WHERE path[1] IN (%s) ORDER BY path,id;`, parent)
	}
	var ans []f.PostThread
	row, err := p.Connection.Query(query, tid)

	if err != nil {
		return ans, err
	}

	defer func() {
		if row != nil {
			row.Close()
		}
	}()

	for row.Next() {
		var post f.PostThread
		var created time.Time

		err = row.Scan(&post.Author, &created, &post.Forum, &post.Id, &post.IsEdited, &post.Message,
			&post.Parent, &post.Thread, &post.Path)

		if err != nil {
			return ans, err
		}
		post.Created = strfmt.DateTime(created.UTC()).String()
		ans = append(ans, post)

	}
	return ans, err
}

func (p *RepStruct) GetPostsWithParams(postSlugOrId f.ThreadForum, lim, from int, sort string, desc bool) ([]f.PostThread, error) {
	var err error
	threadId := 0
	if postSlugOrId.Id <= 0 {
		threadId, err = p.ThreadIDBySlug(postSlugOrId.Slug.String)
		if err != nil {
			return nil, err
		}
	} else {
		threadId = int(postSlugOrId.Id)
	}

	switch sort {
	case "flat":
		return p.getPostsFlat(threadId, lim, from, desc)
	case "tree":
		return p.getPostsTree(threadId, lim, from, desc)
	case "parent_tree":
		return p.getPostsParentTree(threadId, lim, from, desc)
	default:
		return nil, errors.New("THERE IS NO SORT WITH THIS NAME")
	}
}

func (p *RepStruct) GetPostRelated(id int, related []string) (map[string]interface{}, error) {
	query := `SELECT * FROM post WHERE id = $1;`
	var post f.PostThread
	var created time.Time

	err := p.Connection.QueryRow(query, id).Scan(&post.Author, &created, &post.Forum,
		&post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread, &post.Path)
	post.Created = strfmt.DateTime(created.UTC()).String()

	returnMap := map[string]interface{}{
		"post": post,
	}

	for _, relatedObj := range related {
		switch relatedObj {
		case "user":
			author, err := p.Rep.GetUserByNick(post.Author)
			if err != nil {
				return returnMap, err
			}
			returnMap["author"] = author
		case "thread":
			thread, err := p.ThreadByID(int(post.Thread))
			if err != nil {
				return returnMap, err
			}
			returnMap["thread"] = thread
		case "forum":
			forumAns, err := p.ForumBySlug(post.Forum)
			if err != nil {
				return returnMap, err
			}
			returnMap["forum"] = forumAns
		}
	}

	return returnMap, err
}

func (p *RepStruct) UpdatePost(newPost f.PostThread) (f.PostThread, error) {
	query := `UPDATE post SET message = $1, isEdited = true WHERE id = $2 RETURNING *;`

	oldPost, err := p.GetPostRelated(int(newPost.Id), []string{})
	if err != nil {
		return f.PostThread{}, err
	}
	if oldPost["post"].(f.PostThread).Message == newPost.Message {
		return oldPost["post"].(f.PostThread), nil
	}

	if newPost.Message == "" {
		query := `SELECT * FROM post WHERE id = $1`
		var post f.PostThread
		var created time.Time

		err := p.Connection.QueryRow(query, newPost.Id).Scan(&post.Author, &created,
			&post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread, &post.Path)

		post.Created = strfmt.DateTime(created.UTC()).String()
		return post, err
	}

	var post f.PostThread
	var created time.Time

	err = p.Connection.QueryRow(query, newPost.Message, newPost.Id).Scan(&post.Author, &created,
		&post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread, &post.Path)
	post.Created = strfmt.DateTime(created.UTC()).String()

	return post, err
}

func (p *RepStruct) CreatePosts(posts []f.PostThread, tid int) ([]f.PostThread, error) {
	query := `INSERT INTO post(
                 author,
                 created,
                 message,
                 parent,
				 thread,
				 forum) VALUES `
	data := make([]f.PostThread, 0, 0)
	if len(posts) == 0 {
		return data, nil
	}

	slug, err := p.getForumSlug(tid)
	if err != nil {
		return data, err
	}

	Created := time.Now()
	var valNam []string
	var values []interface{}
	i := 1
	for _, element := range posts {
		valNam = append(valNam, fmt.Sprintf(
			"($%d, $%d, $%d, nullif($%d, 0), $%d, $%d)",
			i, i+1, i+2, i+3, i+4, i+5))
		i += 6
		values = append(values, element.Author, Created, element.Message, element.Parent, tid, slug)
	}

	query += strings.Join(valNam[:], ",")
	query += " RETURNING *"
	row, err := p.Connection.Query(query, values...)

	if err != nil {
		return data, err
	}
	defer func() {
		if row != nil {
			row.Close()
		}
	}()

	for row.Next() {
		var post f.PostThread
		var created time.Time

		err = row.Scan(&post.Author, &created, &post.Forum, &post.Id, &post.IsEdited,
			&post.Message, &post.Parent, &post.Thread, &post.Path)

		if err != nil {
			return data, err
		}
		post.Created = strfmt.DateTime(created.UTC()).String()
		data = append(data, post)

	}
	return data, err
}
