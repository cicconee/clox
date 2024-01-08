ALTER TABLE user_tokens
DROP CONSTRAINT user_tokens_user_id_fkey,
ADD CONSTRAINT user_tokens_user_id_fkey FOREIGN KEY (user_id)
REFERENCES users(id)
ON DELETE CASCADE;