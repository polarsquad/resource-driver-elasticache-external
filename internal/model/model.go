package model

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// SelectAllAccessibleDrivers fetches all Drivers that belong to an org or are marked as public.
func (db model) SelectResourceMetadata(id string) (ResourceMetadata, bool, error) {
	row := db.QueryRow(`SELECT
		id,
		type,
		created_at,
		updated_at,
		deleted_at,
		params,
		data
    FROM resource_metadata
    WHERE id = $1`, id)

	var r ResourceMetadata
	err := row.Scan(&r.ID, &r.Type, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt, AsJSON(&r.Params), AsJSON(&r.Data))
	if err == sql.ErrNoRows {
		return ResourceMetadata{}, false, nil
	} else if err != nil {
		log.Printf("Database error fetching resource_metadata with id %s. (%v)", id, err)
		return ResourceMetadata{}, false, fmt.Errorf("select resource_metadata with id %s: %w", id, err)
	}

	return r, true, nil
}

// InsertOrUpdateResource adds or updates resource metadata.
func (db model) InsertOrUpdateResourceMetadata(m ResourceMetadata) error {
	_, err := db.Exec(`INSERT INTO resource_metadata (
		id,
		type,
		created_at,
		updated_at,
		deleted_at,
		params,
		data
  )
	VALUES ($1, $2, $3, $3, NULL, $4, $5)
	ON CONFLICT (id) DO
		UPDATE SET updated_at = $3, deleted_at = NULL, params = $4, data = $5 WHERE resource_metadata.id = $1
`,
		m.ID, m.Type, m.CreatedAt, *AsJSON(&m.Params), *AsJSON(&m.Data))
	if err != nil {
		log.Printf("Database error inserting resource_metadata with ID %s. (%v)", m.ID, err)
		return fmt.Errorf("insert resource_metadata with id %s: %w", m.ID, err)
	}
	return nil
}

// DeleteResourceMetadata removes an appenv.Application.  Also removes all associated environments.
func (db model) DeleteResourceMetadata(id string, deletedAt time.Time) error {
	result, err := db.Exec(`UPDATE resource_metadata SET deleted_at = $2 WHERE id = $1`, id, deletedAt)
	if err != nil {
		log.Printf("Database error deleting resource_metadata with id %s. (%v)", id, err)
		return fmt.Errorf("delete resource_metadata with id %s: %w", id, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Database retrieving rows-affected count after deleting resource_metadata with id %s. (%v)", id, err)
		return fmt.Errorf("delete resource_metadata with id %s: %w", id, err)
	}

	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
