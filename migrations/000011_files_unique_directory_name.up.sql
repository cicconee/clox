ALTER TABLE files
ADD CONSTRAINT unique_file_directory_name
UNIQUE (directory_id, name);