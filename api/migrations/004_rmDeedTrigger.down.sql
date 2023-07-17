create or replace trigger update_drain_is_deleted_trigger after
update
    on
    api.deed for each row
    when ((old.deleted_at is distinct
from
    new.deleted_at)) execute function update_drain_is_deleted();
