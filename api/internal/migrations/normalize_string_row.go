package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type stringNormalizationEntry struct {
	srcRow        string
	normalizedRow string
}

func normalizeStringRow(
	ctx context.Context,
	tx *sql.Tx,
	tableName string,
	srcRowName string,
	destRowName string,
	normalizeFunc func(string) string,
) error {
	rows, err := tx.QueryContext(ctx, fmt.Sprintf("select %s from %s", srcRowName, tableName))
	if err != nil {
		return err
	}
	defer rows.Close()

	entries := make([]stringNormalizationEntry, 0)

	for rows.Next() {
		var entry stringNormalizationEntry
		if err := rows.Scan(&entry.srcRow); err != nil {
			return err
		}
		entry.normalizedRow = normalizeFunc(entry.srcRow)
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	batchLength := 32767
	// max postgres positiona args count: 65535 ( / 2 = 32767.5 )
	// we are having two pos args per entry => 32767 optimal
	for i := 0; i < len(entries); i += batchLength {
		end := min(i+batchLength, len(entries))

		chunk := entries[i:end]

		// process chunk

		if len(chunk) == 0 {
			return nil
		}
		var posArgs []string
		var values []any

		for i, e := range chunk {
			posArgs = append(posArgs, fmt.Sprintf("($%d, $%d)", (i*2)+1, (i*2)+2))
			values = append(values, e.srcRow, e.normalizedRow)
		}

		query := fmt.Sprintf(`
		 update %s as t
		 set %s = v.normalized
		 from (values %s) as v(src, normalized)
		 where t.%s = v.src;
		 `, tableName, destRowName, strings.Join(posArgs, ","), srcRowName)

		_, err := tx.ExecContext(ctx, query, values...)

		// process end

		if err != nil {
			return err
		}
	}
	return nil

}
