-- +goose Up
-- +goose StatementBegin
create table "user"
(
    id         serial primary key,
    name       text      not null,
    email      text      not null,
    role       smallint  not null,
    password   text      not null,
    updated_at timestamp not null default now(),
    created_at timestamp not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table "user";
-- +goose StatementEnd
