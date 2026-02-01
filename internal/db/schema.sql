-- 1. PSICÓLOGOS
CREATE TABLE IF NOT EXISTS psychologists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT UNIQUE,
    cancellation_window_hours INTEGER DEFAULT 24,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 2. PACIENTES
CREATE TABLE IF NOT EXISTS patients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    psychologist_id INTEGER NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT UNIQUE,
    active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (psychologist_id) REFERENCES psychologists(id)
);

-- 3. CONFIGURACIÓN DE DISPONIBILIDAD
CREATE TABLE IF NOT EXISTS schedule_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    psychologist_id INTEGER NOT NULL,
    day_of_week INTEGER NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (psychologist_id) REFERENCES psychologists(id),
    UNIQUE(psychologist_id, day_of_week, start_time)
);

-- 4. REGLAS DE RECURRENCIA (Antes recurring_slots)
-- Esta tabla NO reserva el turno, solo guarda la instrucción de "generar turnos".
CREATE TABLE IF NOT EXISTS recurring_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    psychologist_id INTEGER NOT NULL,
    patient_id INTEGER NOT NULL,
    day_of_week INTEGER NOT NULL,     -- Ej: 1 (Lunes)
    start_time TEXT NOT NULL,         -- Ej: "10:00"
    duration_minutes INTEGER NOT NULL,
    
    active BOOLEAN DEFAULT TRUE,      -- Permite pausar la regla sin borrarla
    start_date DATE,                  -- Opcional: Desde cuándo aplica la regla
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (psychologist_id) REFERENCES psychologists(id),
    FOREIGN KEY (patient_id) REFERENCES patients(id),
    
    -- Evitamos crear dos reglas idénticas para el mismo profesional
    UNIQUE(psychologist_id, day_of_week, start_time)
);

-- 5. TURNOS
-- Aquí viven TANTO los puntuales COMO las instancias generadas de los fijos.
CREATE TABLE IF NOT EXISTS appointments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    psychologist_id INTEGER NOT NULL,
    patient_id INTEGER NOT NULL,
    date DATE NOT NULL,
    start_time TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL,
    
    -- Agregamos 'rescheduled' que es un estado útil
    status TEXT CHECK(status IN ('scheduled', 'cancelled', 'completed', 'rescheduled')) DEFAULT 'scheduled',
    
    rescheduled_from_id INTEGER,
    
    -- NUEVO CAMPO: ¿Vino de una regla fija?
    -- Si es NULL = Turno puntual eventual.
    -- Si tiene ID = Es la instancia concreta de una regla fija.
    recurring_rule_id INTEGER, 

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (psychologist_id) REFERENCES psychologists(id),
    FOREIGN KEY (patient_id) REFERENCES patients(id),
    FOREIGN KEY (rescheduled_from_id) REFERENCES appointments(id),
    FOREIGN KEY (recurring_rule_id) REFERENCES recurring_rules(id),
    
    -- LA GRAN BARRERA:
    -- Esto impide que existan dos turnos a la misma hora y fecha para el mismo psicólogo.
    UNIQUE(psychologist_id, date, start_time)
);

-- ÍNDICES
CREATE INDEX IF NOT EXISTS idx_appointments_calendar 
ON appointments(psychologist_id, date);

CREATE INDEX IF NOT EXISTS idx_patients_psychologist 
ON patients(psychologist_id);

-- Índice para encontrar rápido los turnos generados por una regla específica
CREATE INDEX IF NOT EXISTS idx_appointments_rule 
ON appointments(recurring_rule_id);