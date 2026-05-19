CREATE TABLE items (
	id SERIAL PRIMARY KEY,
	publishedfileid BIGINT NOT NULL,
	title TEXT NOT NULL,
	time_created BIGINT NOT NULL,
	time_updated BIGINT NOT NULL,
	creator BIGINT NOT NULL,
	tags text[] NOT NULL,
	file_size BIGINT NOT NULL,
	file_url TEXT NOT NULL,
	preview_url TEXT NOT NULL,
	subscriptions BIGINT NOT NULL,
	consumer_appid BIGINT NOT NULL,
	maybe_inappropriate_sex BOOLEAN NOT NULL,
	maybe_inappropriate_violence BOOLEAN NOT NULL,
	detail JSONB NOT NULL
);
