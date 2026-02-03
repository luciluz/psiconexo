-- 1. PROFESIONALES
CREATE TABLE IF NOT EXISTS professionals (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT UNIQUE,

    slug TEXT UNIQUE,
    photo_url TEXT,
    title TEXT,
    license_number TEXT,
    bio TEXT,

    cancellation_window_hours INTEGER DEFAULT 24,
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. CONFIGURACIÓN AVANZADA DEL PROFESIONAL
CREATE TABLE IF NOT EXISTS professional_settings (
    professional_id BIGINT PRIMARY KEY,
    
    -- Reglas de Tiempo y Precios
    default_duration_minutes INTEGER DEFAULT 50,
    default_price DECIMAL(10,2) DEFAULT 0,
    buffer_minutes INTEGER DEFAULT 0,      
    time_increment_minutes INTEGER DEFAULT 30,

    -- Configuración de Cobros
    -- Transferencias
    bank_cbu TEXT,
    bank_alias TEXT,
    bank_name TEXT,
    bank_holder_name TEXT,
    send_alias_by_email BOOLEAN DEFAULT FALSE, -- Enviar alias al paciente por email

    -- MP (token encriptado)
    mp_access_token TEXT,
    mp_user_id TEXT,

    -- Facturación AFIP
    afip_crt_url TEXT, -- path al archivo certificado
    afip_key_url TEXT, -- path a la clave privada (encriptada)
    afip_point_of_sale INTEGER,

    notify_by_email BOOLEAN DEFAULT TRUE,
    notify_by_whatsapp BOOLEAN DEFAULT FALSE,
    
    -- Reglas de Límites
    min_booking_notice_hours INTEGER DEFAULT 24, 
    max_daily_appointments INTEGER,
    
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (professional_id) REFERENCES professionals(id)
);

-- 3. CLIENTES
CREATE TABLE IF NOT EXISTS clients (
    id BIGSERIAL PRIMARY KEY,
    professional_id BIGINT NOT NULL,

    name TEXT NOT NULL,
    email TEXT,
    phone TEXT,

    birth_date DATE,
    medications TEXT,
    emergency_contact_name TEXT,
    emergency_contact_phone TEXT,

    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (professional_id) REFERENCES professionals(id)
);

-- 4. CONFIGURACIÓN DE AGENDA
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

-- 5. REGLAS DE RECURRENCIA
CREATE TABLE IF NOT EXISTS recurring_rules (
    id BIGSERIAL PRIMARY KEY,
    professional_id BIGINT NOT NULL,
    client_id BIGINT NOT NULL,
    day_of_week INTEGER NOT NULL,
    start_time TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL,

    modality TEXT CHECK(modality IN ('virtual', 'in_person', 'home')) DEFAULT 'virtual',    
    price DECIMAL(10, 2) DEFAULT 0,    
    active BOOLEAN DEFAULT TRUE,
    start_date DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    FOREIGN KEY (professional_id) REFERENCES professionals(id),
    FOREIGN KEY (client_id) REFERENCES clients(id)
);

-- 6. TURNOS
CREATE TABLE IF NOT EXISTS appointments (
    id BIGSERIAL PRIMARY KEY,
    professional_id BIGINT NOT NULL,
    client_id BIGINT NOT NULL,

    date DATE NOT NULL,
    start_time TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL,
    
    status TEXT CHECK(status IN ('scheduled', 'cancelled', 'completed', 'rescheduled')) DEFAULT 'scheduled',
    
    modality TEXT CHECK(modality IN ('virtual', 'in_person', 'home')) DEFAULT 'virtual',
    meeting_url TEXT,

    -- Finanzas y Facturación
    price DECIMAL(10, 2) DEFAULT 0,
    concept TEXT DEFAULT 'Sesión de Terapia', -- esto será para AFIP

    payment_status TEXT CHECK(payment_status IN ('pending', 'paid', 'refunded')) DEFAULT 'pending',
    payment_method TEXT CHECK(payment_method IN ('mercadopago', 'transfer', 'cash', 'insurance')) DEFAULT 'transfer',
    payment_proof_url TEXT,
    payment_confirmed_at TIMESTAMPTZ, -- Cuándo se confirmó/aprobó el pago

    invoice_status TEXT CHECK(invoice_status IN ('pending', 'invoiced', 'error')) DEFAULT 'pending',
    invoice_url TEXT, -- pdf generado
    invoice_cae TEXT, -- este es un código de autorización de AFIP


    notes TEXT, -- este es un comentario simple, no las notas de la sesión (ej: paciente llega tarde)
    
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

-- 7. NOTAS CLÍNICAS (HISTORIA CLÍNICA)
CREATE TABLE IF NOT EXISTS clinical_notes (
    id BIGSERIAL PRIMARY KEY,
    professional_id BIGINT NOT NULL,
    client_id BIGINT NOT NULL,
    appointment_id BIGINT,

    type TEXT CHECK(type IN ('clinical', 'personal')) DEFAULT 'clinical',
    
    content TEXT NOT NULL, -- Encriptado
    key_version INTEGER DEFAULT 1,

    status TEXT CHECK(status IN ('draft', 'signed')) DEFAULT 'draft',

    signed_at TIMESTAMPTZ,    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    FOREIGN KEY (professional_id) REFERENCES professionals(id),
    FOREIGN KEY (client_id) REFERENCES clients(id),
    FOREIGN KEY (appointment_id) REFERENCES appointments(id)
);

-- ÍNDICES
CREATE INDEX IF NOT EXISTS idx_appointments_calendar ON appointments(professional_id, date);
CREATE INDEX IF NOT EXISTS idx_appointments_payment ON appointments(professional_id, payment_status);
CREATE INDEX IF NOT EXISTS idx_clients_professional ON clients(professional_id);
CREATE INDEX IF NOT EXISTS idx_notes_client ON clinical_notes(client_id);
CREATE INDEX IF NOT EXISTS idx_notes_status ON clinical_notes(professional_id, status);
CREATE INDEX IF NOT EXISTS idx_appointments_rule ON appointments(recurring_rule_id);