CREATE TABLE IF NOT EXISTS users (
    user_id serial PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS segments (
                                        segment_id serial PRIMARY KEY,
                                        slug varchar(255) NOT NULL UNIQUE
    );

CREATE TABLE user_segments (
                               user_id INT,
                               segment_id INT,
                               FOREIGN KEY (user_id) REFERENCES users(user_id),
                               FOREIGN KEY (segment_id) REFERENCES segments(segment_id),
                               PRIMARY KEY (user_id, segment_id)
);

INSERT INTO users SELECT generate_series(1, 100);
