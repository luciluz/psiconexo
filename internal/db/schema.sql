-- 1. PROFESIONALES
CREATE TABLE IF NOT EXISTS professionals (
    id BIGSERIAL PRIMARY KEY,  -- <-- CAMBIO: BIGSERIAL
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT UNIQUE,
    cancellation_window_hours INTEGER DEFAULT 24, -- Se queda en INTEGER (int32)
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. CLIENTES
CREATE TABLE IF NOT EXISTS clients (
    id BIGSERIAL PRIMARY KEY,  -- <-- CAMBIO
    name TEXT NOT NULL,
    professional_id BIGINT NOT NULL, -- <-- CAMBIO: BIGINT
    email TEXT,
    phone TEXT,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (professional_id) REFERENCES professionals(id)
);

-- 3. CONFIGURACIÓN DE AGENDA
CREATE TABLE IF NOT EXISTS schedule_configs (
    id BIGSERIAL PRIMARY KEY, -- <-- CAMBIO
    professional_id BIGINT NOT NULL, -- <-- CAMBIO
    day_of_week INTEGER NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (professional_id) REFERENCES professionals(id),
    UNIQUE(professional_id, day_of_week, start_time)
);

-- 4. REGLAS DE RECURRENCIA
CREATE TABLE IF NOT EXISTS recurring_rules (
    id BIGSERIAL PRIMARY KEY, -- <-- CAMBIO
    professional_id BIGINT NOT NULL, -- <-- CAMBIO
    client_id BIGINT NOT NULL, -- <-- CAMBIO
    day_of_week INTEGER NOT NULL,
    start_time TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL,
    
    price DECIMAL(10, 2) DEFAULT 0,
    
    active BOOLEAN DEFAULT TRUE,
    start_date DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    FOREIGN KEY (professional_id) REFERENCES professionals(id),
    FOREIGN KEY (client_id) REFERENCES clients(id)
);

-- 5. TURNOS
CREATE TABLE IF NOT EXISTS appointments (
    id BIGSERIAL PRIMARY KEY, -- <-- CAMBIO
    professional_id BIGINT NOT NULL, -- <-- CAMBIO
    client_id BIGINT NOT NULL, -- <-- CAMBIO
    date DATE NOT NULL,
    start_time TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL,
    
    price DECIMAL(10, 2) DEFAULT 0,
    
    status TEXT CHECK(status IN ('scheduled', 'cancelled', 'completed', 'rescheduled')) DEFAULT 'scheduled',
    rescheduled_from_id BIGINT, -- <-- CAMBIO
    recurring_rule_id BIGINT,   -- <-- CAMBIO

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    FOREIGN KEY (professional_id) REFERENCES professionals(id),
    FOREIGN KEY (client_id) REFERENCES clients(id),
    FOREIGN KEY (rescheduled_from_id) REFERENCES appointments(id),
    FOREIGN KEY (recurring_rule_id) REFERENCES recurring_rules(id),
    
    UNIQUE(professional_id, date, start_time)
);