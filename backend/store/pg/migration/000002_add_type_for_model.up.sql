-- add type for model
alter table models add column type varchar(255) not null default 'chat';

-- add unique index for type
create unique index idx_models_type on models (type);
