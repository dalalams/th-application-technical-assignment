-- +goose Up
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE series (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES categories(id) NOT NULL,
    language VARCHAR(10),
    series_type TEXT NOT NULL CHECK (series_type IN ('documentary', 'podcast')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE episodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    series_id UUID REFERENCES series(id) ON DELETE CASCADE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    duration_seconds INT,
    publish_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE episode_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE NOT NULL,
    asset_type TEXT NOT NULL CHECK (asset_type IN ('audio', 'video', 'thumbnail')),
    mime_type TEXT NOT NULL,
    size_bytes BIGINT,
    url TEXT,
    storage JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_series_category ON series(category_id);
CREATE INDEX idx_episodes_series ON episodes(series_id);
CREATE INDEX idx_episodes_publish_date ON episodes(publish_date);
CREATE INDEX idx_assets_episode ON episode_assets(episode_id);

-- +goose Down
DROP TABLE IF EXISTS episode_assets;
DROP TABLE IF EXISTS episodes;
DROP TABLE IF EXISTS series;
DROP TABLE IF EXISTS categories;


