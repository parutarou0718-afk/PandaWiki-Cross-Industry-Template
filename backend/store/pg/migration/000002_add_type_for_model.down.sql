-- drop unique index for type
drop index idx_models_type;

-- drop type for model
alter table models drop column type;
