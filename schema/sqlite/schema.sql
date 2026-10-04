CREATE TABLE IF NOT EXISTS {{space_groups.table}} (
    {{space_groups.id}} INTEGER PRIMARY KEY AUTOINCREMENT,
    {{space_groups.name}} TEXT NOT NULL CHECK (length({{space_groups.name}}) <= 255)
) STRICT;

CREATE TABLE IF NOT EXISTS {{spaces.table}} (
    {{spaces.id}} INTEGER PRIMARY KEY AUTOINCREMENT,
    {{spaces.name}} TEXT NOT NULL CHECK (length({{spaces.name}}) <= 255),
    {{spaces.space_group_id}} INTEGER NOT NULL,
    {{spaces.display_order}} INTEGER NOT NULL,
    {{spaces.background_image_updated_at}} INTEGER NOT NULL,
    CONSTRAINT fk_spaces_space_group FOREIGN KEY ({{spaces.space_group_id}}) REFERENCES {{space_groups.table}}({{space_groups.id}}) ON DELETE CASCADE,
    CONSTRAINT uq_spaces_space_group_display_order UNIQUE ({{spaces.space_group_id}}, {{spaces.display_order}}),
    CONSTRAINT uq_spaces_id_space_group_id UNIQUE ({{spaces.id}}, {{spaces.space_group_id}})
) STRICT;

CREATE INDEX IF NOT EXISTS ix_spaces_space_group_id ON {{spaces.table}}({{spaces.space_group_id}});

CREATE TABLE IF NOT EXISTS {{space_background_images.table}} (
    {{space_background_images.space_id}} INTEGER PRIMARY KEY,
    {{space_background_images.content_type}} TEXT NOT NULL,
    {{space_background_images.data}} BLOB NOT NULL,
    CONSTRAINT fk_space_background_images_space FOREIGN KEY ({{space_background_images.space_id}}) REFERENCES {{spaces.table}}({{spaces.id}}) ON DELETE CASCADE
) STRICT;

CREATE TABLE IF NOT EXISTS {{breaker_groups.table}} (
    {{breaker_groups.id}} INTEGER PRIMARY KEY AUTOINCREMENT,
    {{breaker_groups.name}} TEXT NOT NULL CHECK (length({{breaker_groups.name}}) <= 255),
    {{breaker_groups.space_id}} INTEGER NOT NULL,
    {{breaker_groups.space_group_id}} INTEGER NOT NULL,
    CONSTRAINT fk_breaker_groups_space FOREIGN KEY ({{breaker_groups.space_id}}, {{breaker_groups.space_group_id}}) REFERENCES {{spaces.table}}({{spaces.id}}, {{spaces.space_group_id}}) ON DELETE CASCADE,
    CONSTRAINT uq_breaker_groups_id_space_group_id UNIQUE ({{breaker_groups.id}}, {{breaker_groups.space_group_id}})
) STRICT;

CREATE INDEX IF NOT EXISTS ix_breaker_groups_space_id ON {{breaker_groups.table}}({{breaker_groups.space_id}});

CREATE TABLE IF NOT EXISTS {{breaker_group_positions.table}} (
    {{breaker_group_positions.breaker_group_id}} INTEGER PRIMARY KEY,
    {{breaker_group_positions.position_x}} INTEGER NOT NULL,
    {{breaker_group_positions.position_y}} INTEGER NOT NULL,
    {{breaker_group_positions.position_z}} INTEGER NOT NULL,
    CONSTRAINT fk_breaker_group_positions_breaker_group FOREIGN KEY ({{breaker_group_positions.breaker_group_id}}) REFERENCES {{breaker_groups.table}}({{breaker_groups.id}}) ON DELETE CASCADE
) STRICT;

CREATE TABLE IF NOT EXISTS {{breakers.table}} (
    {{breakers.id}} INTEGER PRIMARY KEY AUTOINCREMENT,
    {{breakers.name}} TEXT NOT NULL CHECK (length({{breakers.name}}) <= 255),
    {{breakers.breaker_group_id}} INTEGER NOT NULL,
    {{breakers.space_group_id}} INTEGER NOT NULL,
    {{breakers.display_order}} INTEGER NOT NULL,
    {{breakers.upstream_breaker_id}} INTEGER,
    CONSTRAINT fk_breakers_breaker_group FOREIGN KEY ({{breakers.breaker_group_id}}, {{breakers.space_group_id}}) REFERENCES {{breaker_groups.table}}({{breaker_groups.id}}, {{breaker_groups.space_group_id}}) ON DELETE CASCADE,
    CONSTRAINT uq_breakers_breaker_group_display_order UNIQUE ({{breakers.breaker_group_id}}, {{breakers.display_order}}),
    CONSTRAINT uq_breakers_id_breaker_group_id UNIQUE ({{breakers.id}}, {{breakers.breaker_group_id}}),
    CONSTRAINT fk_breakers_upstream_breaker FOREIGN KEY ({{breakers.upstream_breaker_id}}) REFERENCES {{breakers.table}}({{breakers.id}}) ON DELETE SET NULL,
    CONSTRAINT chk_breakers_not_self_upstream CHECK (
        {{breakers.upstream_breaker_id}} IS NULL
        OR {{breakers.upstream_breaker_id}} <> {{breakers.id}}
    )
) STRICT;

CREATE INDEX IF NOT EXISTS ix_breakers_breaker_group_id ON {{breakers.table}}({{breakers.breaker_group_id}});

CREATE INDEX IF NOT EXISTS ix_breakers_upstream_breaker_id ON {{breakers.table}}({{breakers.upstream_breaker_id}});

CREATE TABLE IF NOT EXISTS {{devices.table}} (
    {{devices.id}} INTEGER PRIMARY KEY AUTOINCREMENT,
    {{devices.name}} TEXT NOT NULL CHECK (length({{devices.name}}) <= 255),
    {{devices.space_id}} INTEGER NOT NULL,
    CONSTRAINT fk_devices_space FOREIGN KEY ({{devices.space_id}}) REFERENCES {{spaces.table}}({{spaces.id}}) ON DELETE CASCADE
) STRICT;

CREATE TABLE IF NOT EXISTS {{device_positions.table}} (
    {{device_positions.device_id}} INTEGER PRIMARY KEY,
    {{device_positions.position_x}} INTEGER NOT NULL,
    {{device_positions.position_y}} INTEGER NOT NULL,
    {{device_positions.position_z}} INTEGER NOT NULL,
    CONSTRAINT fk_device_positions_device FOREIGN KEY ({{device_positions.device_id}}) REFERENCES {{devices.table}}({{devices.id}}) ON DELETE CASCADE
) STRICT;

CREATE INDEX IF NOT EXISTS ix_devices_space_id ON {{devices.table}}({{devices.space_id}});

CREATE TABLE IF NOT EXISTS {{device_breakers.table}} (
    {{device_breakers.device_id}} INTEGER NOT NULL,
    {{device_breakers.breaker_id}} INTEGER NOT NULL,
    CONSTRAINT pk_device_breakers PRIMARY KEY ({{device_breakers.device_id}}, {{device_breakers.breaker_id}}),
    CONSTRAINT fk_device_breakers_device FOREIGN KEY ({{device_breakers.device_id}}) REFERENCES {{devices.table}}({{devices.id}}) ON DELETE CASCADE,
    CONSTRAINT fk_device_breakers_breaker FOREIGN KEY ({{device_breakers.breaker_id}}) REFERENCES {{breakers.table}}({{breakers.id}}) ON DELETE CASCADE
) STRICT;

CREATE INDEX IF NOT EXISTS ix_device_breakers_breaker_id ON {{device_breakers.table}}({{device_breakers.breaker_id}});
