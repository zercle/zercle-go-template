CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	email VARCHAR(255) NOT NULL UNIQUE,
	password_hash VARCHAR(255) NOT NULL,
	full_name VARCHAR(255) NOT NULL,
	phone VARCHAR(50),
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS services (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	name VARCHAR(255) NOT NULL,
	description TEXT,
	duration_minutes INT NOT NULL,
	price DECIMAL(10, 2) NOT NULL,
	max_capacity INT NOT NULL DEFAULT 1,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bookings (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	user_id UUID NOT NULL,
	service_id UUID NOT NULL,
	start_time TIMESTAMP WITH TIME ZONE NOT NULL,
	end_time TIMESTAMP WITH TIME ZONE NOT NULL,
	status VARCHAR(50) NOT NULL DEFAULT 'pending',
	total_price DECIMAL(10, 2) NOT NULL,
	notes TEXT,
	cancelled_at TIMESTAMP WITH TIME ZONE,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_bookings_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT fk_bookings_service FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS bookings_users (
	booking_id UUID NOT NULL,
	user_id UUID NOT NULL,
	is_primary BOOLEAN NOT NULL DEFAULT FALSE,
	PRIMARY KEY (booking_id, user_id),
	CONSTRAINT fk_bookings_users_booking FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE,
	CONSTRAINT fk_bookings_users_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_created_at ON users (created_at);

CREATE INDEX idx_services_name ON services (name);
CREATE INDEX idx_services_is_active ON services (is_active);
CREATE INDEX idx_services_created_at ON services (created_at);

CREATE INDEX idx_bookings_user_id ON bookings (user_id);
CREATE INDEX idx_bookings_service_id ON bookings (service_id);
CREATE INDEX idx_bookings_status ON bookings (status);
CREATE INDEX idx_bookings_start_time ON bookings (start_time);
CREATE INDEX idx_bookings_end_time ON bookings (end_time);
CREATE INDEX idx_bookings_created_at ON bookings (created_at);

CREATE INDEX idx_bookings_users_user_id ON bookings_users (user_id);
CREATE INDEX idx_bookings_users_booking_id ON bookings_users (booking_id);
