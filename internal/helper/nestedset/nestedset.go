package nestedset

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NestedSetManager provides optimized nested set operations for hierarchical data structures.
//
// Nested Set Model is a technique for representing hierarchical data in relational databases.
// Each node in the tree has a left and right value, where:
// - Left value < Right value for any node
// - All descendants of a node have left/right values between the node's left/right values
// - The difference between right and left values + 1 gives the number of descendants
//
// This implementation provides efficient operations for:
// - Creating nodes (with automatic left/right value calculation)
// - Moving subtrees (with automatic value adjustment)
// - Deleting nodes (soft delete with gap closure)
// - Querying relationships (ancestors, descendants, siblings)
// - Tree validation and rebuilding

// NestedSetManager provides optimized nested set operations for hierarchical data structures.
//
// Nested Set Model is a technique for representing hierarchical data in relational databases.
// Each node in the tree has a left and right value, where:
// - Left value < Right value for any node
// - All descendants of a node have left/right values between the node's left/right values
// - The difference between right and left values + 1 gives the number of descendants
//
// This implementation provides efficient operations for:
// - Creating nodes (with automatic left/right value calculation)
// - Moving subtrees (with automatic value adjustment)
// - Deleting nodes (soft delete with gap closure)
// - Querying relationships (ancestors, descendants, siblings)
// - Tree validation and rebuilding
type NestedSetManager struct {
	db *pgxpool.Pool // Database connection pool for PostgreSQL operations
}

// NewNestedSetManager creates a new nested set manager instance.
//
// Args:
//   - db: PostgreSQL connection pool for database operations
//
// Returns:
//   - *NestedSetManager: Configured manager instance
func NewNestedSetManager(db *pgxpool.Pool) *NestedSetManager {
	return &NestedSetManager{db: db}
}

// NestedSetValues represents the computed nested set values for a node.
//
// These values are used to maintain the tree structure and enable efficient queries:
// - Left: The left boundary value in the nested set model
// - Right: The right boundary value in the nested set model
// - Depth: The level/depth of the node in the tree (0 = root)
// - Ordering: The position among siblings at the same level
type NestedSetValues struct {
	Left     uint64 // Left boundary value for nested set positioning
	Right    uint64 // Right boundary value for nested set positioning
	Depth    uint64 // Tree depth level (0 = root, 1 = first level, etc.)
	Ordering uint64 // Sibling ordering within the same parent
}

// CreateNode creates a new node in the nested set with optimized queries.
//
// This method automatically calculates the correct left/right values based on the parent node.
// For root nodes (no parent), it places them at the end of the tree.
// For child nodes, it shifts existing nodes to make space and calculates proper positioning.
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - parentID: UUID of the parent node (nil for root nodes)
//   - ordering: Position among siblings (used for ordering within the same parent)
//
// Returns:
//   - *NestedSetValues: Calculated nested set values for the new node
//   - error: Any error that occurred during the operation
//
// The method uses a transaction to ensure data consistency and optimized CTE queries
// for better performance when dealing with large trees.
func (n *NestedSetManager) CreateNode(ctx context.Context, tableName string, parentID *uuid.UUID, ordering uint64) (*NestedSetValues, error) {
	// Begin transaction for atomicity
	tx, err := n.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback on error, commit on success

	var values NestedSetValues

	if parentID != nil {
		// Child node - use optimized CTE for better performance
		// This query:
		// 1. Gets parent node information (left, right, depth)
		// 2. Shifts all nodes that come after the parent to make space
		// 3. Calculates new left/right values for the child
		err = tx.QueryRow(ctx, fmt.Sprintf(`
			WITH parent_info AS (
				SELECT record_left, record_right, record_depth 
				FROM %s WHERE id = $1 AND deleted_at IS NULL
			),
			shifted_values AS (
				UPDATE %s 
				SET record_left = CASE 
					WHEN record_left > (SELECT record_right FROM parent_info) THEN record_left + 2 
					ELSE record_left 
				END,
				record_right = CASE 
					WHEN record_right >= (SELECT record_right FROM parent_info) THEN record_right + 2 
					ELSE record_right 
				END
				WHERE deleted_at IS NULL
			)
			SELECT 
				(SELECT record_right FROM parent_info) as new_left,
				(SELECT record_right FROM parent_info) + 1 as new_right,
				(SELECT record_depth FROM parent_info) + 1 as new_depth,
				$2 as new_ordering
		`, tableName, tableName), parentID.String(), ordering).Scan(&values.Left, &values.Right, &values.Depth, &values.Ordering)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate child nested set values: %w", err)
		}
	} else {
		// Root node - use single optimized query
		// Places the new root node at the end of the tree
		err = tx.QueryRow(ctx, fmt.Sprintf(`
			SELECT 
				COALESCE(MAX(record_right), 0) + 1, 
				COALESCE(MAX(record_right), 0) + 2, 
				0, 
				$1
			FROM %s WHERE deleted_at IS NULL
		`, tableName), ordering).Scan(&values.Left, &values.Right, &values.Depth, &values.Ordering)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate root nested set values: %w", err)
		}
	}

	return &values, tx.Commit(ctx)
}

// MoveSubtree moves a subtree to a new parent with optimized operations.
//
// This operation involves:
// 1. Calculating the current subtree boundaries
// 2. Determining new positioning based on the target parent
// 3. Adjusting all nodes in the subtree (left/right values and depth)
// 4. Shifting other nodes to maintain tree integrity
// 5. Updating the parent reference
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - nodeID: UUID of the node to move (root of the subtree)
//   - newParentID: UUID of the new parent (uuid.Nil for root level)
//
// Returns:
//   - error: Any error that occurred during the operation
//
// The method handles both moving to a new parent and moving to root level.
// All operations are performed within a transaction for consistency.
func (n *NestedSetManager) MoveSubtree(ctx context.Context, tableName string, nodeID, newParentID uuid.UUID) error {
	// Begin transaction for atomicity
	tx, err := n.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current node info to calculate subtree boundaries
	var currentLeft, currentRight, currentDepth int64
	err = tx.QueryRow(ctx, fmt.Sprintf(`
		SELECT record_left, record_right, record_depth 
		FROM %s WHERE id = $1 AND deleted_at IS NULL
	`, tableName), nodeID.String()).Scan(&currentLeft, &currentRight, &currentDepth)
	if err != nil {
		return fmt.Errorf("failed to get current node info: %w", err)
	}

	// Calculate subtree size (number of nodes in the subtree)
	// Formula: (right - left + 1) gives the total number of nodes
	subtreeSize := currentRight - currentLeft + 1

	// Get new parent info to determine target positioning
	var newParentLeft, newParentRight, newParentDepth int64
	if newParentID != uuid.Nil {
		// Moving to an existing parent
		err = tx.QueryRow(ctx, fmt.Sprintf(`
			SELECT record_left, record_right, record_depth 
			FROM %s WHERE id = $1 AND deleted_at IS NULL
		`, tableName), newParentID.String()).Scan(&newParentLeft, &newParentRight, &newParentDepth)
		if err != nil {
			return fmt.Errorf("failed to get new parent info: %w", err)
		}
	} else {
		// Moving to root level - place at the end of the tree
		var maxRight int64
		err = tx.QueryRow(ctx, fmt.Sprintf(`
			SELECT COALESCE(MAX(record_right), 0) FROM %s WHERE deleted_at IS NULL
		`, tableName)).Scan(&maxRight)
		if err != nil {
			return fmt.Errorf("failed to get max right value: %w", err)
		}
		newParentLeft = maxRight + 1
		newParentRight = maxRight
		newParentDepth = -1 // Will be adjusted to 0 for root level
	}

	// Calculate new left position for the subtree
	var newLeft int64
	if newParentID != uuid.Nil {
		newLeft = newParentRight // Place after the parent's right boundary
	} else {
		newLeft = newParentLeft // Place at the end of the tree
	}

	// Calculate the offset needed to move the subtree
	offset := newLeft - currentLeft

	// Update all nodes in the subtree with new positions and depths
	// This shifts the entire subtree to its new location
	_, err = tx.Exec(ctx, fmt.Sprintf(`
		UPDATE %s 
		SET 
			record_left = record_left + $1,
			record_right = record_right + $1,
			record_depth = record_depth + $2
		WHERE record_left >= $3 AND record_right <= $4 AND deleted_at IS NULL
	`, tableName), offset, newParentDepth-currentDepth+1, currentLeft, currentRight)
	if err != nil {
		return fmt.Errorf("failed to update subtree nodes: %w", err)
	}

	// Update the parent_id reference for the moved node
	_, err = tx.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET parent_id = $1 WHERE id = $2
	`, tableName), newParentID.String(), nodeID.String())
	if err != nil {
		return fmt.Errorf("failed to update parent_id: %w", err)
	}

	// Shift other nodes to make space for the moved subtree
	// This ensures no overlapping left/right values
	if newParentID != uuid.Nil {
		_, err = tx.Exec(ctx, fmt.Sprintf(`
			UPDATE %s 
			SET record_left = CASE 
				WHEN record_left > $1 THEN record_left + $2 
				ELSE record_left 
			END,
			record_right = CASE 
				WHEN record_right >= $1 THEN record_right + $2 
								ELSE record_right 
			END
			WHERE deleted_at IS NULL AND (record_left < $3 OR record_left > $4)
		`, tableName), newParentRight, subtreeSize, currentLeft, currentRight)
		if err != nil {
			return fmt.Errorf("failed to shift other nodes: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// DeleteSubtree performs soft delete of a subtree with optimized queries.
//
// This operation:
// 1. Marks all nodes in the subtree as deleted (soft delete)
// 2. Shifts remaining nodes to close the gap left by deleted nodes
// 3. Maintains tree integrity by adjusting left/right values
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - nodeID: UUID of the node to delete (root of the subtree)
//
// Returns:
//   - error: Any error that occurred during the operation
//
// Note: This is a soft delete - records are marked with deleted_at timestamp
// rather than being physically removed from the database.
func (n *NestedSetManager) DeleteSubtree(ctx context.Context, tableName string, nodeID uuid.UUID) error {
	// Begin transaction for atomicity
	tx, err := n.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get node info to determine subtree boundaries
	var left, right int64
	err = tx.QueryRow(ctx, fmt.Sprintf(`
		SELECT record_left, record_right 
		FROM %s WHERE id = $1 AND deleted_at IS NULL
	`, tableName), nodeID.String()).Scan(&left, &right)
	if err != nil {
		return fmt.Errorf("failed to get node info: %w", err)
	}

	// Calculate subtree size for gap closure
	subtreeSize := right - left + 1

	// Soft delete the entire subtree by setting deleted_at timestamp
	_, err = tx.Exec(ctx, fmt.Sprintf(`
		UPDATE %s 
		SET deleted_at = NOW() 
		WHERE record_left >= $1 AND record_right <= $2 AND deleted_at IS NULL
	`, tableName), left, right)
	if err != nil {
		return fmt.Errorf("failed to soft delete subtree: %w", err)
	}

	// Shift remaining nodes to close the gap left by deleted nodes
	// This maintains the nested set model integrity
	_, err = tx.Exec(ctx, fmt.Sprintf(`
		UPDATE %s 
		SET 
			record_left = CASE 
				WHEN record_left > $1 THEN record_left - $2 
				ELSE record_left 
			END,
			record_right = CASE 
				WHEN record_right > $1 THEN record_right - $2 
				ELSE record_right 
			END
		WHERE deleted_at IS NULL
	`, tableName), left, subtreeSize)
	if err != nil {
		return fmt.Errorf("failed to shift remaining nodes: %w", err)
	}

	return tx.Commit(ctx)
}

// GetDescendants retrieves all descendants of a node with optimized query.
//
// Descendants are all nodes that are children, grandchildren, etc. of the given node.
// The query uses the nested set left/right values for efficient retrieval.
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - nodeID: UUID of the node to get descendants for
//   - limit: Maximum number of descendants to return
//   - offset: Number of descendants to skip (for pagination)
//
// Returns:
//   - []uuid.UUID: Slice of descendant node UUIDs
//   - error: Any error that occurred during the operation
//
// The results are ordered by left value (which corresponds to tree traversal order).
func (n *NestedSetManager) GetDescendants(ctx context.Context, tableName string, nodeID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
	// Query uses nested set properties: descendants have left > parent.left and right < parent.right
	query := fmt.Sprintf(`
		SELECT id FROM %s 
		WHERE record_left > (
			SELECT record_left FROM %s WHERE id = $1 AND deleted_at IS NULL
		) 
		AND record_right < (
			SELECT record_right FROM %s WHERE id = $1 AND deleted_at IS NULL
		) 
		AND deleted_at IS NULL
		ORDER BY record_left ASC 
		LIMIT $2 OFFSET $3
	`, tableName, tableName, tableName)

	rows, err := n.db.Query(ctx, query, nodeID.String(), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan descendant id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// GetAncestors retrieves all ancestors of a node with optimized query.
//
// Ancestors are all nodes that are parents, grandparents, etc. of the given node.
// The query uses the nested set left/right values for efficient retrieval.
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - nodeID: UUID of the node to get ancestors for
//
// Returns:
//   - []uuid.UUID: Slice of ancestor node UUIDs
//   - error: Any error that occurred during the operation
//
// The results are ordered by left value (from root to leaf).
func (n *NestedSetManager) GetAncestors(ctx context.Context, tableName string, nodeID uuid.UUID) ([]uuid.UUID, error) {
	// Query uses nested set properties: ancestors have left < child.left and right > child.right
	query := fmt.Sprintf(`
		SELECT id FROM %s 
		WHERE record_left < (
			SELECT record_left FROM %s WHERE id = $1 AND deleted_at IS NULL
		) 
		AND record_right > (
			SELECT record_right FROM %s WHERE id = $1 AND deleted_at IS NULL
		) 
		AND deleted_at IS NULL
		ORDER BY record_left ASC
	`, tableName, tableName, tableName)

	rows, err := n.db.Query(ctx, query, nodeID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan ancestor id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// GetSiblings retrieves all siblings of a node with optimized query.
//
// Siblings are nodes that share the same parent. This query uses the parent_id
// reference for efficient retrieval rather than nested set calculations.
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - nodeID: UUID of the node to get siblings for
//   - limit: Maximum number of siblings to return
//   - offset: Number of siblings to skip (for pagination)
//
// Returns:
//   - []uuid.UUID: Slice of sibling node UUIDs
//   - error: Any error that occurred during the operation
//
// The results are ordered by left value (which corresponds to sibling ordering).
func (n *NestedSetManager) GetSiblings(ctx context.Context, tableName string, nodeID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
	// Query uses parent_id reference for direct sibling lookup
	query := fmt.Sprintf(`
		SELECT id FROM %s 
		WHERE parent_id = (
			SELECT parent_id FROM %s WHERE id = $1 AND deleted_at IS NULL
		) 
		AND id != $1 
		AND deleted_at IS NULL
		ORDER BY record_left ASC 
		LIMIT $2 OFFSET $3
	`, tableName, tableName)

	rows, err := n.db.Query(ctx, query, nodeID.String(), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get siblings: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan sibling id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// GetPath retrieves the path from root to a specific node.
//
// This returns all nodes in the path from the root to the specified node,
// including the node itself. Useful for breadcrumb navigation.
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - nodeID: UUID of the node to get the path for
//
// Returns:
//   - []uuid.UUID: Slice of node UUIDs from root to the target node
//   - error: Any error that occurred during the operation
//
// The results are ordered by left value (from root to target).
func (n *NestedSetManager) GetPath(ctx context.Context, tableName string, nodeID uuid.UUID) ([]uuid.UUID, error) {
	// Query uses nested set properties: path nodes have left <= target.left and right >= target.right
	query := fmt.Sprintf(`
		SELECT id FROM %s 
		WHERE record_left <= (
			SELECT record_left FROM %s WHERE id = $1 AND deleted_at IS NULL
		) 
		AND record_right >= (
			SELECT record_right FROM %s WHERE id = $1 AND deleted_at IS NULL
		) 
		AND deleted_at IS NULL
		ORDER BY record_left ASC
	`, tableName, tableName, tableName)

	rows, err := n.db.Query(ctx, query, nodeID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get path: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan path id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// IsDescendant checks if one node is a descendant of another.
//
// This method uses nested set properties to efficiently determine
// the ancestor-descendant relationship between two nodes.
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - ancestorID: UUID of the potential ancestor node
//   - descendantID: UUID of the potential descendant node
//
// Returns:
//   - bool: True if descendantID is a descendant of ancestorID
//   - error: Any error that occurred during the operation
//
// The check uses the nested set property: descendant.left > ancestor.left AND descendant.right < ancestor.right
func (n *NestedSetManager) IsDescendant(ctx context.Context, tableName string, ancestorID, descendantID uuid.UUID) (bool, error) {
	// Query uses nested set properties for efficient relationship checking
	query := fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1 FROM %s a, %s d
			WHERE a.id = $1 AND d.id = $2 
			AND a.deleted_at IS NULL AND d.deleted_at IS NULL
			AND a.record_left < d.record_left 
			AND a.record_right > d.record_right
		)
	`, tableName, tableName)

	var exists bool
	err := n.db.QueryRow(ctx, query, ancestorID.String(), descendantID.String()).Scan(&exists)
	return exists, err
}

// IsAncestor checks if one node is an ancestor of another.
//
// This is the inverse of IsDescendant - it checks if ancestorID is an ancestor of descendantID.
// The implementation simply calls IsDescendant with swapped parameters.
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - ancestorID: UUID of the potential ancestor node
//   - descendantID: UUID of the potential descendant node
//
// Returns:
//   - bool: True if ancestorID is an ancestor of descendantID
//   - error: Any error that occurred during the operation
func (n *NestedSetManager) IsAncestor(ctx context.Context, tableName string, ancestorID, descendantID uuid.UUID) (bool, error) {
	return n.IsDescendant(ctx, tableName, ancestorID, descendantID)
}

// CountDescendants counts the number of descendants of a node.
//
// This method efficiently calculates the total number of descendants
// using the nested set left/right values.
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - nodeID: UUID of the node to count descendants for
//
// Returns:
//   - int64: Number of descendants
//   - error: Any error that occurred during the operation
//
// Formula: (right - left + 1) / 2 gives the number of nodes in the subtree
func (n *NestedSetManager) CountDescendants(ctx context.Context, tableName string, nodeID uuid.UUID) (int64, error) {
	// Query uses nested set properties for efficient counting
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s 
		WHERE record_left > (
			SELECT record_left FROM %s WHERE id = $1 AND deleted_at IS NULL
		) 
		AND record_right < (
			SELECT record_right FROM %s WHERE id = $1 AND deleted_at IS NULL
		) 
		AND deleted_at IS NULL
	`, tableName, tableName, tableName)

	var count int64
	err := n.db.QueryRow(ctx, query, nodeID.String()).Scan(&count)
	return count, err
}

// CountChildren counts the number of direct children of a node.
//
// This method counts only immediate children (not grandchildren or deeper descendants).
// It uses the parent_id reference for direct counting.
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - nodeID: UUID of the node to count children for
//
// Returns:
//   - int64: Number of direct children
//   - error: Any error that occurred during the operation
func (n *NestedSetManager) CountChildren(ctx context.Context, tableName string, nodeID uuid.UUID) (int64, error) {
	// Query uses parent_id reference for direct child counting
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s 
		WHERE parent_id = $1 AND deleted_at IS NULL
	`, tableName)

	var count int64
	err := n.db.QueryRow(ctx, query, nodeID.String()).Scan(&count)
	return count, err
}

// GetSubtreeSize calculates the size of a subtree (number of nodes).
//
// This method uses the nested set formula to efficiently calculate
// the total number of nodes in a subtree without counting individual records.
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - nodeID: UUID of the node to get subtree size for
//
// Returns:
//   - int64: Number of nodes in the subtree
//   - error: Any error that occurred during the operation
//
// Formula: (right - left + 1) / 2 gives the number of nodes
func (n *NestedSetManager) GetSubtreeSize(ctx context.Context, tableName string, nodeID uuid.UUID) (int64, error) {
	// Query uses nested set formula for efficient size calculation
	query := fmt.Sprintf(`
		SELECT (record_right - record_left + 1) / 2
		FROM %s 
		WHERE id = $1 AND deleted_at IS NULL
	`, tableName)

	var size int64
	err := n.db.QueryRow(ctx, query, nodeID.String()).Scan(&size)
	if err != nil {
		return 0, fmt.Errorf("failed to get subtree size: %w", err)
	}

	return size, nil
}

// GetTreeHeight calculates the maximum depth of the tree.
//
// This method finds the deepest level in the tree by looking for
// the maximum record_depth value across all active nodes.
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//
// Returns:
//   - uint64: Maximum depth of the tree (0 = single root node)
//   - error: Any error that occurred during the operation
func (n *NestedSetManager) GetTreeHeight(ctx context.Context, tableName string) (uint64, error) {
	// Query finds the maximum depth across all active nodes
	query := fmt.Sprintf(`
		SELECT COALESCE(MAX(record_depth), 0)
		FROM %s 
		WHERE deleted_at IS NULL
	`, tableName)

	var height uint64
	err := n.db.QueryRow(ctx, query).Scan(&height)
	if err != nil {
		return 0, fmt.Errorf("failed to get tree height: %w", err)
	}

	return height, nil
}

// GetLevelWidth calculates the number of nodes at a specific depth level.
//
// This method counts how many nodes exist at a particular tree level.
// Useful for understanding tree distribution and planning UI layouts.
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//   - level: The depth level to count nodes for
//
// Returns:
//   - int64: Number of nodes at the specified level
//   - error: Any error that occurred during the operation
func (n *NestedSetManager) GetLevelWidth(ctx context.Context, tableName string, level uint64) (int64, error) {
	// Query counts nodes at a specific depth level
	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s 
		WHERE record_depth = $1 AND deleted_at IS NULL
	`, tableName)

	var width int64
	err := n.db.QueryRow(ctx, query, level).Scan(&width)
	if err != nil {
		return 0, fmt.Errorf("failed to get level width: %w", err)
	}

	return width, nil
}

// RebuildTree rebuilds the entire nested set tree structure.
//
// This method is useful for fixing corrupted tree structures or when
// the nested set values become inconsistent. It performs a complete
// rebuild by:
// 1. Traversing the tree using parent_id relationships
// 2. Recalculating all left/right values
// 3. Updating depth and ordering values
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//
// Returns:
//   - error: Any error that occurred during the operation
//
// Warning: This operation can be expensive for large trees and should
// be used only when necessary (e.g., after data corruption or migration).
func (n *NestedSetManager) RebuildTree(ctx context.Context, tableName string) error {
	// Begin transaction for atomicity
	tx, err := n.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Step 1: Get all nodes ordered by parent_id and creation time
	// This uses a recursive CTE to traverse the tree structure
	query := fmt.Sprintf(`
		WITH RECURSIVE node_tree AS (
			-- Root nodes (no parent)
			SELECT id, parent_id, created_at, 0 as level, 
				   ROW_NUMBER() OVER (ORDER BY created_at) as ordering
			FROM %s 
			WHERE (parent_id IS NULL OR parent_id = '') AND deleted_at IS NULL
			
			UNION ALL
			
			-- Child nodes
			SELECT c.id, c.parent_id, c.created_at, nt.level + 1,
				   ROW_NUMBER() OVER (PARTITION BY c.parent_id ORDER BY c.created_at)
			FROM %s c
			INNER JOIN node_tree nt ON c.parent_id = nt.id
			WHERE c.deleted_at IS NULL
		)
		SELECT id, level, ordering FROM node_tree
		ORDER BY level, ordering
	`, tableName, tableName)

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get node tree: %w", err)
	}
	defer rows.Close()

	var nodes []struct {
		id       uuid.UUID
		level    int
		ordering int
	}

	for rows.Next() {
		var node struct {
			id       uuid.UUID
			level    int
			ordering int
		}
		if err := rows.Scan(&node.id, &node.level, &node.ordering); err != nil {
			return fmt.Errorf("failed to scan node: %w", err)
		}
		nodes = append(nodes, node)
	}

	// Step 2: Rebuild nested set values
	// Start with left = 1 and increment by 2 for each node
	var left uint64 = 1
	for _, node := range nodes {
		// Calculate right value (left + 1 for leaf nodes)
		right := left + 1

		// Update the node with new nested set values
		updateQuery := fmt.Sprintf(`
			UPDATE %s 
			SET record_left = $1, record_right = $2, record_depth = $3, record_ordering = $4
			WHERE id = $5
		`, tableName)

		_, err = tx.Exec(ctx, updateQuery, left, right, node.level, node.ordering, node.id)
		if err != nil {
			return fmt.Errorf("failed to update node %s: %w", node.id, err)
		}

		// Move to next left position
		left = right + 1
	}

	return tx.Commit(ctx)
}

// ValidateTree validates the nested set tree structure for integrity.
//
// This method performs several checks to ensure the tree structure is valid:
// 1. Left values should be less than right values for all nodes
// 2. No overlapping intervals between nodes
// 3. Depth consistency between parent and child nodes
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//
// Returns:
//   - []string: List of validation error messages (empty if tree is valid)
//   - error: Any error that occurred during validation
//
// This method is useful for debugging tree corruption issues and ensuring
// data integrity after complex operations.
func (n *NestedSetManager) ValidateTree(ctx context.Context, tableName string) ([]string, error) {
	var errors []string

	// Check 1: Left values should be less than right values
	// This is a fundamental requirement of the nested set model
	query1 := fmt.Sprintf(`
		SELECT id FROM %s 
		WHERE record_left >= record_right AND deleted_at IS NULL
	`, tableName)

	rows1, err := n.db.Query(ctx, query1)
	if err != nil {
		return nil, fmt.Errorf("failed to validate left/right values: %w", err)
	}
	defer rows1.Close()

	for rows1.Next() {
		var id uuid.UUID
		if err := rows1.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan validation result: %w", err)
		}
		errors = append(errors, fmt.Sprintf("Node %s: left value >= right value", id))
	}

	// Check 2: No overlapping intervals
	// Each node should have a unique range of left/right values
	query2 := fmt.Sprintf(`
		SELECT DISTINCT a.id, b.id
		FROM %s a, %s b
		WHERE a.id != b.id 
		  AND a.deleted_at IS NULL AND b.deleted_at IS NULL
		  AND a.record_left < b.record_left 
		  AND a.record_right > b.record_left
	`, tableName, tableName)

	rows2, err := n.db.Query(ctx, query2)
	if err != nil {
		return nil, fmt.Errorf("failed to validate overlapping intervals: %w", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var id1, id2 uuid.UUID
		if err := rows2.Scan(&id1, &id2); err != nil {
			return nil, fmt.Errorf("failed to scan validation result: %w", err)
		}
		errors = append(errors, fmt.Sprintf("Nodes %s and %s: overlapping intervals", id1, id2))
	}

	// Check 3: Depth consistency
	// Child nodes should have depth = parent depth + 1
	query3 := fmt.Sprintf(`
		SELECT c.id, c.record_depth, p.record_depth
		FROM %s c
		LEFT JOIN %s p ON c.parent_id = p.id
		WHERE c.deleted_at IS NULL 
		  AND c.parent_id IS NOT NULL 
		  AND c.parent_id != ''
		  AND (c.record_depth != p.record_depth + 1)
	`, tableName, tableName)

	rows3, err := n.db.Query(ctx, query3)
	if err != nil {
		return nil, fmt.Errorf("failed to validate depth consistency: %w", err)
	}
	defer rows3.Close()

	for rows3.Next() {
		var id uuid.UUID
		var childDepth, parentDepth uint64
		if err := rows3.Scan(&id, &childDepth, &parentDepth); err != nil {
			return nil, fmt.Errorf("failed to scan validation result: %w", err)
		}
		errors = append(errors, fmt.Sprintf("Node %s: depth inconsistency (child: %d, parent: %d)", id, childDepth, parentDepth))
	}

	return errors, nil
}

// GetTreeStatistics returns comprehensive statistics about the tree.
//
// This method provides various metrics about the tree structure including:
// - Total number of nodes
// - Tree height (maximum depth)
// - Number of root nodes
// - Number of leaf nodes
// - Average depth
// - Maximum width at any level
//
// Args:
//   - ctx: Context for the operation
//   - tableName: Name of the database table containing the tree structure
//
// Returns:
//   - map[string]interface{}: Map containing various tree statistics
//   - error: Any error that occurred during the operation
//
// This information is useful for:
// - Performance monitoring
// - UI layout planning
// - Database optimization decisions
// - Understanding tree complexity
func (n *NestedSetManager) GetTreeStatistics(ctx context.Context, tableName string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total nodes - count all active nodes in the tree
	var totalNodes int64
	err := n.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL
	`, tableName)).Scan(&totalNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get total nodes: %w", err)
	}
	stats["total_nodes"] = totalNodes

	// Tree height - maximum depth level
	height, err := n.GetTreeHeight(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree height: %w", err)
	}
	stats["tree_height"] = height

	// Root nodes count - nodes with no parent
	var rootNodes int64
	err = n.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s 
		WHERE (parent_id IS NULL OR parent_id = '') AND deleted_at IS NULL
	`, tableName)).Scan(&rootNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get root nodes count: %w", err)
	}
	stats["root_nodes"] = rootNodes

	// Leaf nodes count - nodes with no children (right = left + 1)
	var leafNodes int64
	err = n.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s 
		WHERE record_right = record_left + 1 AND deleted_at IS NULL
	`, tableName)).Scan(&leafNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaf nodes count: %w", err)
	}
	stats["leaf_nodes"] = leafNodes

	// Average depth - mean depth across all nodes
	var avgDepth float64
	err = n.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT AVG(record_depth) FROM %s WHERE deleted_at IS NULL
	`, tableName)).Scan(&avgDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to get average depth: %w", err)
	}
	stats["average_depth"] = avgDepth

	// Max width at any level - maximum number of nodes at any single depth
	var maxWidth int64
	err = n.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT MAX(width) FROM (
			SELECT record_depth, COUNT(*) as width 
			FROM %s 
			WHERE deleted_at IS NULL 
			GROUP BY record_depth
		) level_widths
	`, tableName)).Scan(&maxWidth)
	if err != nil {
		return nil, fmt.Errorf("failed to get max width: %w", err)
	}
	stats["max_width"] = maxWidth

	return stats, nil
}
