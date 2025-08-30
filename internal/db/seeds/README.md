# Database Seeder System

This directory contains the database seeder system for the Go RESTful API. The seeders are responsible for populating the database with initial data, including default roles, menus, and admin users.

## Available Seeders

### 1. Roles Seeder (`roles_seed.go`)
- Creates default roles: Admin, User, Moderator, Editor, Viewer
- Ensures the default "user" role exists
- Can be run independently or as part of the full seeding process

### 2. Menus Seeder (`admin_user_seed.go`)
- Creates default menu items for the system
- Includes: Dashboard, Users, Roles, Menus, Posts, Media, Settings
- Sets up hierarchical menu structure

### 3. Admin User Seeder (`admin_user_seed.go`)
- Creates a default admin user with credentials:
  - **Username**: `admin`
  - **Email**: `admin@example.com`
  - **Phone**: `+1234567890`
  - **Password**: Generated randomly (displayed in console)
- Assigns admin role to the user
- Assigns all menus to the admin role for full access

### 4. Main Seeder (`main_seeder.go`)
- Orchestrates all seeders in the correct order
- Ensures dependencies are seeded before dependent data
- Provides functions for running specific seeders or all seeders

## Usage

### Command Line Interface

The seeder system is integrated with the CLI commands:

```bash
# Run all seeders (recommended for first-time setup)
go run main.go seed

# Run only roles seeder
go run main.go seed:roles

# Run only admin user seeder
go run main.go seed:admin
```

### Programmatic Usage

You can also use the seeders programmatically:

```go
import "github.com/turahe/go-restfull/internal/db/seeds"

// Run all seeders
err := seeds.RunAllSeeders()

// Run specific seeder
err := seeds.RunSpecificSeeder("admin")

// Run individual seeders
err := seeds.SeedRoles()
err := seeds.SeedMenus()
err := seeds.SeedAdminUser()
```

## Seeding Order

The seeders must be run in the following order due to dependencies:

1. **Roles** - Creates the role system
2. **Menus** - Creates the menu structure
3. **Admin User** - Creates admin user and assigns roles/menus

## Database Requirements

Before running the seeders, ensure:

1. Database migrations have been run
2. The following tables exist:
   - `users`
   - `roles`
   - `menus`
   - `user_roles` (junction table)
   - `menu_roles` (junction table)

## Security Notes

⚠️ **Important Security Information:**

- The admin user seeder generates a random password and displays it in the console
- **Always change the admin password after first login**
- The generated password is only shown once during seeding
- In production, consider using environment variables for admin credentials

## Default Data

### Roles
- **Admin**: Full system access
- **User**: Basic user access
- **Moderator**: Content moderation access
- **Editor**: Content creation and editing access
- **Viewer**: Read-only access

### Menus
- **Dashboard**: Main dashboard (`/dashboard`)
- **Users**: User management (`/users`)
- **Roles**: Role management (`/roles`)
- **Menus**: Menu management (`/menus`)
- **Posts**: Post management (`/posts`)
- **Media**: Media management (`/media`)
- **Settings**: System settings (`/settings`)

## Error Handling

The seeders include comprehensive error handling:

- Checks for existing data before creation
- Logs all operations and errors
- Continues processing even if individual items fail
- Provides detailed error messages for debugging

## Customization

To customize the seeded data:

1. Modify the data arrays in the respective seeder files
2. Add new seeder functions for additional data types
3. Update the main seeder to include new seeders
4. Ensure proper dependency order is maintained

## Troubleshooting

### Common Issues

1. **Foreign Key Errors**: Ensure migrations are run before seeding
2. **Duplicate Key Errors**: Seeders check for existing data, but conflicts may occur
3. **Permission Errors**: Ensure database user has INSERT permissions

### Debug Mode

Enable debug logging by setting the log level to "debug" in your configuration:

```yaml
log:
  level: "debug"
```

This will provide detailed information about each seeding operation.
