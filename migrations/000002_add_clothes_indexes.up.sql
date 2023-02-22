CREATE INDEX IF NOT EXISTS clothes_name_idx ON clothes USING GIN (to_tsvector('simple', name));
CREATE INDEX IF NOT EXISTS clothes_sizes_idx ON clothes USING GIN (sizes);