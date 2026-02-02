-- 1. PROFESIONALES
CREATE TABLE IF NOT EXISTS professionals (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT UNIQUE,
    cancellation_window_hours INTEGER DEFAULT 24,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- CONFIGURACIÓN AVANZADA DEL PROFESIONAL
CREATE TABLE IF NOT EXISTS professional_settings (
    professional_id BIGINT PRIMARY KEY, -- 1 a 1 con professionals
    
    -- Reglas de Tiempo
    default_duration_minutes INTEGER DEFAULT 50,
    buffer_minutes INTEGER DEFAULT 0,      -- Tiempo muerto entre sesiones
    time_increment_minutes INTEGER DEFAULT 30, -- 60 = Modo Tetris, 15 = Flexible
    
    -- Reglas de Límites
    min_booking_notice_hours INTEGER DEFAULT 24, -- "No me agendes para hoy"
    max_daily_appointments INTEGER,              -- NULL = Sin límite
    
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (professional_id) REFERENCES professionals(id)
);

-- 2. CLIENTES
CREATE TABLE IF NOT EXISTS clients (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    professional_id BIGINT NOT NULL,
    email TEXT,
    phone TEXT,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (professional_id) REFERENCES professionals(id)
);

-- 3. CONFIGURACIÓN DE AGENDA
CREATE TABLE IF NOT EXISTS schedule_configs (
    id BIGSERIAL PRIMARY KEY,
    professional_id BIGINT NOT NULL,
    day_of_week INTEGER NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (professional_id) REFERENCES professionals(id),
    UNIQUE(professional_id, day_of_week, start_time)
);

-- 4. REGLAS DE RECURRENCIA
CREATE TABLE IF NOT EXISTS recurring_rules (
    id BIGSERIAL PRIMARY KEY,
    professional_id BIGINT NOT NULL,
    client_id BIGINT NOT NULL,
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
    id BIGSERIAL PRIMARY KEY,
    professional_id BIGINT NOT NULL,
    client_id BIGINT NOT NULL,
    date DATE NOT NULL,
    start_time TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL,
    notes TEXT,
    
    price DECIMAL(10, 2) DEFAULT 0,
    
    status TEXT CHECK(status IN ('scheduled', 'cancelled', 'completed', 'rescheduled')) DEFAULT 'scheduled',
    rescheduled_from_id BIGINT,
    recurring_rule_id BIGINT,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    FOREIGN KEY (professional_id) REFERENCES professionals(id),
    FOREIGN KEY (client_id) REFERENCES clients(id),
    FOREIGN KEY (rescheduled_from_id) REFERENCES appointments(id),
    FOREIGN KEY (recurring_rule_id) REFERENCES recurring_rules(id),
    
    UNIQUE(professional_id, date, start_time)
);

-- NOTAS CLÍNICAS (HISTORIA CLÍNICA)
CREATE TABLE IF NOT EXISTS clinical_notes (
    id BIGSERIAL PRIMARY KEY,
    professional_id BIGINT NOT NULL,
    client_id BIGINT NOT NULL,
    appointment_id BIGINT, -- Opcional: Vincular nota a una sesión específica
    
    -- Aquí guardamos el texto encriptado.
    -- IMPORTANTE: No usamos JSONB porque al estar encriptado es solo un string largo opaco.
    content TEXT NOT NULL, 
    
    -- Metadata para saber si está encriptado (útil para migraciones futuras)
    is_encrypted BOOLEAN DEFAULT TRUE, 
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    FOREIGN KEY (professional_id) REFERENCES professionals(id),
    FOREIGN KEY (client_id) REFERENCES clients(id),
    FOREIGN KEY (appointment_id) REFERENCES appointments(id)
);

-- ÍNDICES
CREATE INDEX IF NOT EXISTS idx_appointments_calendar ON appointments(professional_id, date);
CREATE INDEX IF NOT EXISTS idx_clients_professional ON clients(professional_id);
CREATE INDEX IF NOT EXISTS idx_appointments_rule ON appointments(recurring_rule_id);
CREATE INDEX IF NOT EXISTS idx_notes_client ON clinical_notes(client_id);