CREATE TABLE IF NOT EXISTS clothes (
        id bigserial PRIMARY KEY,
        name text NOT NULL,
        price integer NOT NULL,
        brand text NOT NULL,
        color text NOT NULL,
        sizes text[] NOT NULL,
        sex text NOT NULL,
        type text NOT NULL,
        image_url text NOT NULL
);