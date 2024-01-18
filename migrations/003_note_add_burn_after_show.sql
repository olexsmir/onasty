alter table notes
add column burn_after_show boolean default false;

---- create above / drop below ----

alter table notes drop column burn_after_show;
