CREATE TABLE IF NOT EXISTS tasks (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL,
	title VARCHAR(255) NOT NULL,
	description TEXT,
	status VARCHAR(50) NOT NULL DEFAULT 'pending',
	priority VARCHAR(20) NOT NULL DEFAULT 'medium',
	due_date TIMESTAMP WITH TIME ZONE,
	completed_at TIMESTAMP WITH TIME ZONE,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_tasks_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_tasks_user_id ON tasks (user_id);
CREATE INDEX idx_tasks_status ON tasks (status);
CREATE INDEX idx_tasks_priority ON tasks (priority);
CREATE INDEX idx_tasks_created_at ON tasks (created_at);
CREATE INDEX idx_tasks_due_date ON tasks (due_date);
