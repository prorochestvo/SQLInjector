--- #migrate:Up
INSERT INTO os_files_01 (id) VALUES ('bd5dc0fa-db1a-4e15-bea4-34c4fcc2133b');

--- #migrate:Down
DELETE FROM os_files_01;

