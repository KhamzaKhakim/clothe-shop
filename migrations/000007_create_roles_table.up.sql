CREATE TABLE IF NOT EXISTS roles(
                                           id bigserial PRIMARY KEY,
                                           role text NOT NULL
);
CREATE TABLE IF NOT EXISTS users_roles (
                                                 user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
                                                 roles_id bigint NOT NULL REFERENCES roles ON DELETE CASCADE,
                                                 PRIMARY KEY (user_id, roles_id)
);
-- Add the two permissions to the table.
INSERT INTO roles (role)
VALUES
    ('ADMIN'),
    ('USER');
