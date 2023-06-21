alter table deed add column deleted_at timestamptz null; 

alter table drain add column is_deleted boolean not null default false;
alter table drain drop constraint drain_deed_id_fk_deed_id; 
alter table drain add constraint drain_deed_id_fk_deed_id foreign key (deed_id) references deed(id) on update cascade;

create or replace function update_drain_is_deleted() returns trigger as $$
declare var_is_deleted boolean;
begin
  if (NEW.deleted_at is null) then
   var_is_deleted := false;
  else
    var_is_deleted := true;
  end if;
  update drain set is_deleted = var_is_deleted where deed_id = NEW.id;
  return NEW;
end;
$$ language plpgsql;

create trigger update_drain_is_deleted_trigger
after update on deed
for each row
when (OLD.deleted_at is distinct from NEW.deleted_at)
execute function update_drain_is_deleted();
