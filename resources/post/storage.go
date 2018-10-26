package post

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/viewsharp/TexPark_DBMSs/resources/forum"
	"github.com/viewsharp/TexPark_DBMSs/resources/thread"
	"github.com/viewsharp/TexPark_DBMSs/resources/user"
	"regexp"
	"strings"
	"sync"
)

type Storage struct {
	DB           *sql.DB
	postInsertMX sync.Mutex
}

var regexInvalidAuthor, _ = regexp.Compile(`^Key \(user_nn\)=\(([\w\.]+)\) is not present in table "users"\.$`)

func (s *Storage) AddByThreadSlug(posts *Posts, slug string) error {
	var threadId int
	var forumSlug string

	err := s.DB.QueryRow(
		`	SELECT id, forum_slug
            	FROM threads
              	WHERE slug = $1`,
		slug,
	).Scan(&threadId, &forumSlug)

	if err == nil {
		return s.add(posts, threadId, forumSlug)
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return ErrNotFoundThread
	}

	return ErrUnknown
}

func (s *Storage) AddByThreadId(posts *Posts, threadId int) error {
	var forumSlug string

	err := s.DB.QueryRow(
		`	SELECT forum_slug
            	FROM threads
              	WHERE id = $1`,
		threadId,
	).Scan(&forumSlug)

	if err == nil {
		return s.add(posts, threadId, forumSlug)
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return ErrNotFoundThread
	}

	return ErrUnknown
}

func (s *Storage) add(posts *Posts, threadId int, forumSlug string) error {
	queryParams := make([]interface{}, 0, 4*len(*posts))
	var queryBuilder strings.Builder

	queryBuilder.WriteString("INSERT INTO posts (user_nn, message, parent_id, thread_id, path) VALUES")
	count := 0

	for _, post := range *posts {
		if count != 0 {
			queryBuilder.WriteString(",")
		}
		if post.Parent == nil {
			queryBuilder.WriteString(fmt.Sprintf(" ($%d, $%d, NULL, $%d, NULL)", count+1, count+2, count+3))
			queryParams = append(queryParams, post.Author, post.Message, threadId)
			count += 3
		} else {
			queryBuilder.WriteString(fmt.Sprintf(" ($%d, $%d, $%d, "+
				"(SELECT thread_id FROM posts WHERE id = $%d AND thread_id = $%d), "+
				"(SELECT path || id FROM posts WHERE id = $%d AND thread_id = $%d))",
				count+1, count+2, count+3, count+3, count+4, count+3, count+4))
			queryParams = append(queryParams, post.Author, post.Message, post.Parent, threadId)
			count += 4
		}
	}
	queryBuilder.WriteString(" RETURNING id, created, thread_id")

	rows, err := s.DB.Query(queryBuilder.String(), queryParams...)
	if err != nil {
		switch err.(*pq.Error).Code.Name() {
		case "not_null_violation":
			return ErrInvalidParent
		case "foreign_key_violation":
			ErrNotFoundUser.setNickname(regexInvalidAuthor.FindStringSubmatch(err.(*pq.Error).Detail)[1])
			return ErrNotFoundUser
		}

		return ErrUnknown
	}
	defer rows.Close()

	_, err = s.DB.Exec("UPDATE forums SET posts = posts+$1 WHERE slug = $2", len(*posts), forumSlug)
	if err != nil {
		panic(err)
	}

	go s.addInsertForumUsers(posts, forumSlug)

	for i := range *posts {
		rows.Next()
		post := (*posts)[i]

		err = rows.Scan(&post.Id, &post.Created, &post.Thread)
		if err != nil {
			return ErrUnknown
		}
		post.Forum = &forumSlug
	}

	return nil
}

func (s *Storage) addInsertForumUsers(posts *Posts, forumSlug string) {
	queryParams := make([]interface{}, 0, 2*len(*posts))
	var queryBuilder strings.Builder

	queryBuilder.WriteString("INSERT INTO forum_user (forum_slug, user_nn) VALUES")
	count := 0

	for _, post := range *posts {
		if count != 0 {
			queryBuilder.WriteString(",")
		}
		queryBuilder.WriteString(fmt.Sprintf(" ($%d, $%d)", count+1, count+2))
		queryParams = append(queryParams, forumSlug, post.Author)
		count += 2
	}

	queryBuilder.WriteString(" ON CONFLICT DO NOTHING")

	s.postInsertMX.Lock()

	_, err := s.DB.Exec(queryBuilder.String(), queryParams...)
	if err != nil {
		panic(err)
	}

	s.postInsertMX.Unlock()
}

func (s *Storage) ById(id int, related []string) (*PostFull, error) {
	userObj := user.User{}
	forumObj := forum.Forum{}
	postObj := Post{}
	threadObj := thread.Thread{}
	result := PostFull{}

	err := s.DB.QueryRow(
		`	SELECT 
					u.about, u.email, u.fullname, u.nickname, 
					f.posts, f.slug, f.threads, f.title, f.user_nn, 
					p.user_nn, p.created, f.slug, p.id, p.isedited, p.message, p.parent_id, p.thread_id,
					t.user_nn, t.created, f.slug, t.id, t.message, t.slug, t.title, t.votes
				FROM posts p
					JOIN users u on p.user_nn = u.nickname
					JOIN threads t on p.thread_id = t.id
					JOIN forums f on t.forum_slug = f.slug
				WHERE p.id = $1`,
		id,
	).Scan(
		&userObj.About, &userObj.Email, &userObj.FullName, &userObj.Nickname,
		&forumObj.Posts, &forumObj.Slug, &forumObj.Threads, &forumObj.Title, &forumObj.User,
		&postObj.Author, &postObj.Created, &postObj.Forum, &postObj.Id, &postObj.IsEdited, &postObj.Message, &postObj.Parent, &postObj.Thread,
		&threadObj.Author, &threadObj.Created, &threadObj.Forum, &threadObj.Id, &threadObj.Message, &threadObj.Slug, &threadObj.Title, &threadObj.Votes,
	)

	if err == nil {
		result.Post = &postObj
		for _, relate := range related {
			switch relate {
			case "user":
				result.Author = &userObj
			case "thread":
				result.Thread = &threadObj
			case "forum":
				result.Forum = &forumObj
			}
		}
		return &result, nil
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return nil, ErrNotFound
	}

	return nil, ErrUnknown
}

func (s *Storage) UpdateById(id int, post PostUpdate) error {
	if post.Message == nil {
		return nil
	}

	_, err := s.DB.Exec(
		`	UPDATE posts 
				SET message = $1, isedited = TRUE
				WHERE id = $2`,
		post.Message, id,
	)

	if err == nil {
		return nil
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return ErrNotFound
	}

	return ErrUnknown
}

func (s *Storage) FlatByThreadSlug(slug string, limit int, desc bool, since int) (Posts, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`	SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, p.thread_id
										FROM posts p
											JOIN threads t on p.thread_id = t.id
										WHERE t.slug = $1`)

	if since != 0 {
		if desc {
			queryBuilder.WriteString(" AND p.id < $3")
		} else {
			queryBuilder.WriteString(" AND p.id > $3")
		}
	}

	if desc {
		queryBuilder.WriteString(" ORDER BY p.created DESC, p.id DESC LIMIT $2")
	} else {
		queryBuilder.WriteString(" ORDER BY p.created, p.id LIMIT $2")
	}

	return s.bySlug(queryBuilder.String(), slug, limit, since)
}

func (s *Storage) FlatByThreadId(id int, limit int, desc bool, since int) (Posts, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`	SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, p.thread_id
										FROM posts p
											JOIN threads t on p.thread_id = t.id
										WHERE t.id = $1`)

	if since != 0 {
		if desc {
			queryBuilder.WriteString(" AND p.id < $3")
		} else {
			queryBuilder.WriteString(" AND p.id > $3")
		}
	}

	if desc {
		queryBuilder.WriteString(" ORDER BY p.created DESC, p.id DESC LIMIT $2")
	} else {
		queryBuilder.WriteString(" ORDER BY p.created, p.id LIMIT $2")
	}

	return s.byId(queryBuilder.String(), id, limit, since)
}

func (s *Storage) TreeByThreadSlug(slug string, limit int, desc bool, since int) (Posts, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(
		`	SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, p.thread_id
    			FROM posts p JOIN threads t ON t.id = p.thread_id`,
	)

	if since != 0 {
		if desc {
			queryBuilder.WriteString(" JOIN posts ON posts.id = $3 WHERE p.path || p.id < posts.path || posts.id")
		} else {
			queryBuilder.WriteString(" JOIN posts ON posts.id = $3 WHERE p.path || p.id > posts.path || posts.id")
		}
		queryBuilder.WriteString(" AND t.slug = $1 ORDER BY p.path || p.id")
	} else {
		queryBuilder.WriteString(" WHERE t.slug = $1 ORDER BY p.path || p.id")
	}

	if desc {
		queryBuilder.WriteString(" DESC")
	}
	queryBuilder.WriteString(" LIMIT $2")

	return s.bySlug(queryBuilder.String(), slug, limit, since)
}

func (s *Storage) TreeByThreadId(id int, limit int, desc bool, since int) (Posts, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(
		`	SELECT p.user_nn, p.created, (SELECT forum_slug FROM threads WHERE id = $1), p.id, p.message, p.parent_id, p.thread_id
    			FROM posts p`,
	)

	if since != 0 {
		if desc {
			queryBuilder.WriteString(" JOIN posts ON posts.id = $3 WHERE p.path || p.id < posts.path || posts.id")
		} else {
			queryBuilder.WriteString(" JOIN posts ON posts.id = $3 WHERE p.path || p.id > posts.path || posts.id")
		}
		queryBuilder.WriteString(" AND p.thread_id = $1 ORDER BY p.path || p.id")
	} else {
		queryBuilder.WriteString(" WHERE p.thread_id = $1 ORDER BY p.path || p.id")
	}

	if desc {
		queryBuilder.WriteString(" DESC")
	}
	queryBuilder.WriteString(" LIMIT $2")

	return s.byId(queryBuilder.String(), id, limit, since)
}
func (s *Storage) ParentTreeByThreadSlug(slug string, limit int, desc bool, since int) (Posts, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("WITH ranked_posts AS (SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, p.thread_id,p.path || p.id AS path,")

	if desc {
		queryBuilder.WriteString(" dense_rank() over (ORDER BY COALESCE(path [1], p.id) desc) AS rank")
	} else {
		queryBuilder.WriteString(" dense_rank() over (ORDER BY COALESCE(path [1], p.id)) AS rank")
	}
	queryBuilder.WriteString(
		`	FROM posts p JOIN threads t on p.thread_id = t.id WHERE t.slug = $1)
				SELECT p.user_nn, p.created, p.forum_slug, p.id, p.message, p.parent_id, p.thread_id 
				FROM ranked_posts p`)

	if since != 0 {
		queryBuilder.WriteString(
			`	JOIN ranked_posts posts ON posts.id = $3 
				WHERE p.rank <= $2 + posts.rank AND (p.rank > posts.rank OR p.rank = posts.rank AND p.path > posts.path) 
				ORDER BY p.rank, p.path`)
	} else {
		queryBuilder.WriteString(" WHERE p.rank <= $2 ORDER BY p.rank, p.path")
	}

	return s.bySlug(queryBuilder.String(), slug, limit, since)
}

func (s *Storage) ParentTreeByThreadId(id int, limit int, desc bool, since int) (Posts, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("WITH ranked_posts AS (SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, p.thread_id,p.path || p.id AS path,")

	if desc {
		queryBuilder.WriteString(" dense_rank() over (ORDER BY COALESCE(path [1], p.id) desc) AS rank")
	} else {
		queryBuilder.WriteString(" dense_rank() over (ORDER BY COALESCE(path [1], p.id)) AS rank")
	}
	queryBuilder.WriteString(
		`	FROM posts p JOIN threads t on p.thread_id = t.id WHERE t.id = $1)
				SELECT p.user_nn, p.created, p.forum_slug, p.id, p.message, p.parent_id, p.thread_id 
				FROM ranked_posts p`)

	if since != 0 {
		queryBuilder.WriteString(
			`	JOIN ranked_posts posts ON posts.id = $3 
				WHERE p.rank <= $2 + posts.rank AND (p.rank > posts.rank OR p.rank = posts.rank AND p.path > posts.path) 
				ORDER BY p.rank, p.path`)
	} else {
		queryBuilder.WriteString(" WHERE p.rank <= $2 ORDER BY p.rank, p.path")
	}

	return s.byId(queryBuilder.String(), id, limit, since)
}

func (s *Storage) byId(query string, id int, limit int, since int) (Posts, error) {
	var rows *sql.Rows
	var err error
	if since != 0 {
		rows, err = s.DB.Query(query, id, limit, since)
	} else {
		rows, err = s.DB.Query(query, id, limit)
	}

	if err != nil {
		return nil, ErrUnknown
	}
	defer rows.Close()

	posts := make(Posts, 0, 1)
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			return nil, ErrUnknown
		}
		posts = append(posts, &post)
	}

	if len(posts) == 0 {
		err := s.DB.QueryRow("SELECT id FROM threads WHERE id = $1", id).Scan(&id)
		if err != nil {
			return nil, ErrNotFoundThread
		}
	}

	return posts, nil
}

func (s *Storage) bySlug(query string, slug string, limit int, since int) (Posts, error) {
	var rows *sql.Rows
	var err error
	if since != 0 {
		rows, err = s.DB.Query(query, slug, limit, since)
	} else {
		rows, err = s.DB.Query(query, slug, limit)
	}

	if err != nil {
		return nil, ErrUnknown
	}
	defer rows.Close()

	posts := make(Posts, 0, 1)
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			return nil, ErrUnknown
		}
		posts = append(posts, &post)
	}

	if len(posts) == 0 {
		err := s.DB.QueryRow("SELECT slug FROM threads WHERE slug = $1", slug).Scan(&slug)
		if err != nil {
			return nil, ErrNotFoundThread
		}
	}

	return posts, nil
}
