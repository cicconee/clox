ALTER TABLE directories
ADD CONSTRAINT unique_directory_name_parent
UNIQUE (name, parent_id);