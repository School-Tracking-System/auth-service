# Auth Service

Microservicio de autenticación del School Tracking System. Gestiona el registro de usuarios, inicio de sesión y emisión/renovación de tokens JWT.

## Arquitectura

El servicio sigue **Arquitectura Hexagonal** (Ports & Adapters):

```
services/auth/
├── cmd/api/              # Punto de entrada y wiring con Uber fx
├── configs/db/migrations # Migraciones Flyway (SQL)
├── docs/api/             # Swagger generado por swag
├── internal/
│   ├── core/
│   │   ├── auth/         # Implementación del servicio y JWT manager
│   │   ├── domain/       # Modelos de dominio (User, Token, Claims)
│   │   └── ports/        # Interfaces (AuthService, UserRepository, JWTManager)
│   └── infrastructure/
│       ├── api/          # Adaptador HTTP REST: controllers, DTOs, errores, routes
│       ├── persistence/  # Repositorio PostgreSQL con GORM
│       └── messaging/    # Publisher NATS (pendiente)
└── pkg/
    ├── env/              # Configuración por variables de entorno
    └── logger/           # Logger estructurado con Zap
```

> **Convención de transporte:** Este servicio expone HTTP REST, por eso la carpeta de transporte se llama `infrastructure/api/`. Si en el futuro se expone gRPC, se añadirá `infrastructure/grpc/` en paralelo. Ver `docs/plans/backend/00-backend-overview.md §2`.

## Stack

| Componente       | Tecnología                |
|------------------|---------------------------|
| Lenguaje         | Go 1.23                   |
| HTTP Router      | go-chi/chi v5             |
| Base de datos    | PostgreSQL 16             |
| ORM              | GORM                      |
| Migraciones      | Flyway 10                 |
| Autenticación    | JWT (golang-jwt/jwt v5)   |
| DI               | Uber fx                   |
| Logger           | Zap                       |
| Config           | caarlos0/env v10          |
| API Docs         | swaggo/swag               |

## Requisitos previos

- **Go** >= 1.23
- **Docker** y **Docker Compose** (para PostgreSQL, Redis, NATS)
- **swag** CLI (opcional, para regenerar docs Swagger)

## Variables de entorno

| Variable       | Default                                                                  | Descripción                    |
|----------------|--------------------------------------------------------------------------|--------------------------------|
| `SERVICE_NAME` | `auth`                                                                   | Nombre del servicio            |
| `HTTP_PORT`    | `8080`                                                                   | Puerto del servidor HTTP       |
| `GRPC_PORT`    | `9090`                                                                   | Puerto gRPC (futuro)           |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/school_tracking?sslmode=disable` | Conexión a PostgreSQL     |
| `REDIS_URL`    | `redis://localhost:6379`                                                 | Conexión a Redis (futuro)      |
| `NATS_URL`     | `nats://localhost:4222`                                                  | Conexión a NATS (futuro)       |
| `JWT_SECRET`   | `dev-secret-change-in-prod`                                              | Secreto para firmar JWT        |
| `ENVIRONMENT`  | `development`                                                            | Entorno (development/production) |
| `LOG_LEVEL`    | `debug`                                                                  | Nivel de log (debug/info/warn/error) |

> **Importante**: En producción, cambiar `JWT_SECRET` y `DATABASE_URL` con valores seguros.

### Archivo .env

El servicio carga automáticamente un archivo `.env` si existe en la raíz del servicio (`services/auth/.env`). Las variables del sistema siempre tienen prioridad sobre las del archivo.

Para configurar el entorno local:

```bash
cd services/auth
cp .env.template .env
# Editar .env con los valores deseados
```

- **`.env`** — No se sube a Git (ignorado en `.gitignore`). Contiene valores reales.
- **`.env.template`** — Se sube a Git como referencia. Contiene descripciones y valores por defecto seguros.

## Levantar el servicio

### 1. Iniciar infraestructura con Docker

Desde la raíz del monorepo:

```bash
docker compose up -d postgres redis nats
```

### 2. Ejecutar migraciones

```bash
docker compose up flyway-auth
```

Esto crea la tabla `users` con sus índices y triggers en la base de datos `school_tracking`.

### 3. Compilar y ejecutar

```bash
cd services/auth
go build -o bin/api ./cmd/api/
./bin/api
```

El servidor arrancará en `http://localhost:8080`.

## Endpoints

Base path: `/api/v1/auth`

### POST /api/v1/auth/register

Registra un nuevo usuario.

**Validaciones:**

| Campo | Obligatorio | Regla |
|---|---|---|
| `email` | Sí | Formato email válido |
| `password` | Sí | Mínimo 6 caracteres |
| `first_name` | Sí | — |
| `last_name` | Sí | — |
| `phone` | No | — |
| `role` | Sí | `admin` \| `driver` \| `guardian` \| `school_staff` |

**Request:**
```json
{
  "email": "usuario@example.com",
  "password": "MiPassword123!",
  "first_name": "Fernando",
  "last_name": "Garcia",
  "phone": "+34600000000",
  "role": "admin"
}
```

**Response (201 Created):**
```json
{
  "id": "07de5ca9-66fd-4756-84c6-93d28f78cf7d",
  "email": "usuario@example.com",
  "first_name": "Fernando",
  "last_name": "Garcia",
  "phone": "+34600000000",
  "role": "admin",
  "is_active": true
}
```

**Error Response Ejemplo (400 Bad Request por Validación):**
```json
{
  "code": 400,
  "type": "invalid_request",
  "message": "validation failed for one or more fields",
  "details": [
    "field 'Password' failed validation on 'min' tag"
  ]
}
```

### POST /api/v1/auth/login

Autentica un usuario y devuelve tokens JWT.

**Validaciones:**

| Campo | Obligatorio | Regla |
|---|---|---|
| `email` | Sí | Formato email válido |
| `password` | Sí | — |

**Request:**
```json
{
  "email": "usuario@example.com",
  "password": "MiPassword123!"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Error Response Ejemplo (401 Unauthorized):**
```json
{
  "code": 401,
  "type": "invalid_request",
  "message": "invalid email or password"
}
```

### POST /api/v1/auth/refresh

Renueva los tokens usando un refresh token válido.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

## Roles disponibles

| Valor          | Descripción                |
|----------------|----------------------------|
| `admin`        | Administrador del sistema  |
| `driver`       | Conductor de ruta escolar  |
| `guardian`     | Padre/madre/tutor          |
| `school_staff` | Personal de la escuela     |

## Swagger UI

Con el servicio corriendo, acceder a:

```
http://localhost:8080/swagger/index.html
```

Para regenerar la documentación Swagger:

```bash
cd services/auth
swag init -g cmd/api/main.go -d cmd/api,internal/infrastructure/api/controllers,internal/infrastructure/api/dtos,internal/infrastructure/api/errors -o docs/api --parseDependency --parseInternal
```

## Base de datos

La tabla `users` se gestiona mediante migraciones Flyway ubicadas en `configs/db/migrations/`.

**Schema:**

| Columna        | Tipo           | Restricciones              |
|----------------|----------------|----------------------------|
| `id`           | UUID           | PK, default uuid_generate_v4() |
| `email`        | VARCHAR(255)   | UNIQUE, NOT NULL           |
| `phone`        | VARCHAR(20)    | Nullable                   |
| `password_hash`| TEXT           | NOT NULL                   |
| `role`         | user_role ENUM | NOT NULL, default 'guardian' |
| `first_name`   | VARCHAR(100)   | NOT NULL                   |
| `last_name`    | VARCHAR(100)   | NOT NULL                   |
| `fcm_token`    | TEXT           | Nullable                   |
| `is_active`    | BOOLEAN        | NOT NULL, default true     |
| `created_at`   | TIMESTAMPTZ    | NOT NULL, default NOW()    |
| `updated_at`   | TIMESTAMPTZ    | NOT NULL, auto-updated via trigger |