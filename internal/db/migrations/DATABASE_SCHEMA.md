# Database Schema Documentation

## Overview

This document provides a comprehensive overview of the database schema for the Go RESTful application. The database is designed using PostgreSQL with a focus on scalability, maintainability, and performance.

## Database Design Patterns

### 1. **Nested Set Model (NSM)**
Several tables implement the Nested Set Model for efficient hierarchical data management:
- **organizations**: Hierarchical organization structure
- **taxonomies**: Content classification hierarchy
- **menus**: Navigation menu hierarchy
- **media**: File organization hierarchy
- **comments**: Threaded comment hierarchy

**Benefits:**
- Efficient tree traversal queries
- Fast insertion/deletion operations
- Optimized for read-heavy workloads

**Fields:**
- `record_left`: Left boundary value
- `record_right`: Right boundary value
- `record_depth`: Depth level in hierarchy
- `record_ordering`: Display order within same level

### 2. **Polymorphic Associations**
The system uses polymorphic relationships for flexible content management:
- **comments**: Can be attached to any entity type
- **contents**: Store content for various model types
- **settings**: Entity-specific configuration
- **addresses**: Associated with users or organizations

**Implementation:**
- `model_type`: String identifier for entity type
- `model_id`: UUID reference to the specific entity

### 3. **Soft Delete Pattern**
All major entities implement soft delete functionality:
- `deleted_at`: Timestamp when record was soft deleted
- `deleted_by`: User who performed the deletion
- Records remain in database for audit purposes

### 4. **Audit Trail**
Comprehensive audit trail across all entities:
- `created_at`: Record creation timestamp
- `updated_at`: Last modification timestamp
- `created_by`: User who created the record
- `updated_by`: User who last modified the record

## Table Structure

### Core User Management

#### `users`
**Purpose:** Central user account management
**Key Features:**
- Unique username and email constraints
- Phone number verification support
- Email verification tracking
- Password storage (hashed)

**Fields:**
- `id`: Primary key (UUID)
- `username`: Unique username for login
- `email`: Unique email address
- `phone`: Optional phone number
- `password`: Hashed password (excluded from JSON)
- `email_verified_at`: Email verification timestamp
- `phone_verified_at`: Phone verification timestamp
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

**Indexes:**
- Primary key on `id`
- Unique constraint on `username`
- Unique constraint on `email`
- Unique constraint on `phone`

#### `roles`
**Purpose:** Role-based access control (RBAC)
**Key Features:**
- Predefined system roles
- Active/inactive status management
- Soft delete support

**Fields:**
- `id`: Primary key (UUID)
- `name`: Role display name
- `slug`: URL-friendly identifier
- `description`: Role purpose description
- `is_active`: Active status flag
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

**Predefined Roles:**
- Admin, User, Moderator, Editor, Viewer

#### `role_entities`
**Purpose:** Many-to-many relationship between users and roles
**Key Features:**
- Junction table for user-role assignments
- Cascade delete for referential integrity

**Fields:**
- `id`: Primary key (UUID)
- `entity_id`: User ID reference
- `entity_type`: Entity type (currently only 'user')
- `role_id`: Role ID reference
- `created_at`, `updated_at`: Audit timestamps

### Content Management

#### `posts`
**Purpose:** Blog posts and articles
**Key Features:**
- SEO-friendly slugs
- Publishing workflow support
- Sticky post functionality
- Multi-language support

**Fields:**
- `id`: Primary key (UUID)
- `slug`: URL-friendly identifier (unique)
- `title`: Post title
- `subtitle`: Optional subtitle
- `description`: Post summary
- `type`: Content type
- `is_sticky`: Pinned post flag
- `published_at`: Publication timestamp
- `language`: Content language code
- `layout`: Template layout identifier
- `record_ordering`: Display order
- `created_by`, `updated_by`, `deleted_by`: User references
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

#### `contents`
**Purpose:** Store raw and formatted content
**Key Features:**
- Dual content storage (raw + HTML)
- Polymorphic associations
- Content versioning support

**Fields:**
- `id`: Primary key (UUID)
- `model_type`: Associated entity type
- `model_id`: Associated entity ID
- `content_raw`: Raw content (markdown, plain text)
- `content_html`: HTML-formatted content
- `created_by`, `updated_by`, `deleted_by`: User references
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

#### `taxonomies`
**Purpose:** Content classification and categorization
**Key Features:**
- Hierarchical structure (Nested Set Model)
- Flexible classification system
- SEO-friendly slugs

**Fields:**
- `id`: Primary key (UUID)
- `name`: Taxonomy name
- `slug`: URL-friendly identifier (unique)
- `code`: Optional taxonomy code
- `description`: Taxonomy description
- `parent_id`: Parent taxonomy reference
- `record_left`, `record_right`, `record_depth`, `record_ordering`: NSM fields
- `created_by`, `updated_by`, `deleted_by`: User references
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

#### `tags`
**Purpose:** Content tagging and labeling
**Key Features:**
- Visual customization with colors
- SEO-friendly slugs
- Polymorphic associations

**Fields:**
- `id`: Primary key (UUID)
- `name`: Tag name
- `slug`: URL-friendly identifier (unique)
- `description`: Tag purpose description
- `color`: Visual representation color
- `created_by`, `updated_by`, `deleted_by`: User references
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

#### `taggables`
**Purpose:** Many-to-many relationship between tags and content
**Key Features:**
- Junction table for tag-content associations
- Polymorphic support for different content types

**Fields:**
- `id`: Primary key (UUID)
- `tag_id`: Tag reference
- `taggable_id`: Content entity ID
- `taggable_type`: Content entity type
- `created_at`: Association timestamp

### Media Management

#### `media`
**Purpose:** File and media asset management
**Key Features:**
- Hierarchical organization (Nested Set Model)
- File metadata storage
- Polymorphic associations
- Storage disk management

**Fields:**
- `id`: Primary key (UUID)
- `name`: Display name
- `hash`: File integrity hash
- `file_name`: Original filename
- `disk`: Storage disk identifier
- `mime_type`: File MIME type
- `size`: File size in bytes
- `parent_id`: Parent media reference
- `custom_attributes`: Additional file attributes
- `record_left`, `record_right`, `record_depth`, `record_ordering`: NSM fields
- `created_by`, `updated_by`, `deleted_by`: User references
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

**Indexes:**
- NSM operation indexes
- Parent relationship index
- File type and size indexes

#### `mediables`
**Purpose:** Many-to-many relationship between media and content
**Key Features:**
- Junction table for media-content associations
- Polymorphic support for different content types
- Grouping functionality

**Fields:**
- `media_id`: Media reference
- `mediable_id`: Content entity ID
- `mediable_type`: Content entity type
- `group`: Media grouping identifier

### Communication System

#### `comments`
**Purpose:** User-generated comments and discussions
**Key Features:**
- Hierarchical threading (Nested Set Model)
- Moderation workflow
- Polymorphic associations
- Status management

**Fields:**
- `id`: Primary key (UUID)
- `model_type`: Associated entity type
- `model_id`: Associated entity ID
- `title`: Comment title
- `status`: Moderation status (pending, approved, rejected, spam, trash)
- `parent_id`: Parent comment reference
- `record_left`, `record_right`, `record_depth`, `record_ordering`: NSM fields
- `created_by`, `updated_by`, `deleted_by`: User references
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

**Status Values:**
- `pending`: Awaiting moderation
- `approved`: Visible to users
- `rejected`: Not visible to users
- `spam`: Marked as spam
- `trash`: Deleted comment

**Indexes:**
- NSM operation indexes
- Status and moderation indexes
- Parent relationship index

### Navigation System

#### `menus`
**Purpose:** Navigation menu management
**Key Features:**
- Hierarchical structure (Nested Set Model)
- Role-based access control
- Visual customization
- Target configuration

**Fields:**
- `id`: Primary key (UUID)
- `name`: Menu display name
- `slug`: URL-friendly identifier (unique)
- `description`: Menu description
- `url`: Target URL
- `icon`: Icon identifier
- `parent_id`: Parent menu reference
- `record_left`, `record_right`, `record_depth`, `record_ordering`: NSM fields
- `is_active`: Active status flag
- `is_visible`: Visibility flag
- `target`: Link target (_self, _blank, _parent, _top)
- `created_by`, `updated_by`, `deleted_by`: User references
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

**Constraints:**
- Target validation for valid HTML target values
- Active and visible boolean constraints
- NSM integrity constraints

#### `menu_entities`
**Purpose:** Many-to-many relationship between menus and entities
**Key Features:**
- Junction table for menu-entity associations
- Polymorphic support for different entity types

**Fields:**
- `menu_id`: Menu reference
- `entity_id`: Entity ID
- `entity_type`: Entity type
- `created_at`, `updated_at`: Audit timestamps

### Organization Management

#### `organizations`
**Purpose:** Organizational structure management
**Key Features:**
- Hierarchical structure (Nested Set Model)
- Comprehensive organization types
- Status management
- Code-based identification

**Fields:**
- `id`: Primary key (UUID)
- `name`: Organization name
- `description`: Organization description
- `code`: Unique organization code
- `type`: Organization type classification
- `status`: Operational status (active, inactive, suspended)
- `parent_id`: Parent organization reference
- `record_left`, `record_right`, `record_depth`, `record_ordering`: NSM fields
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

**Organization Types:**
- Company, Subsidiary, Agent, Licensee, Distributor
- Outlet, Store, Department, Division
- Institution, Community, Foundation
- Branch Office, Regional, Franchisee, Partner

**Status Values:**
- `active`: Fully operational
- `inactive`: Temporarily inactive
- `suspended`: Suspended from operations

**Constraints:**
- NSM integrity constraints
- Status validation
- Record ordering constraints

### Configuration Management

#### `settings`
**Purpose:** Application configuration and settings
**Key Features:**
- Polymorphic associations
- Key-value storage
- Entity-specific configuration

**Fields:**
- `id`: Primary key (UUID)
- `model_type`: Associated entity type
- `model_id`: Associated entity ID
- `key`: Setting identifier
- `value`: Setting value
- `created_by`, `updated_by`, `deleted_by`: User references
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

### Address Management

#### `addresses`
**Purpose:** Geographic address storage
**Key Features:**
- Polymorphic associations
- Geographic coordinates
- Address type classification
- Primary address designation

**Fields:**
- `id`: Primary key (UUID)
- `addressable_id`: Associated entity ID
- `addressable_type`: Associated entity type
- `address_line1`: Primary address line
- `address_line2`: Secondary address line
- `city`: City name
- `state`: State/province
- `province`, `regency`, `district`, `sub_district`, `village`, `street`, `ward`: Geographic subdivisions
- `postal_code`: Postal/ZIP code
- `country`: Country name
- `latitude`, `longitude`: Geographic coordinates
- `is_primary`: Primary address flag
- `address_type`: Address purpose (home, work, billing, shipping, other)
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

**Constraints:**
- Addressable type validation
- Address type validation
- Geographic coordinate precision

**Indexes:**
- Composite index on addressable_id and addressable_type
- Primary address index
- Address type index
- Geographic coordinate indexes

### Activity Tracking

#### `activities`
**Purpose:** User activity and audit logging
**Key Features:**
- Comprehensive activity tracking
- JSON properties for flexible data storage
- Model association tracking

**Fields:**
- `id`: Primary key (UUID)
- `user_id`: User reference
- `action`: Performed action
- `model_type`: Affected entity type
- `model_id`: Affected entity ID
- `description`: Activity description
- `properties`: JSON properties for additional data
- `created_at`, `updated_at`, `deleted_at`: Audit timestamps

**Indexes:**
- User activity index
- Model type and ID indexes
- Deletion tracking index

## Database Relationships

### Primary Relationships

1. **User-Centric Relationships:**
   - Users can have multiple roles (through `role_entities`)
   - Users can create various content types
   - Users can have multiple addresses
   - Users can perform various activities

2. **Content Relationships:**
   - Posts can have content (through `contents`)
   - Posts can be tagged (through `taggables`)
   - Posts can have comments (through `comments`)
   - Posts can have associated media (through `mediables`)

3. **Hierarchical Relationships:**
   - Organizations form hierarchical structures
   - Taxonomies provide content classification
   - Menus create navigation hierarchies
   - Media files can be organized hierarchically
   - Comments support threaded discussions

4. **Polymorphic Relationships:**
   - Comments can be attached to any entity
   - Content can be stored for any entity
   - Settings can be configured for any entity
   - Addresses can be associated with any entity
   - Media can be attached to any entity

### Foreign Key Constraints

All foreign key relationships maintain referential integrity with appropriate cascade behaviors:
- **Cascade Delete:** Junction tables and dependent entities
- **Set Null:** User references when users are deleted
- **Restrict:** Critical business data relationships

## Performance Considerations

### Indexing Strategy

1. **Primary Key Indexes:** All tables have UUID primary keys
2. **Unique Constraint Indexes:** Username, email, slug fields
3. **Foreign Key Indexes:** All foreign key relationships
4. **Nested Set Indexes:** Left, right, depth, and ordering fields
5. **Query Performance Indexes:** Status, type, and date fields
6. **Composite Indexes:** Frequently queried field combinations

### Query Optimization

1. **Nested Set Model:** Efficient tree traversal operations
2. **Polymorphic Queries:** Optimized through proper indexing
3. **Soft Delete Filtering:** Indexed deletion timestamps
4. **Audit Trail Queries:** Optimized timestamp-based queries

## Security Features

1. **Password Security:** Hashed password storage
2. **Access Control:** Role-based permissions
3. **Audit Logging:** Comprehensive activity tracking
4. **Data Validation:** Constraint-based data integrity
5. **Soft Delete:** Data retention for compliance

## Migration Strategy

### Version Control
- Timestamped migration files
- Up and down migration support
- Dependency management through execution order

### Data Integrity
- Constraint validation during migration
- Index creation for performance
- Foreign key relationship establishment

### Rollback Support
- Complete down migration scripts
- Data preservation during rollbacks
- Constraint and index cleanup

## Future Considerations

### Scalability
- Partitioning strategies for large tables
- Read replica configurations
- Connection pooling optimization

### Performance
- Query optimization and monitoring
- Index maintenance and analysis
- Performance testing and benchmarking

### Maintenance
- Regular index maintenance
- Constraint validation
- Performance monitoring and tuning

## Conclusion

This database schema provides a robust foundation for a scalable, maintainable, and performant application. The use of established design patterns like Nested Set Model, polymorphic associations, and comprehensive audit trails ensures the system can grow and evolve while maintaining data integrity and performance.

The schema supports:
- Complex hierarchical data structures
- Flexible content management
- Comprehensive user management
- Robust audit and compliance features
- Scalable performance characteristics

Regular monitoring and maintenance will ensure optimal performance as the system grows and evolves.
