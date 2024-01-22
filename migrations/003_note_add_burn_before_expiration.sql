alter table notes
add column burn_before_expiration boolean default false;

---- create above / drop below ----

alter table notes drop column burn_before_expiration;
