CREATE TABLE IF NOT EXISTS {{space_groups.table}} (
    {{space_groups.id}} INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    {{space_groups.name}} VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS {{spaces.table}} (
    {{spaces.id}} INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    {{spaces.name}} VARCHAR(255) NOT NULL,
    {{spaces.space_group_id}} INT NOT NULL,
    {{spaces.display_order}} INT NOT NULL,
    {{spaces.background_image_updated_at}} TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_spaces_space_group FOREIGN KEY ({{spaces.space_group_id}}) REFERENCES {{space_groups.table}}({{space_groups.id}}) ON DELETE CASCADE,
    CONSTRAINT uq_spaces_space_group_display_order UNIQUE ({{spaces.space_group_id}}, {{spaces.display_order}}) DEFERRABLE INITIALLY IMMEDIATE,
    CONSTRAINT uq_spaces_id_space_group_id UNIQUE ({{spaces.id}}, {{spaces.space_group_id}})
);

CREATE INDEX IF NOT EXISTS ix_spaces_space_group_id ON {{spaces.table}}({{spaces.space_group_id}});

CREATE TABLE IF NOT EXISTS {{breaker_groups.table}} (
    {{breaker_groups.id}} INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    {{breaker_groups.name}} VARCHAR(255) NOT NULL,
    {{breaker_groups.space_id}} INT NOT NULL,
    {{breaker_groups.space_group_id}} INT NOT NULL,
    CONSTRAINT fk_breaker_groups_space FOREIGN KEY ({{breaker_groups.space_id}}, {{breaker_groups.space_group_id}}) REFERENCES {{spaces.table}}({{spaces.id}}, {{spaces.space_group_id}}) ON DELETE CASCADE,
    CONSTRAINT uq_breaker_groups_id_space_group_id UNIQUE ({{breaker_groups.id}}, {{breaker_groups.space_group_id}})
);

CREATE INDEX IF NOT EXISTS ix_breaker_groups_space_id ON {{breaker_groups.table}}({{breaker_groups.space_id}});

CREATE TABLE IF NOT EXISTS {{breaker_group_positions.table}} (
    {{breaker_group_positions.breaker_group_id}} INT PRIMARY KEY,
    {{breaker_group_positions.position_x}} INT NOT NULL,
    {{breaker_group_positions.position_y}} INT NOT NULL,
    {{breaker_group_positions.position_z}} INT NOT NULL,
    CONSTRAINT fk_breaker_group_positions_breaker_group FOREIGN KEY ({{breaker_group_positions.breaker_group_id}}) REFERENCES {{breaker_groups.table}}({{breaker_groups.id}}) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS {{breakers.table}} (
    {{breakers.id}} INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    {{breakers.name}} VARCHAR(255) NOT NULL,
    {{breakers.breaker_group_id}} INT NOT NULL,
    {{breakers.space_group_id}} INT NOT NULL,
    {{breakers.display_order}} INT NOT NULL,
    {{breakers.upstream_breaker_id}} INT,
    CONSTRAINT fk_breakers_breaker_group FOREIGN KEY ({{breakers.breaker_group_id}}, {{breakers.space_group_id}}) REFERENCES {{breaker_groups.table}}({{breaker_groups.id}}, {{breaker_groups.space_group_id}}) ON DELETE CASCADE,
    CONSTRAINT uq_breakers_breaker_group_display_order UNIQUE ({{breakers.breaker_group_id}}, {{breakers.display_order}}) DEFERRABLE INITIALLY IMMEDIATE,
    CONSTRAINT uq_breakers_id_breaker_group_id UNIQUE ({{breakers.id}}, {{breakers.breaker_group_id}}),
    CONSTRAINT fk_breakers_upstream_breaker FOREIGN KEY ({{breakers.upstream_breaker_id}}) REFERENCES {{breakers.table}}({{breakers.id}}) ON DELETE SET NULL,
    CONSTRAINT chk_breakers_not_self_upstream CHECK (
        {{breakers.upstream_breaker_id}} IS NULL
        OR {{breakers.upstream_breaker_id}} <> {{breakers.id}}
    )
);

CREATE INDEX IF NOT EXISTS ix_breakers_breaker_group_id ON {{breakers.table}}({{breakers.breaker_group_id}});

CREATE INDEX IF NOT EXISTS ix_breakers_upstream_breaker_id ON {{breakers.table}}({{breakers.upstream_breaker_id}});

CREATE TABLE IF NOT EXISTS {{devices.table}} (
    {{devices.id}} INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    {{devices.name}} VARCHAR(255) NOT NULL,
    {{devices.space_id}} INT NOT NULL,
    CONSTRAINT fk_devices_space FOREIGN KEY ({{devices.space_id}}) REFERENCES {{spaces.table}}({{spaces.id}}) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS {{device_positions.table}} (
    {{device_positions.device_id}} INT PRIMARY KEY,
    {{device_positions.position_x}} INT NOT NULL,
    {{device_positions.position_y}} INT NOT NULL,
    {{device_positions.position_z}} INT NOT NULL,
    CONSTRAINT fk_device_positions_device FOREIGN KEY ({{device_positions.device_id}}) REFERENCES {{devices.table}}({{devices.id}}) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS ix_devices_space_id ON {{devices.table}}({{devices.space_id}});

CREATE TABLE IF NOT EXISTS {{device_breakers.table}} (
    {{device_breakers.device_id}} INT NOT NULL,
    {{device_breakers.breaker_id}} INT NOT NULL,
    CONSTRAINT pk_device_breakers PRIMARY KEY ({{device_breakers.device_id}}, {{device_breakers.breaker_id}}),
    CONSTRAINT fk_device_breakers_device FOREIGN KEY ({{device_breakers.device_id}}) REFERENCES {{devices.table}}({{devices.id}}) ON DELETE CASCADE,
    CONSTRAINT fk_device_breakers_breaker FOREIGN KEY ({{device_breakers.breaker_id}}) REFERENCES {{breakers.table}}({{breakers.id}}) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS ix_device_breakers_breaker_id ON {{device_breakers.table}}({{device_breakers.breaker_id}});

CREATE TABLE IF NOT EXISTS {{users.table}} (
    {{users.id}} INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    {{users.username}} VARCHAR(127) NOT NULL,
    CONSTRAINT uq_users_username UNIQUE ({{users.username}})
);

CREATE TABLE IF NOT EXISTS {{user_secrets.table}} (
    {{user_secrets.user_id}} INT PRIMARY KEY,
    {{user_secrets.password_salt}} BYTEA NOT NULL,
    {{user_secrets.password_hash}} BYTEA NOT NULL,
    CONSTRAINT fk_user_secrets_user FOREIGN KEY ({{user_secrets.user_id}}) REFERENCES {{users.table}}({{users.id}}) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS {{space_group_owners.table}} (
    {{space_group_owners.user_id}} INT NOT NULL,
    {{space_group_owners.space_group_id}} INT NOT NULL,
    CONSTRAINT pk_space_group_owners PRIMARY KEY ({{space_group_owners.user_id}}, {{space_group_owners.space_group_id}}),
    CONSTRAINT fk_space_group_owners_user FOREIGN KEY ({{space_group_owners.user_id}}) REFERENCES {{users.table}}({{users.id}}) ON DELETE CASCADE,
    CONSTRAINT fk_space_group_owners_space_group FOREIGN KEY ({{space_group_owners.space_group_id}}) REFERENCES {{space_groups.table}}({{space_groups.id}}) ON DELETE CASCADE
);
