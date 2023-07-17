/* create schema */
CREATE SCHEMA IF NOT EXISTS auxstream;


CREATE TABLE IF NOT EXISTS auxstream.artists (
    id serial primary key,
    name text not null unique,
    created_at time default CURRENT_TIMESTAMP
);

/* create tables */
CREATE TABLE IF NOT EXISTS auxstream.tracks (
    id serial primary key,
    title text not null,
    artist_id integer references auxstream.artists(id),
    file text not null,
    created_at timestamp default CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS auxstream.users (
    id serial primary key,
    username text not null,
    password_hash text not null,
    created_at timestamp default CURRENT_TIMESTAMP
);
