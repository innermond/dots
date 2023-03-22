alter table subject_action drop constraint subject_action_subject_action;
alter table user_subject drop constraint user_subject_user_id_subject;
drop table subject_action, user_subject, subject, action, can;
