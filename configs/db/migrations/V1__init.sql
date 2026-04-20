-- Auth DB — Extensiones y tipos base
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE user_role AS ENUM ('admin', 'driver', 'guardian', 'school_staff');
