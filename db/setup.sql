/* create schema */
CREATE SCHEMA IF NOT EXISTS auxstream; 

/* create tables */
CREATE TABLE IF NOT EXISTS auxstream.tracks (
    id serial primary key,
    title text not null,
    artist_id integer references auxstream.artists(id),
    file text not null,
    created_at timestamp default CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS auxstream.artists (
    id serial primary key, 
    name text not null, 
    created_at time default CURRENT_TIMESTAMP
)
