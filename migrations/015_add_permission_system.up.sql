create table subject (
  name text not null unique
);

create table action (
  name text not null unique
);

create table can (
  action text references action (name),
  target text
);

alter table can add constraint can_action_target unique(action, target);

create table subject_action (
  subject text references subject (name),
  action text references action (name)
);

alter table subject_action add constraint subject_action_subject_action unique(subject, action);

create table user_subject (
  user_id int references "user" (id),
  subject text references subject (name)
);

alter table user_subject add constraint user_subject_user_id_subject unique(user_id, subject);
