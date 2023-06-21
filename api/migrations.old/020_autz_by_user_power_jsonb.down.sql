alter table "user" drop column power;

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

alter table can add column subject text references subject (name);
alter table can add constraint can_subject_action_target unique (subject, action, target);

insert into subject values
('visitor'), ('company.admin'), ('all.admin');

insert into action values
('read'), ('write'), ('create'), ('delete'), ('delete.soft');

insert into subject_action values
('company.admin', 'read'),
('company.admin', 'write'),
('company.admin', 'create'),
('company.admin', 'delete.soft');

insert into user_subject values
(1, 'company.admin'),
(1, 'all.admin'),
(2, 'company.admin');

insert into can (subject, action, target) values
('company.admin', 'read', 'own'),
('company.admin', 'write', 'own'),
('company.admin', 'create', 'own'),
('company.admin', 'delete.soft', 'own');
