---------- /forum/create ----------
INSERT INTO forums (slug, title, user_id)
VALUES (?, ?, (SELECT id FROM users WHERE nickname = ?));

-- INSERT INTO forums (slug, title, user_id)
-- VALUES ('some', 'Some forum', (SELECT id FROM users WHERE nickname = 'viewsharp'));


---------- /forum/{slug}/details ----------
SELECT * FROM forums WHERE slug = ?;

-- SELECT * FROM forums WHERE slug = 'some';


---------- /user/{nickname}/create ----------
INSERT INTO users (nickname, fullname, email, about)
VALUES (?, ?, ?, ?);
-- or
INSERT INTO users (nickname, fullname, email)
VALUES (?, ?, ?);

-- INSERT INTO users (nickname, fullname, email)
-- VALUES ('viewsharp', 'Vladimir Atamanov', 'viewsharp@yandex.ru');


---------- /user/{nickname}/profile ----------
SELECT * FROM users WHERE nickname = ?;

-- SELECT * FROM users WHERE nickname = 'viewsharp';