CREATE UNIQUE INDEX unique_user_root_directory
ON directories (user_id, name)
WHERE parent_id IS NULL;