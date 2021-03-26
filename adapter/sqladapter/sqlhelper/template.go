package sqlhelper

const (
	InsertTemplate = `
		INSERT INTO {{.Table}}
		  {{if .Columns }}({{.Columns}}){{end}}
		VALUES
		  {{.Values}}`

	UpdateTemplate = `
		UPDATE
		  {{.Table}}
		SET {{.ColumnValues}}
		  {{if .Where}}
			WHERE {{.Where}}
		  {{end}}`

	DeleteTemplate = `
		DELETE
		  FROM {{.Table}}
		  {{if .Where}}
			WHERE {{.Where}}
		  {{end}}
		{{if .Limit}}
		  LIMIT {{.Limit}}
		{{end}}
		{{if .Offset}}
		  OFFSET {{.Offset}}
		{{end}}`

	SelectTemplate = `
		SELECT
		  {{if .Distinct}}
			DISTINCT
		  {{end}}
	
		  {{if .Columns}}
			{{.Columns}}
		  {{else}}
			*
		  {{end}}
	
		  {{if .Table}}
			FROM {{.Table}}
		  {{end}}
	
		  {{.Joins}}
	
		  {{if .Where}}
		  	WHERE {{.Where}}
		  {{end}}
	
		  {{if .GroupBy}}
		  	GROUP BY {{.GroupBy}}
		  {{end}}
	
		  {{if .OrderBy}}
		  	ORDER BY {{.OrderBy}}
		  {{end}}
	
		  {{if .Limit}}
			LIMIT {{.Limit}}
		  {{end}}
	
		  {{if .Offset}}
			OFFSET {{.Offset}}
		  {{end}}
	  `

	CountTemplate = `
		SELECT
		  COUNT(1) AS _t
		FROM {{.Table}}
		  {{if .Where}}
			WHERE {{.Where}}
		  {{end}}
	
		  {{if .Limit}}
			LIMIT {{.Limit}}
		  {{end}}
	
		  {{if .Offset}}
			OFFSET {{.Offset}}
		  {{end}}
	  `
)
