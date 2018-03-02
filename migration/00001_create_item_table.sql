-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE EXTENSION pg_trgm;
CREATE TABLE item
(
    id        serial primary key,
    price     integer default 0,
    diff      integer default 0,
    name      text default '',
    category  text default '',
    url       text default '',
    imgsrc       text default '',
    source       text default '',
    note       text default '',
    created timestamp default NOW(),
    updated timestamp
);

CREATE INDEX item_id ON item ( id );
CREATE INDEX item_price ON item ( price );
CREATE INDEX item_url ON item ( url );
create index item_name on item using gin(name gin_trgm_ops);


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE item;
DROP EXTENSION pg_trgm;
