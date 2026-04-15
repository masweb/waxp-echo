CREATE TABLE section_counters (
    site_id BIGINT NOT NULL PRIMARY KEY REFERENCES sites(id) ON DELETE CASCADE,
    current_value BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX idx_section_counters_site ON section_counters(site_id);
