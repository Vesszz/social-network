ALTER TABLE users ADD COLUMN created_at TIMESTAMP;
WITH earliest_post AS (
    SELECT 
        author_id,
        MIN(created_at) AS min_created_at
    FROM posts
    GROUP BY author_id
)
UPDATE users
SET created_at = COALESCE(earliest_post.min_created_at, CURRENT_TIMESTAMP)
FROM earliest_post
WHERE users.id = earliest_post.author_id;
ALTER TABLE users ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;
