ALTER TABLE products
ADD COLUMN workspace_id INT,
ADD CONSTRAINT fk_workspace
FOREIGN KEY (workspace_id) REFERENCES workspaces(id); 