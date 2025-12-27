CREATE TABLE IF NOT EXISTS availability_slots (
	id UUID PRIMARY KEY DEFAULT uuidv7(),
	service_id UUID NOT NULL,
	day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
	start_time TIME NOT NULL,
	end_time TIME NOT NULL,
	max_bookings INT NOT NULL DEFAULT 1,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_availability_service FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS payments (
	id UUID PRIMARY KEY DEFAULT uuidv7(),
	booking_id UUID NOT NULL,
	amount DECIMAL(10, 2) NOT NULL,
	status VARCHAR(50) NOT NULL DEFAULT 'pending',
	payment_method VARCHAR(50),
	transaction_id VARCHAR(255),
	paid_at TIMESTAMP WITH TIME ZONE,
	refunded_at TIMESTAMP WITH TIME ZONE,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_payments_booking FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE
);

CREATE INDEX idx_availability_service_id ON availability_slots (service_id);
CREATE INDEX idx_availability_day_of_week ON availability_slots (day_of_week);
CREATE INDEX idx_availability_is_active ON availability_slots (is_active);

CREATE INDEX idx_payments_booking_id ON payments (booking_id);
CREATE INDEX idx_payments_status ON payments (status);
CREATE INDEX idx_payments_transaction_id ON payments (transaction_id);
CREATE INDEX idx_payments_created_at ON payments (created_at);
