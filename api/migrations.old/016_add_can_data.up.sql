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
