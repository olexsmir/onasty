alter table notes
add column dont_delete_before_expiration boolean default false;

---- create above / drop below ----

alter table notes drop column dont_delete_before_expiration ;
