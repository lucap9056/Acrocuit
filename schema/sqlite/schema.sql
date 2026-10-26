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

CREATE TABLE IF NOT EXISTS {{switch_panels.table}} (
    {{switch_panels.id}} INTEGER PRIMARY KEY AUTOINCREMENT,
    {{switch_panels.name}} TEXT NOT NULL CHECK (length({{switch_panels.name}}) <= 255),
    {{switch_panels.space_id}} INTEGER NOT NULL,
    {{switch_panels.space_group_id}} INTEGER NOT NULL,
    {{switch_panels.attributes}} TEXT NOT NULL DEFAULT '{}',
    CONSTRAINT fk_switch_panels_space FOREIGN KEY ({{switch_panels.space_id}}, {{switch_panels.space_group_id}}) REFERENCES {{spaces.table}}({{spaces.id}}, {{spaces.space_group_id}}) ON DELETE CASCADE,
    CONSTRAINT uq_switch_panels_id_space_group_id UNIQUE ({{switch_panels.id}}, {{switch_panels.space_group_id}}),
    CONSTRAINT chk_switch_panels_attributes CHECK (
        json_valid({{switch_panels.attributes}})
        AND json_type({{switch_panels.attributes}}) = 'object'
        AND length(CAST({{switch_panels.attributes}} AS BLOB)) <= {{attributes_max_bytes}}
    )
) STRICT;

CREATE INDEX IF NOT EXISTS ix_switch_panels_space_id ON {{switch_panels.table}}({{switch_panels.space_id}});

CREATE TABLE IF NOT EXISTS {{switch_panel_positions.table}} (
    {{switch_panel_positions.switch_panel_id}} INTEGER PRIMARY KEY,
    {{switch_panel_positions.position_x}} INTEGER NOT NULL,
    {{switch_panel_positions.position_y}} INTEGER NOT NULL,
    {{switch_panel_positions.position_z}} INTEGER NOT NULL,
    CONSTRAINT fk_switch_panel_positions_switch_panel FOREIGN KEY ({{switch_panel_positions.switch_panel_id}}) REFERENCES {{switch_panels.table}}({{switch_panels.id}}) ON DELETE CASCADE
) STRICT;

CREATE TABLE IF NOT EXISTS {{switches.table}} (
    {{switches.id}} INTEGER PRIMARY KEY AUTOINCREMENT,
    {{switches.name}} TEXT NOT NULL CHECK (length({{switches.name}}) <= 255),
    {{switches.switch_panel_id}} INTEGER NOT NULL,
    {{switches.space_group_id}} INTEGER NOT NULL,
    {{switches.display_order}} INTEGER NOT NULL,
    {{switches.attributes}} TEXT NOT NULL DEFAULT '{}',
    CONSTRAINT fk_switches_switch_panel FOREIGN KEY ({{switches.switch_panel_id}}, {{switches.space_group_id}}) REFERENCES {{switch_panels.table}}({{switch_panels.id}}, {{switch_panels.space_group_id}}) ON DELETE CASCADE,
    CONSTRAINT uq_switches_switch_panel_display_order UNIQUE ({{switches.switch_panel_id}}, {{switches.display_order}}),
    CONSTRAINT uq_switches_id_switch_panel_id UNIQUE ({{switches.id}}, {{switches.switch_panel_id}}),
    CONSTRAINT uq_switches_id_space_group_id UNIQUE ({{switches.id}}, {{switches.space_group_id}}),
    CONSTRAINT chk_switches_attributes CHECK (
        json_valid({{switches.attributes}})
        AND json_type({{switches.attributes}}) = 'object'
        AND length(CAST({{switches.attributes}} AS BLOB)) <= {{attributes_max_bytes}}
    )
) STRICT;

CREATE INDEX IF NOT EXISTS ix_switches_switch_panel_id ON {{switches.table}}({{switches.switch_panel_id}});

CREATE TABLE IF NOT EXISTS {{switch_links.table}} (
    {{switch_links.upstream_switch_id}} INTEGER NOT NULL,
    {{switch_links.downstream_switch_id}} INTEGER NOT NULL,
    {{switch_links.space_group_id}} INTEGER NOT NULL,
    {{switch_links.attributes}} TEXT NOT NULL DEFAULT '{}',
    CONSTRAINT pk_switch_links PRIMARY KEY ({{switch_links.downstream_switch_id}}, {{switch_links.upstream_switch_id}}),
    CONSTRAINT fk_switch_links_upstream_switch FOREIGN KEY ({{switch_links.upstream_switch_id}}, {{switch_links.space_group_id}}) REFERENCES {{switches.table}}({{switches.id}}, {{switches.space_group_id}}) ON DELETE CASCADE,
    CONSTRAINT fk_switch_links_downstream_switch FOREIGN KEY ({{switch_links.downstream_switch_id}}, {{switch_links.space_group_id}}) REFERENCES {{switches.table}}({{switches.id}}, {{switches.space_group_id}}) ON DELETE CASCADE,
    CONSTRAINT chk_switch_links_not_self CHECK ({{switch_links.upstream_switch_id}} <> {{switch_links.downstream_switch_id}}),
    CONSTRAINT chk_switch_links_attributes CHECK (
        json_valid({{switch_links.attributes}})
        AND json_type({{switch_links.attributes}}) = 'object'
        AND length(CAST({{switch_links.attributes}} AS BLOB)) <= {{attributes_max_bytes}}
    )
) STRICT;

CREATE INDEX IF NOT EXISTS ix_switch_links_upstream_switch_id ON {{switch_links.table}}({{switch_links.upstream_switch_id}});

CREATE TABLE IF NOT EXISTS {{devices.table}} (
    {{devices.id}} INTEGER PRIMARY KEY AUTOINCREMENT,
    {{devices.name}} TEXT NOT NULL CHECK (length({{devices.name}}) <= 255),
    {{devices.space_id}} INTEGER NOT NULL,
    {{devices.attributes}} TEXT NOT NULL DEFAULT '{}',
    CONSTRAINT fk_devices_space FOREIGN KEY ({{devices.space_id}}) REFERENCES {{spaces.table}}({{spaces.id}}) ON DELETE CASCADE,
    CONSTRAINT chk_devices_attributes CHECK (
        json_valid({{devices.attributes}})
        AND json_type({{devices.attributes}}) = 'object'
        AND length(CAST({{devices.attributes}} AS BLOB)) <= {{attributes_max_bytes}}
    )
) STRICT;

CREATE TABLE IF NOT EXISTS {{device_positions.table}} (
    {{device_positions.device_id}} INTEGER PRIMARY KEY,
    {{device_positions.position_x}} INTEGER NOT NULL,
    {{device_positions.position_y}} INTEGER NOT NULL,
    {{device_positions.position_z}} INTEGER NOT NULL,
    CONSTRAINT fk_device_positions_device FOREIGN KEY ({{device_positions.device_id}}) REFERENCES {{devices.table}}({{devices.id}}) ON DELETE CASCADE
) STRICT;

CREATE INDEX IF NOT EXISTS ix_devices_space_id ON {{devices.table}}({{devices.space_id}});

CREATE TABLE IF NOT EXISTS {{device_switches.table}} (
    {{device_switches.device_id}} INTEGER NOT NULL,
    {{device_switches.switch_id}} INTEGER NOT NULL,
    {{device_switches.attributes}} TEXT NOT NULL DEFAULT '{}',
    CONSTRAINT pk_device_switches PRIMARY KEY ({{device_switches.device_id}}, {{device_switches.switch_id}}),
    CONSTRAINT fk_device_switches_device FOREIGN KEY ({{device_switches.device_id}}) REFERENCES {{devices.table}}({{devices.id}}) ON DELETE CASCADE,
    CONSTRAINT fk_device_switches_switch FOREIGN KEY ({{device_switches.switch_id}}) REFERENCES {{switches.table}}({{switches.id}}) ON DELETE CASCADE,
    CONSTRAINT chk_device_switches_attributes CHECK (
        json_valid({{device_switches.attributes}})
        AND json_type({{device_switches.attributes}}) = 'object'
        AND length(CAST({{device_switches.attributes}} AS BLOB)) <= {{attributes_max_bytes}}
    )
) STRICT;

CREATE INDEX IF NOT EXISTS ix_device_switches_switch_id ON {{device_switches.table}}({{device_switches.switch_id}});
