INSERT INTO section_counters (site_id, current_value)
SELECT s.id, COALESCE((
    SELECT MAX((value->>'id')::BIGINT)
    FROM pages, jsonb_array_elements(layout)
    WHERE site_id = s.id
), 0)
FROM sites s
WHERE NOT EXISTS (
    SELECT 1 FROM section_counters sc WHERE sc.site_id = s.id
);
