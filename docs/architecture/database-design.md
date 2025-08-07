# Database Entity Relationship Diagram (ERD)

This document contains the Entity Relationship Diagram for the Go RESTful API database schema using Mermaid syntax.

## Complete Database Schema

```mermaid
erDiagram
    %% Core User Management
    users {
        UUID id PK
        VARCHAR username UK
        VARCHAR email UK
        VARCHAR phone UK
        VARCHAR password
        TIMESTAMP email_verified_at
        TIMESTAMP phone_verified_at
        TIMESTAMP created_at
        TIMESTAMP updated_at
        TIMESTAMP deleted_at
    }

    roles {
        UUID id PK
        VARCHAR name
        VARCHAR slug UK
        TEXT description
        BOOLEAN is_active
        TIMESTAMP created_at
        TIMESTAMP updated_at
        TIMESTAMP deleted_at
        VARCHAR created_by
        VARCHAR updated_by
        VARCHAR deleted_by
    }

    user_roles {
        UUID id PK
        UUID user_id FK
        UUID role_id FK
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    %% Content Management
    posts {
        UUID id PK
        VARCHAR slug UK
        VARCHAR title
        VARCHAR subtitle
        TEXT description
        VARCHAR type
        BOOLEAN is_sticky
        BIGINT published_at
        VARCHAR language
        VARCHAR layout
        INT record_ordering
        UUID created_by FK
        UUID updated_by FK
        UUID deleted_by FK
        TIMESTAMP deleted_at
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    contents {
        UUID id PK
        VARCHAR model_type
        UUID model_id
        TEXT content_raw
        TEXT content_html
        UUID created_by FK
        UUID updated_by FK
        UUID deleted_by FK
        TIMESTAMP deleted_at
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    comments {
        UUID id PK
        VARCHAR model_type
        UUID model_id
        VARCHAR title
        VARCHAR status
        UUID parent_id FK
        INT record_left
        INT record_right
        INT record_ordering
        UUID created_by FK
        UUID updated_by FK
        UUID deleted_by FK
        TIMESTAMP deleted_at
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    %% Taxonomy and Categorization
    taxonomies {
        UUID id PK
        VARCHAR name
        VARCHAR slug UK
        VARCHAR code
        TEXT description
        INT record_left
        INT record_right
        INT record_ordering
        UUID parent_id FK
        UUID created_by FK
        UUID updated_by FK
        UUID deleted_by FK
        TIMESTAMP deleted_at
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    tags {
        UUID id PK
        VARCHAR name
        VARCHAR slug UK
        UUID created_by FK
        UUID updated_by FK
        UUID deleted_by FK
        TIMESTAMP deleted_at
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    taggables {
        UUID id PK
        UUID tag_id FK
        UUID taggable_id
        VARCHAR taggable_type
        BIGINT created_at
    }

    %% Media Management
    media {
        UUID id PK
        VARCHAR name
        VARCHAR hash
        VARCHAR file_name
        VARCHAR disk
        VARCHAR mime_type
        INT size
        INT record_left
        INT record_right
        INT record_depth
        INT record_ordering
        UUID parent_id FK
        VARCHAR custom_attributes
        UUID created_by FK
        UUID updated_by FK
        UUID deleted_by FK
        TIMESTAMP deleted_at
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    mediables {
        UUID media_id FK
        UUID mediable_id
        VARCHAR mediable_type
        VARCHAR group
    }

    %% Menu System
    menus {
        UUID id PK
        VARCHAR name
        VARCHAR slug UK
        TEXT description
        VARCHAR url
        VARCHAR icon
        UUID parent_id FK
        BIGINT record_left
        BIGINT record_right
        BIGINT record_ordering
        BOOLEAN is_active
        BOOLEAN is_visible
        VARCHAR target
        TIMESTAMP created_at
        TIMESTAMP updated_at
        TIMESTAMP deleted_at
        VARCHAR created_by
        VARCHAR updated_by
        VARCHAR deleted_by
    }

    menu_roles {
        UUID id PK
        UUID menu_id FK
        UUID role_id FK
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    %% Settings and Configuration
    settings {
        UUID id PK
        VARCHAR model_type
        UUID model_id
        VARCHAR key
        VARCHAR value
        UUID created_by FK
        UUID updated_by FK
        UUID deleted_by FK
        TIMESTAMP deleted_at
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    %% Job Queue System
    jobs {
        UUID id PK
        VARCHAR queue
        VARCHAR handler_name
        JSONB payload
        INT max_attempts
        INT delay
        VARCHAR status
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    failed_jobs {
        SERIAL id PK
        UUID job_id UK
        VARCHAR queue
        JSONB payload
        TEXT error
        TIMESTAMP failed_at
    }

    %% Relationships
    users ||--o{ user_roles : "has"
    roles ||--o{ user_roles : "assigned_to"
    users ||--o{ posts : "creates"
    users ||--o{ posts : "updates"
    users ||--o{ posts : "deletes"
    users ||--o{ contents : "creates"
    users ||--o{ contents : "updates"
    users ||--o{ contents : "deletes"
    users ||--o{ comments : "creates"
    users ||--o{ comments : "updates"
    users ||--o{ comments : "deletes"
    users ||--o{ taxonomies : "creates"
    users ||--o{ taxonomies : "updates"
    users ||--o{ taxonomies : "deletes"
    users ||--o{ tags : "creates"
    users ||--o{ tags : "updates"
    users ||--o{ tags : "deletes"
    users ||--o{ media : "creates"
    users ||--o{ media : "updates"
    users ||--o{ media : "deletes"
    users ||--o{ settings : "creates"
    users ||--o{ settings : "updates"
    users ||--o{ settings : "deletes"

    posts ||--o{ contents : "has"
    posts ||--o{ comments : "has"
    taxonomies ||--o{ taxonomies : "parent_child"
    menus ||--o{ menus : "parent_child"
    media ||--o{ media : "parent_child"
    comments ||--o{ comments : "parent_child"

    tags ||--o{ taggables : "used_in"
    media ||--o{ mediables : "attached_to"

    menus ||--o{ menu_roles : "accessible_by"
    roles ||--o{ menu_roles : "has_access_to"

    jobs ||--o{ failed_jobs : "may_fail"
```

## Key Features

### 1. **User Management System**
- **users**: Core user accounts with authentication fields
- **roles**: Role definitions for RBAC (Role-Based Access Control)
- **user_roles**: Many-to-many relationship between users and roles

### 2. **Content Management System**
- **posts**: Blog posts and articles with polymorphic content
- **contents**: Rich content storage with raw and HTML versions
- **comments**: Hierarchical comment system with nested comments support

### 3. **Taxonomy System**
- **taxonomies**: Hierarchical categorization (categories, etc.)
- **tags**: Simple tagging system
- **taggables**: Polymorphic tagging for any entity

### 4. **Media Management**
- **media**: File storage with hierarchical organization
- **mediables**: Polymorphic media attachments

### 5. **Menu System**
- **menus**: Hierarchical navigation menus
- **menu_roles**: Role-based menu access control

### 6. **Configuration**
- **settings**: Flexible key-value configuration system

### 7. **Job Queue**
- **jobs**: Background job processing
- **failed_jobs**: Failed job tracking

## Design Patterns

### 1. **Polymorphic Associations**
- `contents` table uses `model_type` and `model_id` for polymorphic relationships
- `comments` table supports comments on any entity
- `taggables` table enables tagging of any entity
- `mediables` table allows media attachments to any entity
- `settings` table supports entity-specific settings

### 2. **Hierarchical Data**
- **Nested Set Model**: Used in `taxonomies`, `menus`, and `media` tables with `record_left`, `record_right`, and `record_ordering` fields
- **Adjacency List**: Used in `comments` table with `parent_id` field

### 3. **Audit Trail**
- All major tables include `created_by`, `updated_by`, `deleted_by` fields
- Soft deletes with `deleted_at` timestamps
- Standard `created_at` and `updated_at` timestamps

### 4. **RBAC Implementation**
- Role-based access control through `roles`, `user_roles`, and `menu_roles` tables
- Menu access control based on user roles

## Database Constraints

### Primary Keys
- All tables use UUID primary keys for scalability and security

### Foreign Keys
- Proper foreign key constraints with appropriate cascade rules
- User references use `ON DELETE SET NULL` for audit trail preservation

### Unique Constraints
- Email and username uniqueness in users table
- Slug uniqueness in content tables
- Composite unique constraints for many-to-many relationships

### Check Constraints
- Menu table has constraints for nested set model integrity
- Status field constraints in jobs table

## Indexes
- Queue and status indexes on jobs table for performance
- Queue index on failed_jobs table
- Implicit indexes on primary keys and foreign keys

This ERD represents a comprehensive content management system with user management, role-based access control, hierarchical data structures, and background job processing capabilities. 