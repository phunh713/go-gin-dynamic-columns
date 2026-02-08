package constants

import "fmt"

const TEMP_TABLE_NAME = "tmp_dynamiccolumn_ids"

var FORMULA_TEMPLATE = fmt.Sprintf(`
WITH {{cte}}
{{t_name}}_{{c_name}} AS (
    SELECT 
        {{t_name}}.id,
        {{formula}}
    FROM {{t_name}}
    JOIN %s tdi ON {{t_name}}.id = tdi.id
	{{cte_joins}}
)
UPDATE {{t_name}}
SET {{c_name}} = ct.{{c_name}}
FROM {{t_name}}_{{c_name}} ct
WHERE {{t_name}}.id = ct.id AND {{t_name}}.{{c_name}} IS DISTINCT FROM ct.{{c_name}}
`, TEMP_TABLE_NAME)

const SAMPLE_VARIABLES = `
var {{deployment}}.non_completed_count = COUNT(*) FILTER (WHERE {{deployment}}.status <> 'Completed')
var {{deployment}}.total_count = COUNT(*)
var {{invoice}}.total_count = COUNT(*)
var {{invoice}}.overdue_count = COUNT(*) FILTER (WHERE {{invoice}}.status = 'Overdue')
`

const SAMPLE_FORMULA_1 = `
CASE
	WHEN {{contract}}.is_cancelled = true THEN '%s'
	WHEN {{company}}.status <> '%s' THEN '%s'
	WHEN CURRENT_DATE < {{contract}}.start_date THEN '%s'
	WHEN CURRENT_DATE > {{contract}}.end_date THEN
		CASE
			WHEN COALESCE({{deployment}}.total_count, 0) = 0 THEN '%s'
			WHEN {{deployment}}.non_completed_count > 0 THEN '%s'
			WHEN COALESCE({{invoice}}.total_count, 0) = 0 THEN '%s'
			WHEN {{invoice}}.overdue_count > 0 THEN '%s'
			ELSE '%s'
		END
	WHEN COALESCE({{deployment}}.total_count, 0) = 0 THEN '%s'
	ELSE '%s'
END AS status
`

const SAMPLE_FORMULA = `
COUNT(*) FILTER (WHERE {{payment}}.amount > 1000)
`
