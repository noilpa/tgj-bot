ALTER TABLE mrs ADD COLUMN need_jira_update bool DEFAULT TRUE;

UPDATE mrs SET need_jira_update=FALSE WHERE gitlab_id=0 AND is_closed=TRUE;