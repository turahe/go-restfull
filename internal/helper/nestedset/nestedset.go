package nestedset

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NestedSetManager provides optimized nested set operations
type NestedSetManager struct {
	db *pgxpool.Pool
}

// NewNestedSetManager creates a new nested set manager
func NewNestedSetManager(db *pgxpool.Pool) *NestedSetManager {
	return &NestedSetManager{db: db}
}

// NestedSetValues represents the computed nested set values
type NestedSetValues struct {
	Left     uint64
	Right    uint64
	Depth    uint64
	Ordering uint64
}

// CreateNode creates a new node in the nested set with optimized queries
func (n *NestedSetManager) CreateNode(ctx context.Context, tableName string, parentID *uuid.UUID, ordering uint64) (*NestedSetValues, error) {
	tx, err := n.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var values NestedSetValues

	if parentID != nil {
		// Child node - use optimized CTE for better performance
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

// MoveSubtree moves a subtree to a new parent with optimized operations
func (n *NestedSetManager) MoveSubtree(ctx context.Context, tableName string, nodeID, newParentID uuid.UUID) error {
	tx, err := n.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current node info
	var currentLeft, currentRight, currentDepth int64
	err = tx.QueryRow(ctx, fmt.Sprintf(`
		SELECT record_left, record_right, record_depth 
		FROM %s WHERE id = $1 AND deleted_at IS NULL
	`, tableName), nodeID.String()).Scan(&currentLeft, &currentRight, &currentDepth)
	if err != nil {
		return fmt.Errorf("failed to get current node info: %w", err)
	}

	// Calculate subtree size
	subtreeSize := currentRight - currentLeft + 1

	// Get new parent info
	var newParentLeft, newParentRight, newParentDepth int64
	if newParentID != uuid.Nil {
		err = tx.QueryRow(ctx, fmt.Sprintf(`
			SELECT record_left, record_right, record_depth 
			FROM %s WHERE id = $1 AND deleted_at IS NULL
		`, tableName), newParentID.String()).Scan(&newParentLeft, &newParentRight, &newParentDepth)
		if err != nil {
			return fmt.Errorf("failed to get new parent info: %w", err)
		}
	} else {
		// Moving to root level
		var maxRight int64
		err = tx.QueryRow(ctx, fmt.Sprintf(`
			SELECT COALESCE(MAX(record_right), 0) FROM %s WHERE deleted_at IS NULL
		`, tableName)).Scan(&maxRight)
		if err != nil {
			return fmt.Errorf("failed to get max right value: %w", err)
		}
		newParentLeft = maxRight + 1
		newParentRight = maxRight
		newParentDepth = -1
	}

	// Calculate new positions
	var newLeft int64
	if newParentID != uuid.Nil {
		newLeft = newParentRight
	} else {
		newLeft = newParentLeft
	}

	// Calculate offset
	offset := newLeft - currentLeft

	// Update all nodes in the subtree
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

	// Update parent_id reference
	_, err = tx.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET parent_id = $1 WHERE id = $2
	`, tableName), newParentID.String(), nodeID.String())
	if err != nil {
		return fmt.Errorf("failed to update parent_id: %w", err)
	}

	// Shift other nodes to make space
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

// DeleteSubtree performs soft delete of a subtree with optimized queries
func (n *NestedSetManager) DeleteSubtree(ctx context.Context, tableName string, nodeID uuid.UUID) error {
	tx, err := n.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get node info
	var left, right int64
	err = tx.QueryRow(ctx, fmt.Sprintf(`
		SELECT record_left, record_right 
		FROM %s WHERE id = $1 AND deleted_at IS NULL
	`, tableName), nodeID.String()).Scan(&left, &right)
	if err != nil {
		return fmt.Errorf("failed to get node info: %w", err)
	}

	// Calculate subtree size
	subtreeSize := right - left + 1

	// Soft delete the subtree
	_, err = tx.Exec(ctx, fmt.Sprintf(`
		UPDATE %s 
		SET deleted_at = NOW() 
		WHERE record_left >= $1 AND record_right <= $2 AND deleted_at IS NULL
	`, tableName), left, right)
	if err != nil {
		return fmt.Errorf("failed to soft delete subtree: %w", err)
	}

	// Shift remaining nodes to close the gap
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

// GetDescendants retrieves all descendants of a node with optimized query
func (n *NestedSetManager) GetDescendants(ctx context.Context, tableName string, nodeID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
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

// GetAncestors retrieves all ancestors of a node with optimized query
func (n *NestedSetManager) GetAncestors(ctx context.Context, tableName string, nodeID uuid.UUID) ([]uuid.UUID, error) {
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

// GetSiblings retrieves all siblings of a node with optimized query
func (n *NestedSetManager) GetSiblings(ctx context.Context, tableName string, nodeID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
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

// GetPath retrieves the path from root to a specific node
func (n *NestedSetManager) GetPath(ctx context.Context, tableName string, nodeID uuid.UUID) ([]uuid.UUID, error) {
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

// IsDescendant checks if one node is a descendant of another
func (n *NestedSetManager) IsDescendant(ctx context.Context, tableName string, ancestorID, descendantID uuid.UUID) (bool, error) {
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

// IsAncestor checks if one node is an ancestor of another
func (n *NestedSetManager) IsAncestor(ctx context.Context, tableName string, ancestorID, descendantID uuid.UUID) (bool, error) {
	return n.IsDescendant(ctx, tableName, ancestorID, descendantID)
}

// CountDescendants counts the number of descendants of a node
func (n *NestedSetManager) CountDescendants(ctx context.Context, tableName string, nodeID uuid.UUID) (int64, error) {
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

// CountChildren counts the number of direct children of a node
func (n *NestedSetManager) CountChildren(ctx context.Context, tableName string, nodeID uuid.UUID) (int64, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s 
		WHERE parent_id = $1 AND deleted_at IS NULL
	`, tableName)

	var count int64
	err := n.db.QueryRow(ctx, query, nodeID.String()).Scan(&count)
	return count, err
}

// GetSubtreeSize calculates the size of a subtree (number of nodes)
func (n *NestedSetManager) GetSubtreeSize(ctx context.Context, tableName string, nodeID uuid.UUID) (int64, error) {
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

// GetTreeHeight calculates the maximum depth of the tree
func (n *NestedSetManager) GetTreeHeight(ctx context.Context, tableName string) (uint64, error) {
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

// GetLevelWidth calculates the number of nodes at a specific depth level
func (n *NestedSetManager) GetLevelWidth(ctx context.Context, tableName string, level uint64) (int64, error) {
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

// RebuildTree rebuilds the entire nested set tree structure
// This is useful for fixing corrupted tree structures
func (n *NestedSetManager) RebuildTree(ctx context.Context, tableName string) error {
	tx, err := n.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Step 1: Get all nodes ordered by parent_id and creation time
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
	var left uint64 = 1
	for _, node := range nodes {
		// Calculate right value
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

		left = right + 1
	}

	return tx.Commit(ctx)
}

// ValidateTree validates the nested set tree structure for integrity
func (n *NestedSetManager) ValidateTree(ctx context.Context, tableName string) ([]string, error) {
	var errors []string

	// Check 1: Left values should be less than right values
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

// GetTreeStatistics returns comprehensive statistics about the tree
func (n *NestedSetManager) GetTreeStatistics(ctx context.Context, tableName string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total nodes
	var totalNodes int64
	err := n.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL
	`, tableName)).Scan(&totalNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get total nodes: %w", err)
	}
	stats["total_nodes"] = totalNodes

	// Tree height
	height, err := n.GetTreeHeight(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree height: %w", err)
	}
	stats["tree_height"] = height

	// Root nodes count
	var rootNodes int64
	err = n.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s 
		WHERE (parent_id IS NULL OR parent_id = '') AND deleted_at IS NULL
	`, tableName)).Scan(&rootNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get root nodes count: %w", err)
	}
	stats["root_nodes"] = rootNodes

	// Leaf nodes count (nodes with no children)
	var leafNodes int64
	err = n.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s 
		WHERE record_right = record_left + 1 AND deleted_at IS NULL
	`, tableName)).Scan(&leafNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaf nodes count: %w", err)
	}
	stats["leaf_nodes"] = leafNodes

	// Average depth
	var avgDepth float64
	err = n.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT AVG(record_depth) FROM %s WHERE deleted_at IS NULL
	`, tableName)).Scan(&avgDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to get average depth: %w", err)
	}
	stats["average_depth"] = avgDepth

	// Max width at any level
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
