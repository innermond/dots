drop trigger update_drain_is_deleted_trigger on deed;
drop function update_drain_is_deleted;

alter table drain drop constraint drain_deed_id_fk_deed_id; 
alter table drain add constraint drain_deed_id_fk_deed_id foreign key (deed_id) references deed(id);
alter table drain drop column is_deleted;

alter table deed drop column deleted_at;
