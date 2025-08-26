# Laravel-Style Resource Responses Implementation

This document provides a comprehensive overview of the Laravel-style resource responses implementation across multiple endpoints in the Go RESTful API project.

## Overview

The project has been updated to use Laravel-style API resource responses, providing consistent and well-structured API responses that follow Laravel's API resource patterns. This implementation covers multiple controllers and provides a foundation for consistent API responses throughout the system.

## Implemented Endpoints

### 1. Address Controller ✅ COMPLETED

**File**: `internal/interfaces/http/controllers/address_controller.go`
**Resource File**: `internal/interfaces/http/responses/address_responses.go`

**Endpoints Updated**:
- `POST /api/v1/addresses` - Create address
- `GET /api/v1/addresses/{id}` - Get address by ID
- `PUT /api/v1/addresses/{id}` - Update address
- `GET /api/v1/addressables/{type}/{id}/addresses` - Get addresses by addressable
- `GET /api/v1/addressables/{type}/{id}/addresses/primary` - Get primary address
- `GET /api/v1/addressables/{type}/{id}/addresses/type/{type}` - Get addresses by type
- `GET /api/v1/addresses/search/city` - Search by city (paginated)
- `GET /api/v1/addresses/search/state` - Search by state (paginated)
- `GET /api/v1/addresses/search/country` - Search by country (paginated)
- `GET /api/v1/addresses/search/postal-code` - Search by postal code (paginated)

**Response Types**:
- `AddressResourceResponse` - Single address
- `AddressCollectionResponse` - Collection of addresses
- `AddressCollectionResponse` with pagination metadata

**Features**:
- Computed fields: `full_address`, `has_coordinates`
- Pagination support with metadata and navigation links
- Consistent response structure

### 2. User Controller ✅ COMPLETED

**File**: `internal/interfaces/http/controllers/user_controller.go`
**Resource File**: `internal/interfaces/http/responses/user_responses.go`

**Endpoints Updated**:
- `POST /users` - Create user
- `GET /users/{id}` - Get user by ID
- `GET /users` - Get all users (paginated)
- `PUT /users/{id}` - Update user

**Response Types**:
- `UserResourceResponse` - Single user
- `UserCollectionResponse` - Collection of users
- `UserCollectionResponse` with pagination metadata

**Features**:
- Computed fields: `is_email_verified`, `is_phone_verified`, `has_avatar`
- Nested resources: `roles`, `menus`
- Pagination support with metadata and navigation links

### 3. Post Controller ✅ COMPLETED

**File**: `internal/interfaces/http/controllers/post_controller.go`
**Resource File**: `internal/interfaces/http/responses/post_responses.go`

**Endpoints Updated**:
- `POST /posts` - Create post
- `GET /posts/{id}` - Get post by ID
- `GET /posts/slug/{slug}` - Get post by slug
- `GET /posts` - Get all posts (paginated)
- `GET /posts/author/{authorID}` - Get posts by author (paginated)
- `PUT /posts/{id}` - Update post
- `PUT /posts/{id}/publish` - Publish post
- `PUT /posts/{id}/unpublish` - Unpublish post

**Response Types**:
- `PostResourceResponse` - Single post
- `PostCollectionResponse` - Collection of posts
- `PostCollectionResponse` with pagination metadata

**Features**:
- Computed fields: `is_published`, `status`
- Pagination support with metadata and navigation links
- Consistent response structure

### 4. Organization Controller ✅ COMPLETED

**File**: `internal/interfaces/http/controllers/organization_controller.go`
**Resource File**: `internal/interfaces/http/responses/organization_responses.go`

**Endpoints Updated**:
- `POST /organizations` - Create organization
- `GET /organizations/{id}` - Get organization by ID
- `GET /organizations` - Get all organizations (paginated)
- `PUT /organizations/{id}` - Update organization
- `GET /organizations/root` - Get root organizations
- `GET /organizations/{id}/children` - Get organization children
- `GET /organizations/{id}/descendants` - Get organization descendants
- `GET /organizations/{id}/ancestors` - Get organization ancestors
- `GET /organizations/{id}/siblings` - Get organization siblings
- `GET /organizations/{id}/path` - Get organization path
- `GET /organizations/tree` - Get organization tree
- `GET /organizations/{id}/subtree` - Get organization subtree
- `POST /organizations/{id}/children` - Add organization child
- `POST /organizations/{id}/move` - Move organization subtree
- `DELETE /organizations/{id}/subtree` - Delete organization subtree
- `PUT /organizations/{id}/status` - Set organization status
- `GET /organizations/search` - Search organizations (paginated)
- `GET /organizations/{id}/stats` - Get organization statistics
- `GET /organizations/validate-hierarchy` - Validate organization hierarchy

**Response Types**:
- `OrganizationResourceResponse` - Single organization
- `OrganizationCollectionResponse` - Collection of organizations
- `OrganizationCollectionResponse` with pagination metadata

**Features**:
- Computed fields: `is_root`, `has_children`, `has_parent`, `level`
- Nested resources: `parent`, `children`
- Hierarchical structure support
- Pagination support with metadata and navigation links

### 5. Taxonomy Controller ✅ COMPLETED

**File**: `internal/interfaces/http/controllers/taxonomy_controller.go`
**Resource File**: `internal/interfaces/http/responses/taxonomy_responses.go`

**Endpoints Updated**:
- `POST /api/v1/taxonomies` - Create taxonomy
- `GET /api/v1/taxonomies/{id}` - Get taxonomy by ID
- `GET /api/v1/taxonomies/slug/{slug}` - Get taxonomy by slug
- `GET /api/v1/taxonomies` - Get all taxonomies (paginated)
- `GET /api/v1/taxonomies/root` - Get root taxonomies
- `GET /api/v1/taxonomies/hierarchy` - Get taxonomy hierarchy
- `GET /api/v1/taxonomies/{id}/children` - Get taxonomy children
- `GET /api/v1/taxonomies/{id}/descendants` - Get taxonomy descendants
- `GET /api/v1/taxonomies/{id}/ancestors` - Get taxonomy ancestors
- `GET /api/v1/taxonomies/{id}/siblings` - Get taxonomy siblings
- `GET /api/v1/taxonomies/search` - Search taxonomies (paginated)
- `GET /api/v1/taxonomies/search/advanced` - Advanced search with pagination

**Response Types**:
- `TaxonomyResourceResponse` - Single taxonomy
- `TaxonomyCollectionResponse` - Collection of taxonomies
- `TaxonomyCollectionResponse` with pagination metadata

**Features**:
- Computed fields: `is_root`, `has_children`, `has_parent`, `level`
- Nested resources: `parent`, `children`
- Hierarchical structure support
- Pagination support with metadata and navigation links

### 6. Comment Controller ✅ COMPLETED

**File**: `internal/interfaces/http/controllers/comment_controller.go`
**Resource File**: `internal/interfaces/http/responses/comment_responses.go`

**Endpoints Updated**:
- `GET /api/v1/comments` - Get all comments (paginated with filters)
- `GET /api/v1/comments/{id}` - Get comment by ID
- `POST /api/v1/comments` - Create comment
- `PUT /api/v1/comments/{id}` - Update comment
- `PUT /api/v1/comments/{id}/approve` - Approve comment
- `PUT /api/v1/comments/{id}/reject` - Reject comment

**Response Types**:
- `CommentResourceResponse` - Single comment
- `CommentCollectionResponse` - Collection of comments
- `CommentCollectionResponse` with pagination metadata

**Features**:
- Content support: `ContentResource` with raw and HTML content
- Author information: `UserResource` for comment authors
- Computed fields: `is_reply`, `is_approved`, `is_pending`, `is_rejected`, `is_deleted`
- Hierarchical comments: Support for nested comment structures
- Pagination support with metadata and navigation links
- Status management: Proper handling of comment approval/rejection workflows

### 7. Tag Controller ✅ COMPLETED

**File**: `internal/interfaces/http/controllers/tag_controller.go`
**Resource File**: `internal/interfaces/http/responses/tag_responses.go`

**Endpoints Updated**:
- `GET /api/v1/tags` - Get all tags (paginated)
- `GET /api/v1/tags/{id}` - Get tag by ID
- `POST /api/v1/tags` - Create tag
- `PUT /api/v1/tags/{id}` - Update tag
- `DELETE /api/v1/tags/{id}` - Delete tag
- `GET /api/v1/tags/search` - Search tags

**Response Types**:
- `TagResourceResponse` - Single tag
- `TagCollectionResponse` - Collection of tags
- `TagCollectionResponse` with pagination metadata

**Features**:
- Computed fields: `is_deleted`
- Pagination support with metadata and navigation links
- Search functionality with collection responses
- Consistent response structure across all endpoints
- Swagger documentation updated with new response types

### 8. Media Controller ✅ COMPLETED

**File**: `internal/interfaces/http/controllers/media_controller.go`
**Resource File**: `internal/interfaces/http/responses/media_responses.go`

**Endpoints Updated**:
- `GET /api/v1/media` - Get all media (paginated with search)
- `GET /api/v1/media/{id}` - Get media by ID
- `POST /api/v1/media` - Upload media
- `PUT /api/v1/media/{id}` - Update media metadata
- `DELETE /api/v1/media/{id}` - Delete media

**Response Types**:
- `MediaResourceResponse` - Single media item
- `MediaCollectionResponse` - Collection of media items
- `MediaCollectionResponse` with pagination metadata

**Features**:
- Rich computed fields: `is_image`, `is_video`, `is_audio`, `file_extension`, `url`, `file_size_in_mb`, `file_size_in_kb`
- File type detection and categorization
- Pagination support with metadata and navigation links
- Search functionality with collection responses
- Consistent response structure across all endpoints
- Swagger documentation updated with new response types
- Support for hierarchical organization (nested set model)

### 9. Menu Controller ✅ COMPLETED

**File**: `internal/interfaces/http/controllers/menu_controller.go`
**Resource File**: `internal/interfaces/http/responses/menu_responses.go`

**Endpoints Updated**:
- `GET /api/v1/menus` - Get all menus (paginated with filters)
- `GET /api/v1/menus/{id}` - Get menu by ID
- `GET /api/v1/menus/slug/{slug}` - Get menu by slug
- `GET /api/v1/menus/root` - Get root menus
- `GET /api/v1/menus/hierarchy` - Get menu hierarchy
- `GET /api/v1/users/{user_id}/menus` - Get user menus
- `GET /api/v1/menus/search` - Search menus
- `POST /api/v1/menus` - Create menu
- `PUT /api/v1/menus/{id}` - Update menu
- `DELETE /api/v1/menus/{id}` - Delete menu
- `PATCH /api/v1/menus/{id}/activate` - Activate menu
- `PATCH /api/v1/menus/{id}/deactivate` - Deactivate menu
- `PATCH /api/v1/menus/{id}/show` - Show menu
- `PATCH /api/v1/menus/{id}/hide` - Hide menu

**Response Types**:
- `MenuResourceResponse` - Single menu item
- `MenuCollectionResponse` - Collection of menu items
- `MenuCollectionResponse` with pagination metadata

**Features**:
- Rich computed fields: `is_root`, `is_leaf`, `depth`, `width`
- Hierarchical structure support with nested set model
- Nested resources: `parent`, `children`, `roles`
- Pagination support with metadata and navigation links
- Search functionality with collection responses
- Consistent response structure across all endpoints
- Swagger documentation updated with new response types
- Menu state management (active/inactive, visible/hidden)
- Role-based access control integration

## Resource Response Structure

### Common Response Format

All resource responses follow this consistent structure:

```json
{
  "status": "success",
  "data": {
    // Resource-specific data
  }
}
```

### Collection Response Format

Collection responses include pagination metadata and navigation links:

```json
{
  "status": "success",
  "data": {
    "data": [
      // Array of resources
    ],
    "meta": {
      "current_page": 1,
      "per_page": 10,
      "total_items": 25,
      "total_pages": 3,
      "has_next_page": true,
      "has_prev_page": false,
      "next_page": 2,
      "previous_page": 0,
      "from": 1,
      "to": 10
    },
    "links": {
      "first": "/api/v1/endpoint",
      "last": "/api/v1/endpoint?page=3",
      "prev": "",
      "next": "/api/v1/endpoint?page=2"
    }
  }
}
```

## Key Features

### 1. Computed Fields
- **Address**: `full_address`, `has_coordinates`
- **User**: `is_email_verified`, `is_phone_verified`, `has_avatar`
- **Post**: `is_published`, `status`
- **Organization**: `is_root`, `has_children`, `has_parent`, `level`
- **Taxonomy**: `is_root`, `has_children`, `has_parent`, `level`
- **Comment**: `is_reply`, `is_approved`, `is_pending`, `is_rejected`, `is_deleted`
- **Tag**: `is_deleted`
- **Media**: `is_image`, `is_video`, `is_audio`, `file_extension`, `url`, `file_size_in_mb`, `file_size_in_kb`
- **Menu**: `is_root`, `is_leaf`, `depth`, `width`

### 2. Pagination Support
- Built-in pagination metadata
- Navigation links for easy pagination
- Automatic page calculation from offset/limit
- Consistent pagination structure across all endpoints

### 3. Nested Resources
- **User**: Includes `roles` and `menus` as nested resources
- **Address**: Includes computed fields based on entity methods
- **Post**: Includes status information based on entity state
- **Organization**: Includes `parent` and `children` as nested resources
- **Taxonomy**: Includes `parent` and `children` as nested resources
- **Comment**: Includes `content` and `author` as nested resources

### 4. Consistent Structure
- All responses follow the same format
- Standardized error handling
- Consistent field naming conventions
- Laravel-familiar response patterns

## Benefits

1. **Professional API**: Consistent, well-structured responses
2. **Developer Experience**: Predictable response formats
3. **Maintainability**: Centralized response formatting
4. **Pagination**: Built-in pagination support
5. **Computed Fields**: Additional useful information automatically included
6. **Laravel Familiarity**: Recognizable patterns for Laravel developers
7. **API Documentation**: Better Swagger/OpenAPI documentation
8. **Frontend Integration**: Easier integration with frontend applications

## Migration from Old Response Format

### Before (Old Format)
```go
return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
    Status: "success",
    Data:   entity,
})
```

### After (New Format)
```go
return ctx.Status(http.StatusOK).JSON(responses.NewEntityResourceResponse(entity))
```

## Swagger Documentation Updates

All endpoints have been updated with new response types in their Swagger documentation:

- `responses.SuccessResponse{data=Entity}` → `responses.EntityResourceResponse`
- `responses.SuccessResponse{data=[]Entity}` → `responses.EntityCollectionResponse`

## Testing

To test the new resource responses:

1. Run the application
2. Make requests to the updated endpoints
3. Verify that responses follow the new structure
4. Check that pagination works correctly for collection endpoints
5. Ensure computed fields are present in responses
6. Verify that nested resources are properly included

## Role Controller Implementation

The Role Controller has been successfully refactored to use Laravel-style resource responses. Here's what was implemented:

### Updated Endpoints

1. **GetRoles** (`GET /roles`)
   - **Response Type**: `responses.RoleCollectionResponse`
   - **Features**: Pagination with page/limit parameters, active status filtering
   - **Pagination**: Includes metadata and navigation links

2. **GetRoleByID** (`GET /roles/{id}`)
   - **Response Type**: `responses.RoleResourceResponse`
   - **Features**: Single role retrieval with full resource details

3. **GetRoleBySlug** (`GET /roles/slug/{slug}`)
   - **Response Type**: `responses.RoleResourceResponse`
   - **Features**: Role retrieval by URL-friendly slug

4. **CreateRole** (`POST /roles`)
   - **Response Type**: `responses.RoleResourceResponse`
   - **Features**: Role creation with 201 status and success message

5. **UpdateRole** (`PUT /roles/{id}`)
   - **Response Type**: `responses.RoleResourceResponse`
   - **Features**: Role update with full resource response

6. **DeleteRole** (`DELETE /roles/{id}`)
   - **Response Type**: `responses.ErrorResponse` (success case)
   - **Features**: Soft delete with success confirmation

7. **ActivateRole** (`PUT /roles/{id}/activate`)
   - **Response Type**: `responses.RoleResourceResponse`
   - **Features**: Role activation with updated resource

8. **DeactivateRole** (`PUT /roles/{id}/deactivate`)
   - **Response Type**: `responses.RoleResourceResponse`
   - **Features**: Role deactivation with updated resource

9. **SearchRoles** (`GET /roles/search`)
   - **Response Type**: `responses.RoleCollectionResponse`
   - **Features**: Role search by name/description with pagination

### RoleResource Features

- **Core Fields**: ID, Name, Slug, Description, IsActive
- **Audit Fields**: CreatedBy, UpdatedBy, DeletedBy, CreatedAt, UpdatedAt, DeletedAt
- **Computed Fields**: IsDeleted, IsActiveRole
- **Pagination**: Full pagination metadata and navigation links
- **Error Handling**: Consistent error response format

### Swagger Documentation Updates

All endpoints have been updated with new response types:
- `@Success 200 {object} responses.RoleResourceResponse` for single role responses
- `@Success 200 {object} responses.RoleCollectionResponse` for role collections
- `@Failure` annotations updated to use `responses.ErrorResponse`

## 10. Auth Controller ✅ COMPLETED

**File**: `internal/interfaces/http/controllers/auth_controller.go`
**Resource File**: `internal/interfaces/http/responses/auth_responses.go`

**Endpoints Updated**:
- `POST /auth/register` - User registration
- `POST /auth/login` - User authentication
- `POST /auth/refresh` - Token refresh
- `POST /auth/logout` - User logout
- `POST /auth/forget-password` - Password reset request
- `POST /auth/reset-password` - Password reset

**Response Types**:
- `AuthResourceResponse` - Authentication with user details
- `TokenResourceResponse` - Token pair responses
- `ErrorResponse` - Error cases
- `ValidationErrorResponse` - Validation errors

**Features**:
- Nested resources: `User` (simplified version)
- Token management: Access and refresh tokens
- Consistent error handling
- User authentication state

## 11. User Role Controller ✅ COMPLETED

**File**: `internal/interfaces/http/controllers/user_role_controller.go`
**Resource File**: `internal/interfaces/http/responses/user_role_responses.go`

**Endpoints Updated**:
- `POST /api/v1/users/{user_id}/roles/{role_id}` - Assign role to user
- `DELETE /api/v1/users/{user_id}/roles/{role_id}` - Remove role from user
- `GET /api/v1/users/{user_id}/roles` - Get user roles
- `GET /api/v1/roles/{role_id}/users` - Get role users (paginated)
- `GET /api/v1/users/{user_id}/roles/{role_id}/check` - Check if user has role
- `GET /api/v1/roles/{role_id}/users/count` - Get user count for role

**Response Types**:
- `UserRoleResourceResponse` - Single user-role relationship
- `UserRoleCollectionResponse` - Collection of user-role relationships
- `RoleUserCollectionResponse` - Collection of role-user relationships
- `SuccessResponse` - Success messages
- `ErrorResponse` - Error cases

**Features**:
- Nested resources: `User` and `Role` details
- Pagination support for role users
- Relationship management
- Consistent error handling

## 12. RBAC Controller ✅ COMPLETED

**File**: `internal/interfaces/http/controllers/rbac_controller.go`
**Resource File**: `internal/interfaces/http/responses/rbac_responses.go`

**Endpoints Updated**:
- `GET /rbac/policies` - Get all RBAC policies
- `POST /rbac/policies` - Add new RBAC policy
- `DELETE /rbac/policies` - Remove RBAC policy
- `GET /rbac/users/{user}/roles` - Get roles for user
- `POST /rbac/users/{user}/roles/{role}` - Add role to user
- `DELETE /rbac/users/{user}/roles/{role}` - Remove role from user
- `GET /rbac/roles/{role}/users` - Get users for role

**Response Types**:
- `RBACPolicyResourceResponse` - Single policy response
- `RBACPolicyCollectionResponse` - Collection of policies
- `RBACRoleCollectionResponse` - Collection of roles
- `RBACUserCollectionResponse` - Collection of users
- `SuccessResponse` - Success messages
- `ErrorResponse` - Error cases

**Features**:
- Policy management: Subject-Object-Action relationships
- Role-user assignments
- Consistent error handling
- Structured policy representation

## Future Enhancements

### 1. Additional Controllers
- **Role Controller**: ✅ Implemented role resource responses
- **Auth Controller**: ✅ Implemented auth resource responses
- **User Role Controller**: ✅ Implemented user-role resource responses
- **RBAC Controller**: ✅ Implemented RBAC resource responses

### 2. Advanced Features
- **Conditional Fields**: Include/exclude fields based on user permissions
- **Resource Transformers**: Reusable transformers for common entity types
- **API Versioning**: Support different response formats for different API versions
- **Caching**: Cache transformed resources for better performance
- **Validation**: Add response validation to ensure consistency

### 3. Standardization
- **Common Resource Interface**: Define common interfaces for all resources
- **Response Factory**: Centralized factory for creating responses
- **Error Handling**: Standardized error response formats
- **Logging**: Enhanced logging for response generation

## Files Modified

### Controllers
- `internal/interfaces/http/controllers/address_controller.go` - Updated to use new responses
- `internal/interfaces/http/controllers/user_controller.go` - Updated to use new responses
- `internal/interfaces/http/controllers/post_controller.go` - Updated to use new responses
- `internal/interfaces/http/controllers/organization_controller.go` - Updated to use new responses
- `internal/interfaces/http/controllers/taxonomy_controller.go` - Updated to use new responses
- `internal/interfaces/http/controllers/comment_controller.go` - Updated to use new responses
- `internal/interfaces/http/controllers/tag_controller.go` - Updated to use new responses
- `internal/interfaces/http/controllers/media_controller.go` - Updated to use new responses
- `internal/interfaces/http/controllers/menu_controller.go` - Updated to use new responses
- `internal/interfaces/http/controllers/role_controller.go` - Updated to use new responses
- `internal/interfaces/http/controllers/auth_controller.go` - Updated to use new responses
- `internal/interfaces/http/controllers/user_role_controller.go` - Updated to use new responses
- `internal/interfaces/http/controllers/rbac_controller.go` - Updated to use new responses

### Response Files
- `internal/interfaces/http/responses/address_responses.go` - New resource response structures
- `internal/interfaces/http/responses/user_responses.go` - New resource response structures
- `internal/interfaces/http/responses/post_responses.go` - New resource response structures
- `internal/interfaces/http/responses/organization_responses.go` - New resource response structures
- `internal/interfaces/http/responses/taxonomy_responses.go` - New resource response structures
- `internal/interfaces/http/responses/comment_responses.go` - New resource response structures
- `internal/interfaces/http/responses/tag_responses.go` - New resource response structures
- `internal/interfaces/http/responses/media_responses.go` - New resource response structures
- `internal/interfaces/http/responses/menu_responses.go` - New resource response structures
- `internal/interfaces/http/responses/role_responses.go` - New resource response structures
- `internal/interfaces/http/responses/auth_responses.go` - New resource response structures
- `internal/interfaces/http/responses/user_role_responses.go` - New resource response structures
- `internal/interfaces/http/responses/rbac_responses.go` - New resource response structures

### Documentation
- `docs/development/laravel-resource-responses.md` - Detailed implementation guide
- `docs/development/laravel-resource-responses-implementation.md` - This comprehensive overview

## Conclusion

The implementation of Laravel-style resource responses provides a solid foundation for consistent and professional API responses. The structured approach makes it easier for frontend developers to work with the API and provides better documentation through the response format itself.

The pattern established here can be easily extended to other controllers and provides a scalable solution for maintaining consistent API responses across the entire system.

## Next Steps

1. **Continue Implementation**: Apply the same pattern to remaining controllers
2. **Testing**: Comprehensive testing of all updated endpoints
3. **Documentation**: Update API documentation and examples
4. **Performance**: Monitor and optimize response generation
5. **Feedback**: Gather feedback from frontend developers and API consumers
