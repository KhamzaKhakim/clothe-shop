CREATE TABLE IF NOT EXISTS brands (
                                       id bigserial PRIMARY KEY,
                                       name text NOT NULL,
                                       country text NOT NULL,
                                       description text NOT NULL,
                                       image_url text NOT NULL
);