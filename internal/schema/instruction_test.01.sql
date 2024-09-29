--- #migrate:Up
CREATE TABLE os_files_01 (id TEXT PRIMARY KEY);

--- #migrate:Down
DROP TABLE os_files_01;

