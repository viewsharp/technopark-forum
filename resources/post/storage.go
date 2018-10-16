package post

import (
	"database/sql"
	"fmt"
	"strings"
)

type Storage struct {
	DB *sql.DB
}

func (s *Storage) AddByThreadSlug(posts *Posts, slug string) error {
	var threadId int
	var forumSlug string

	err := s.DB.QueryRow(
		`	SELECT id, forum_slug
            	FROM threads
              	WHERE slug = $1`,
		slug,
	).Scan(&threadId, &forumSlug)

	if err != nil {
		return ErrUnknown
	}

	// TODO: handle errors

	return s.add(posts, threadId, forumSlug)
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

	// TODO: handle errors

	return ErrUnknown
}

func (s *Storage) add(posts *Posts, threadId int, forumSlug string) error {
	queryParams := make([]interface{}, 0, 4*len(*posts))

	var queryBuilder strings.Builder
	queryBuilder.WriteString(`	INSERT INTO posts (user_nn, message, parent_id, thread_id)
									VALUES `)

	for i, post := range *posts {
		if i != 0 {
			queryBuilder.WriteString(",")
		}
		queryBuilder.WriteString(fmt.Sprintf(" ($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		queryParams = append(queryParams, post.Author, post.Message, post.Parent, threadId)
	}
	queryBuilder.WriteString(" RETURNING id, created, thread_id")

	rows, err := s.DB.Query(queryBuilder.String(), queryParams...)
	if err != nil {
		return ErrUnknown
	}
	defer rows.Close()

	for i, _ := range *posts {
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

func (s *Storage) ById(id int) (*Post, error) {
	var result Post
	var err error
	//err := s.DB.QueryRow(
	//	` SELECT id, user_nn, created, forum_slug, message, p.parent_id, p.thread_id
	//			FROM posts
	//			WHERE nickname = $1 `,
	//	nickname,
	//).Scan(&result.Nickname, &result.FullName, &result.Email, &result.About)

	if err == nil {
		return &result, nil
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return nil, ErrNotFound
	}

	return nil, ErrUnknown
}

func (s *Storage) FlatByThreadSlug(slug string, limit int, desc bool, since int) (Posts, error) {
	dbQuery := func() (*sql.Rows, error) {
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

		if since != 0 {
			return s.DB.Query(queryBuilder.String(), slug, limit, since)
		} else {
			return s.DB.Query(queryBuilder.String(), slug, limit)
		}
	}

	return getWrapper(dbQuery)
}

func (s *Storage) FlatByThreadId(id int, limit int, desc bool, since int) (Posts, error) {
	dbQuery := func() (*sql.Rows, error) {
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

		if since != 0 {
			return s.DB.Query(queryBuilder.String(), id, limit, since)
		} else {
			return s.DB.Query(queryBuilder.String(), id, limit)
		}
	}

	return getWrapper(dbQuery)
}

func (s *Storage) TreeByThreadSlug(slug string, limit int, desc bool, since int) (Posts, error) {
	dbQuery := func() (*sql.Rows, error) {
		var queryBuilder strings.Builder
		queryBuilder.WriteString(
			`	WITH RECURSIVE recurseposts (user_nn, created, forum_slug, id, message, parent_id, thread_id, path) AS (
    						SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, t.id, ARRAY[p.id] as path
    						FROM posts p
    					  		JOIN threads t on p.thread_id = t.id
    					    WHERE p.parent_id is NULL AND t.slug = $1
    					    UNION ALL
    					        SELECT p.user_nn, p.created, rp.forum_slug, p.id, p.message, p.parent_id, p.thread_id, rp.path || p.id
    					        FROM posts p
    					            JOIN recurseposts rp ON rp.id = p.parent_id)
    					SELECT rp.user_nn, rp.created, rp.forum_slug, rp.id, rp.message, rp.parent_id, rp.thread_id
    					FROM recurseposts rp`,
		)

		if since != 0 {
			if desc {
				queryBuilder.WriteString(" JOIN recurseposts ON recurseposts.id = $3 WHERE rp.path < recurseposts.path")
			} else {
				queryBuilder.WriteString(" JOIN recurseposts ON recurseposts.id = $3 WHERE rp.path > recurseposts.path")
			}
		}

		queryBuilder.WriteString(" ORDER BY rp.path")
		if desc {
			queryBuilder.WriteString(" DESC")
		}
		queryBuilder.WriteString(" LIMIT $2")

		if since != 0 {
			return s.DB.Query(queryBuilder.String(), slug, limit, since)
		}
		return s.DB.Query(queryBuilder.String(), slug, limit)
	}

	return getWrapper(dbQuery)
}

func (s *Storage) TreeByThreadId(id int, limit int, desc bool, since int) (Posts, error) {
	dbQuery := func() (*sql.Rows, error) {
		var queryBuilder strings.Builder
		queryBuilder.WriteString(
			`	WITH RECURSIVE recurseposts (user_nn, created, forum_slug, id, message, parent_id, thread_id, path) AS (
    						SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, t.id, ARRAY[p.id] as path
    						FROM posts p
    					  		JOIN threads t on p.thread_id = t.id
    					    WHERE p.parent_id is NULL AND t.id = $1
    					    UNION ALL
    					        SELECT p.user_nn, p.created, rp.forum_slug, p.id, p.message, p.parent_id, p.thread_id, rp.path || p.id
    					        FROM posts p
    					            JOIN recurseposts rp ON rp.id = p.parent_id)
    					SELECT rp.user_nn, rp.created, rp.forum_slug, rp.id, rp.message, rp.parent_id, rp.thread_id
    					FROM recurseposts rp`,
		)

		if since != 0 {
			if desc {
				queryBuilder.WriteString(" JOIN recurseposts ON recurseposts.id = $3 WHERE rp.path < recurseposts.path")
			} else {
				queryBuilder.WriteString(" JOIN recurseposts ON recurseposts.id = $3 WHERE rp.path > recurseposts.path")
			}
		}

		queryBuilder.WriteString(" ORDER BY rp.path")
		if desc {
			queryBuilder.WriteString(" DESC")
		}
		queryBuilder.WriteString(" LIMIT $2")

		if since != 0 {
			return s.DB.Query(queryBuilder.String(), id, limit, since)
		}
		return s.DB.Query(queryBuilder.String(), id, limit)
	}

	return getWrapper(dbQuery)
}

func (s *Storage) ParentTreeByThreadSlug(slug string, limit int, desc bool, since int) (Posts, error) {
	dbQuery := func() (*sql.Rows, error) {
		var queryBuilder strings.Builder
		queryBuilder.WriteString(`WITH RECURSIVE recurseposts (user_nn, created, forum_slug, id, message, parent_id, thread_id, path, rank) AS (
									 SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, t.id, ARRAY[p.id] as path, row_number() OVER`)

		if desc {
			queryBuilder.WriteString(" (ORDER BY p.id DESC)")
		} else {
			queryBuilder.WriteString(" (ORDER BY p.id)")
		}

		queryBuilder.WriteString(
			`FROM posts p
					JOIN threads t on p.thread_id = t.id
				WHERE p.parent_id is NULL AND t.slug = $1
				UNION ALL
					SELECT p.user_nn, p.created, rp.forum_slug, p.id, p.message, p.parent_id, p.thread_id, rp.path || p.id, rp.rank
					FROM posts p
						JOIN recurseposts rp ON rp.id = p.parent_id)
				SELECT rp.user_nn, rp.created, rp.forum_slug, rp.id, rp.message, rp.parent_id, rp.thread_id
				FROM recurseposts rp`)

		if since != 0 {

			queryBuilder.WriteString(`	JOIN recurseposts ON recurseposts.id = $3 
											WHERE rp.rank <= $2 + recurseposts.rank AND (rp.rank > recurseposts.rank OR rp.rank = recurseposts.rank AND rp.path > recurseposts.path) 
											ORDER BY rp.rank, rp.path`)
		} else {
			queryBuilder.WriteString(`	WHERE rp.rank <= $2 
											ORDER BY rp.rank, rp.path`)
		}

		if since != 0 {
			fmt.Println(queryBuilder.String())
			return s.DB.Query(queryBuilder.String(), slug, limit, since)
		}
		fmt.Println(queryBuilder.String())
		return s.DB.Query(queryBuilder.String(), slug, limit)
	}

	return getWrapper(dbQuery)
}

func (s *Storage) ParentTreeByThreadId(id int, limit int, desc bool, since int) (Posts, error) {
	dbQuery := func() (*sql.Rows, error) {
		var queryBuilder strings.Builder
		queryBuilder.WriteString(`WITH RECURSIVE recurseposts (user_nn, created, forum_slug, id, message, parent_id, thread_id, path, rank) AS (
									 SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, t.id, ARRAY[p.id] as path, row_number() OVER`)

		if desc {
			queryBuilder.WriteString(" (ORDER BY p.id DESC)")
		} else {
			queryBuilder.WriteString(" (ORDER BY p.id)")
		}

		queryBuilder.WriteString(
			`FROM posts p
					JOIN threads t on p.thread_id = t.id
				WHERE p.parent_id is NULL AND t.id = $1
				UNION ALL
					SELECT p.user_nn, p.created, rp.forum_slug, p.id, p.message, p.parent_id, p.thread_id, rp.path || p.id, rp.rank
					FROM posts p
						JOIN recurseposts rp ON rp.id = p.parent_id)
				SELECT rp.user_nn, rp.created, rp.forum_slug, rp.id, rp.message, rp.parent_id, rp.thread_id
				FROM recurseposts rp`)

		if since != 0 {

			queryBuilder.WriteString(`	JOIN recurseposts ON recurseposts.id = $3 
											WHERE rp.rank <= $2 + recurseposts.rank AND (rp.rank > recurseposts.rank OR rp.rank = recurseposts.rank AND rp.path > recurseposts.path) 
											ORDER BY rp.rank, rp.path`)
		} else {
			queryBuilder.WriteString(`	WHERE rp.rank <= $2 
											ORDER BY rp.rank, rp.path`)
		}

		if since != 0 {
			fmt.Println(queryBuilder.String())
			return s.DB.Query(queryBuilder.String(), id, limit, since)
		}
		fmt.Println(queryBuilder.String())
		return s.DB.Query(queryBuilder.String(), id, limit)
	}

	return getWrapper(dbQuery)
}

func getWrapper(dbQuery func() (*sql.Rows, error)) (Posts, error) {
	rows, err := dbQuery()
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
	return posts, nil
}
