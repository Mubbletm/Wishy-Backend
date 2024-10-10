package sqlUtils

import (
	"database/sql"
	"reflect"
)

// Target should be a pointer
//
// Returns true if rows has next row. False if not, or if there's an error.
func BindRow(rows *sql.Rows, target interface{}) (bool, error) {
	names, err := rows.Columns()

	if err != nil || !rows.Next() {
		return false, err
	}

	values := make([]interface{}, len(names))
	for i := range values {
		values[i] = &values[i]
	}

	if err := rows.Scan(values...); err != nil {
		return false, err
	}

	targetReflect := reflect.ValueOf(target).Elem()

fieldLoop:
	for i := 0; i < targetReflect.Type().NumField(); i++ {
		// Loop through the struct's fields
		name := targetReflect.Type().Field(i).Name
		tag := targetReflect.Type().Field(i).Tag.Get("db")

		for _, searchName := range []string{name, tag} {
			// Loop through the column names in sorted order

			for columnIndex, columnName := range names {
				if columnName == searchName {
					value := values[columnIndex]
					targetReflect.FieldByName(name).Set(reflect.ValueOf(value))
					continue fieldLoop
				}
			}
		}
	}

	return true, nil
}
