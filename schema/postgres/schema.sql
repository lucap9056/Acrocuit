CREATE TABLE IF NOT EXISTS {{space_groups.table}} (
    {{space_groups.id}} UUID PRIMARY KEY DEFAULT uuidv7(),
    {{space_groups.name}} VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS {{spaces.table}} (
    {{spaces.id}} UUID PRIMARY KEY DEFAULT uuidv7(),
    {{spaces.name}} VARCHAR(255) NOT NULL,
    {{spaces.space_group_id}} UUID NOT NULL,
    {{spaces.display_order}} INT NOT NULL,
    {{spaces.background_image_updated_at}} BIGINT NOT NULL,
    CONSTRAINT fk_spaces_space_group FOREIGN KEY ({{spaces.space_group_id}}) REFERENCES {{space_groups.table}}({{space_groups.id}}) ON DELETE CASCADE,
    CONSTRAINT uq_spaces_space_group_display_order UNIQUE ({{spaces.space_group_id}}, {{spaces.display_order}}) DEFERRABLE INITIALLY IMMEDIATE,
    CONSTRAINT uq_spaces_id_space_group_id UNIQUE ({{spaces.id}}, {{spaces.space_group_id}})
);

CREATE INDEX IF NOT EXISTS ix_spaces_space_group_id ON {{spaces.table}}({{spaces.space_group_id}});

CREATE TABLE IF NOT EXISTS {{space_background_images.table}} (
    {{space_background_images.space_id}} UUID PRIMARY KEY,
    {{space_background_images.content_type}} TEXT NOT NULL,
    {{space_background_images.data}} BYTEA NOT NULL,
    CONSTRAINT fk_space_background_images_space FOREIGN KEY ({{space_background_images.space_id}}) REFERENCES {{spaces.table}}({{spaces.id}}) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS {{switch_panels.table}} (
    {{switch_panels.id}} UUID PRIMARY KEY DEFAULT uuidv7(),
    {{switch_panels.name}} VARCHAR(255) NOT NULL,
    {{switch_panels.space_id}} UUID NOT NULL,
    {{switch_panels.space_group_id}} UUID NOT NULL,
    {{switch_panels.attributes}} JSON NOT NULL DEFAULT '{}',
    CONSTRAINT fk_switch_panels_space FOREIGN KEY ({{switch_panels.space_id}}, {{switch_panels.space_group_id}}) REFERENCES {{spaces.table}}({{spaces.id}}, {{spaces.space_group_id}}) ON DELETE CASCADE,
    CONSTRAINT uq_switch_panels_id_space_group_id UNIQUE ({{switch_panels.id}}, {{switch_panels.space_group_id}}),
    CONSTRAINT chk_switch_panels_attributes CHECK (
        json_typeof({{switch_panels.attributes}}) = 'object'
        AND octet_length({{switch_panels.attributes}}::text) <= {{attributes_max_bytes}}
    )
);

CREATE INDEX IF NOT EXISTS ix_switch_panels_space_id ON {{switch_panels.table}}({{switch_panels.space_id}});

CREATE TABLE IF NOT EXISTS {{switch_panel_positions.table}} (
    {{switch_panel_positions.switch_panel_id}} UUID PRIMARY KEY,
    {{switch_panel_positions.position_x}} INT NOT NULL,
    {{switch_panel_positions.position_y}} INT NOT NULL,
    {{switch_panel_positions.position_z}} INT NOT NULL,
    CONSTRAINT fk_switch_panel_positions_switch_panel FOREIGN KEY ({{switch_panel_positions.switch_panel_id}}) REFERENCES {{switch_panels.table}}({{switch_panels.id}}) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS {{switches.table}} (
    {{switches.id}} UUID PRIMARY KEY DEFAULT uuidv7(),
    {{switches.name}} VARCHAR(255) NOT NULL,
    {{switches.switch_panel_id}} UUID NOT NULL,
    {{switches.space_group_id}} UUID NOT NULL,
    {{switches.display_order}} INT NOT NULL,
    {{switches.attributes}} JSON NOT NULL DEFAULT '{}',
    CONSTRAINT fk_switches_switch_panel FOREIGN KEY ({{switches.switch_panel_id}}, {{switches.space_group_id}}) REFERENCES {{switch_panels.table}}({{switch_panels.id}}, {{switch_panels.space_group_id}}) ON DELETE CASCADE,
    CONSTRAINT uq_switches_switch_panel_display_order UNIQUE ({{switches.switch_panel_id}}, {{switches.display_order}}) DEFERRABLE INITIALLY IMMEDIATE,
    CONSTRAINT uq_switches_id_switch_panel_id UNIQUE ({{switches.id}}, {{switches.switch_panel_id}}),
    CONSTRAINT uq_switches_id_space_group_id UNIQUE ({{switches.id}}, {{switches.space_group_id}}),
    CONSTRAINT chk_switches_attributes CHECK (
        json_typeof({{switches.attributes}}) = 'object'
        AND octet_length({{switches.attributes}}::text) <= {{attributes_max_bytes}}
    )
);

CREATE INDEX IF NOT EXISTS ix_switches_switch_panel_id ON {{switches.table}}({{switches.switch_panel_id}});

CREATE TABLE IF NOT EXISTS {{switch_links.table}} (
    {{switch_links.upstream_switch_id}} UUID NOT NULL,
    {{switch_links.downstream_switch_id}} UUID NOT NULL,
    {{switch_links.space_group_id}} UUID NOT NULL,
    {{switch_links.attributes}} JSON NOT NULL DEFAULT '{}',
    CONSTRAINT pk_switch_links PRIMARY KEY ({{switch_links.downstream_switch_id}}, {{switch_links.upstream_switch_id}}),
    CONSTRAINT fk_switch_links_upstream_switch FOREIGN KEY ({{switch_links.upstream_switch_id}}, {{switch_links.space_group_id}}) REFERENCES {{switches.table}}({{switches.id}}, {{switches.space_group_id}}) ON DELETE CASCADE,
    CONSTRAINT fk_switch_links_downstream_switch FOREIGN KEY ({{switch_links.downstream_switch_id}}, {{switch_links.space_group_id}}) REFERENCES {{switches.table}}({{switches.id}}, {{switches.space_group_id}}) ON DELETE CASCADE,
    CONSTRAINT chk_switch_links_not_self CHECK ({{switch_links.upstream_switch_id}} <> {{switch_links.downstream_switch_id}}),
    CONSTRAINT chk_switch_links_attributes CHECK (
        json_typeof({{switch_links.attributes}}) = 'object'
        AND octet_length({{switch_links.attributes}}::text) <= {{attributes_max_bytes}}
    )
);

CREATE INDEX IF NOT EXISTS ix_switch_links_upstream_switch_id ON {{switch_links.table}}({{switch_links.upstream_switch_id}});

CREATE TABLE IF NOT EXISTS {{devices.table}} (
    {{devices.id}} UUID PRIMARY KEY DEFAULT uuidv7(),
    {{devices.name}} VARCHAR(255) NOT NULL,
    {{devices.space_id}} UUID NOT NULL,
    {{devices.attributes}} JSON NOT NULL DEFAULT '{}',
    CONSTRAINT fk_devices_space FOREIGN KEY ({{devices.space_id}}) REFERENCES {{spaces.table}}({{spaces.id}}) ON DELETE CASCADE,
    CONSTRAINT chk_devices_attributes CHECK (
        json_typeof({{devices.attributes}}) = 'object'
        AND octet_length({{devices.attributes}}::text) <= {{attributes_max_bytes}}
    )
);

CREATE TABLE IF NOT EXISTS {{device_positions.table}} (
    {{device_positions.device_id}} UUID PRIMARY KEY,
    {{device_positions.position_x}} INT NOT NULL,
    {{device_positions.position_y}} INT NOT NULL,
    {{device_positions.position_z}} INT NOT NULL,
    CONSTRAINT fk_device_positions_device FOREIGN KEY ({{device_positions.device_id}}) REFERENCES {{devices.table}}({{devices.id}}) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS ix_devices_space_id ON {{devices.table}}({{devices.space_id}});

CREATE TABLE IF NOT EXISTS {{device_switches.table}} (
    {{device_switches.device_id}} UUID NOT NULL,
    {{device_switches.switch_id}} UUID NOT NULL,
    {{device_switches.attributes}} JSON NOT NULL DEFAULT '{}',
    CONSTRAINT pk_device_switches PRIMARY KEY ({{device_switches.device_id}}, {{device_switches.switch_id}}),
    CONSTRAINT fk_device_switches_device FOREIGN KEY ({{device_switches.device_id}}) REFERENCES {{devices.table}}({{devices.id}}) ON DELETE CASCADE,
    CONSTRAINT fk_device_switches_switch FOREIGN KEY ({{device_switches.switch_id}}) REFERENCES {{switches.table}}({{switches.id}}) ON DELETE CASCADE,
    CONSTRAINT chk_device_switches_attributes CHECK (
        json_typeof({{device_switches.attributes}}) = 'object'
        AND octet_length({{device_switches.attributes}}::text) <= {{attributes_max_bytes}}
    )
);

CREATE INDEX IF NOT EXISTS ix_device_switches_switch_id ON {{device_switches.table}}({{device_switches.switch_id}});

CREATE TABLE IF NOT EXISTS {{space_group_acls.table}} (
    {{space_group_acls.user_email}} TEXT NOT NULL,
    {{space_group_acls.space_group_id}} UUID NOT NULL,
    {{space_group_acls.role}} VARCHAR(50) NOT NULL DEFAULT '{{space_group_acls.viewer}}',
    CONSTRAINT chk_role_valid CHECK ({{space_group_acls.role}} IN ('{{space_group_acls.viewer}}', '{{space_group_acls.editor}}', '{{space_group_acls.owner}}')),
    CONSTRAINT pk_space_group_acls PRIMARY KEY ({{space_group_acls.user_email}}, {{space_group_acls.space_group_id}}),
    CONSTRAINT fk_space_group_acls_space_group FOREIGN KEY ({{space_group_acls.space_group_id}}) REFERENCES {{space_groups.table}}({{space_groups.id}}) ON DELETE CASCADE
);
