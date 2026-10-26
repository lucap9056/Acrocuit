DROP FUNCTION IF EXISTS {{fn_set_space.func}}(TEXT, INT, VARCHAR, INT);
DROP FUNCTION IF EXISTS {{fn_set_space.func}}(TEXT, INT, VARCHAR);

CREATE OR REPLACE FUNCTION {{fn_set_space.func}}(
    p_user_email TEXT,
    p_space_id INT,
    p_name VARCHAR(255) DEFAULT NULL
) RETURNS TABLE (
    {{fn_set_space.o_id}} INT,
    {{fn_set_space.o_name}} VARCHAR(255),
    {{fn_set_space.o_space_group_id}} INT,
    {{fn_set_space.o_display_order}} INT,
    {{fn_set_space.o_background_image_updated_at}} BIGINT
) LANGUAGE plpgsql AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM {{spaces.table}} s
        JOIN {{space_group_owners.table}} so ON so.{{space_group_owners.space_group_id}} = s.{{spaces.space_group_id}}
        WHERE s.{{spaces.id}} = p_space_id AND so.{{space_group_owners.user_email}} = p_user_email
    ) THEN
        RETURN;
    END IF;

    IF p_name IS NOT NULL THEN
        UPDATE {{spaces.table}} SET {{spaces.name}} = p_name WHERE {{spaces.id}} = p_space_id;
    END IF;

    RETURN QUERY
    SELECT s.{{spaces.id}}, s.{{spaces.name}}, s.{{spaces.space_group_id}}, s.{{spaces.display_order}}, s.{{spaces.background_image_updated_at}}
    FROM {{spaces.table}} s
    WHERE s.{{spaces.id}} = p_space_id;
END;
$$;

DROP FUNCTION IF EXISTS {{fn_move_space.func}}(TEXT, INT, INT);

CREATE OR REPLACE FUNCTION {{fn_move_space.func}}(
    p_user_email TEXT,
    p_space_id INT,
    p_display_order INT
) RETURNS TABLE (
    {{fn_move_space.o_id}} INT,
    {{fn_move_space.o_name}} VARCHAR(255),
    {{fn_move_space.o_space_group_id}} INT,
    {{fn_move_space.o_display_order}} INT,
    {{fn_move_space.o_background_image_updated_at}} BIGINT
) LANGUAGE plpgsql AS $$
DECLARE
    v_space_group_id INT;
    v_current_order INT;
    v_delta INT;
BEGIN
    SELECT s.{{spaces.space_group_id}}, s.{{spaces.display_order}}
    INTO v_space_group_id, v_current_order
    FROM {{spaces.table}} s
    JOIN {{space_groups.table}} sg ON sg.{{space_groups.id}} = s.{{spaces.space_group_id}}
    JOIN {{space_group_owners.table}} so ON so.{{space_group_owners.space_group_id}} = sg.{{space_groups.id}}
    WHERE s.{{spaces.id}} = p_space_id AND so.{{space_group_owners.user_email}} = p_user_email
    FOR UPDATE OF sg;

    IF v_space_group_id IS NULL THEN
        RETURN;
    END IF;

    IF p_display_order <> v_current_order THEN
        v_delta := p_display_order - v_current_order;

        IF ABS(v_delta) = 1 THEN
            UPDATE {{spaces.table}}
            SET {{spaces.display_order}} = CASE WHEN {{spaces.id}} = p_space_id THEN p_display_order ELSE v_current_order END
            WHERE {{spaces.space_group_id}} = v_space_group_id AND ({{spaces.id}} = p_space_id OR {{spaces.display_order}} = p_display_order);
        ELSIF v_delta > 0 THEN
            UPDATE {{spaces.table}}
            SET {{spaces.display_order}} = CASE WHEN {{spaces.id}} = p_space_id THEN p_display_order ELSE {{spaces.display_order}} - 1 END
            WHERE {{spaces.space_group_id}} = v_space_group_id AND {{spaces.display_order}} >= v_current_order AND {{spaces.display_order}} <= p_display_order;
        ELSE
            UPDATE {{spaces.table}}
            SET {{spaces.display_order}} = CASE WHEN {{spaces.id}} = p_space_id THEN p_display_order ELSE {{spaces.display_order}} + 1 END
            WHERE {{spaces.space_group_id}} = v_space_group_id AND {{spaces.display_order}} >= p_display_order AND {{spaces.display_order}} <= v_current_order;
        END IF;
    END IF;

    RETURN QUERY
    SELECT s.{{spaces.id}}, s.{{spaces.name}}, s.{{spaces.space_group_id}}, s.{{spaces.display_order}}, s.{{spaces.background_image_updated_at}}
    FROM {{spaces.table}} s
    WHERE s.{{spaces.space_group_id}} = v_space_group_id
    ORDER BY s.{{spaces.display_order}};
END;
$$;

DROP FUNCTION IF EXISTS {{fn_set_switch.func}}(TEXT, INT, VARCHAR, INT);
DROP FUNCTION IF EXISTS {{fn_set_switch.func}}(TEXT, INT, VARCHAR);
DROP FUNCTION IF EXISTS {{fn_set_switch.func}}(TEXT, INT, VARCHAR, JSON);

CREATE OR REPLACE FUNCTION {{fn_set_switch.func}}(
    p_user_email TEXT,
    p_switch_id INT,
    p_name VARCHAR(255) DEFAULT NULL,
    p_attributes JSON DEFAULT NULL
) RETURNS TABLE (
    {{fn_set_switch.o_id}} INT,
    {{fn_set_switch.o_name}} VARCHAR(255),
    {{fn_set_switch.o_switch_panel_id}} INT,
    {{fn_set_switch.o_space_group_id}} INT,
    {{fn_set_switch.o_display_order}} INT,
    {{fn_set_switch.o_attributes}} JSON
) LANGUAGE plpgsql AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM {{switches.table}} b
        JOIN {{space_group_owners.table}} so ON so.{{space_group_owners.space_group_id}} = b.{{switches.space_group_id}}
        WHERE b.{{switches.id}} = p_switch_id AND so.{{space_group_owners.user_email}} = p_user_email
    ) THEN
        RETURN;
    END IF;

    UPDATE {{switches.table}}
    SET {{switches.name}} = COALESCE(p_name, {{switches.name}}),
        {{switches.attributes}} = COALESCE(p_attributes, {{switches.attributes}})
    WHERE {{switches.id}} = p_switch_id;

    RETURN QUERY
    SELECT b.{{switches.id}}, b.{{switches.name}}, b.{{switches.switch_panel_id}}, b.{{switches.space_group_id}}, b.{{switches.display_order}}, b.{{switches.attributes}}
    FROM {{switches.table}} b
    WHERE b.{{switches.id}} = p_switch_id;
END;
$$;

DROP FUNCTION IF EXISTS {{fn_move_switch.func}}(TEXT, INT, INT);

CREATE OR REPLACE FUNCTION {{fn_move_switch.func}}(
    p_user_email TEXT,
    p_switch_id INT,
    p_display_order INT
) RETURNS TABLE (
    {{fn_move_switch.o_id}} INT,
    {{fn_move_switch.o_name}} VARCHAR(255),
    {{fn_move_switch.o_switch_panel_id}} INT,
    {{fn_move_switch.o_space_group_id}} INT,
    {{fn_move_switch.o_display_order}} INT,
    {{fn_move_switch.o_attributes}} JSON
) LANGUAGE plpgsql AS $$
DECLARE
    v_switch_panel_id INT;
    v_current_order INT;
    v_delta INT;
BEGIN
    SELECT b.{{switches.switch_panel_id}}, b.{{switches.display_order}}
    INTO v_switch_panel_id, v_current_order
    FROM {{switches.table}} b
    JOIN {{switch_panels.table}} bg ON bg.{{switch_panels.id}} = b.{{switches.switch_panel_id}}
    JOIN {{space_group_owners.table}} so ON so.{{space_group_owners.space_group_id}} = bg.{{switch_panels.space_group_id}}
    WHERE b.{{switches.id}} = p_switch_id AND so.{{space_group_owners.user_email}} = p_user_email
    FOR UPDATE OF bg;

    IF v_switch_panel_id IS NULL THEN
        RETURN;
    END IF;

    IF p_display_order <> v_current_order THEN
        v_delta := p_display_order - v_current_order;

        IF ABS(v_delta) = 1 THEN
            UPDATE {{switches.table}}
            SET {{switches.display_order}} = CASE WHEN {{switches.id}} = p_switch_id THEN p_display_order ELSE v_current_order END
            WHERE {{switches.switch_panel_id}} = v_switch_panel_id AND ({{switches.id}} = p_switch_id OR {{switches.display_order}} = p_display_order);
        ELSIF v_delta > 0 THEN
            UPDATE {{switches.table}}
            SET {{switches.display_order}} = CASE WHEN {{switches.id}} = p_switch_id THEN p_display_order ELSE {{switches.display_order}} - 1 END
            WHERE {{switches.switch_panel_id}} = v_switch_panel_id AND {{switches.display_order}} >= v_current_order AND {{switches.display_order}} <= p_display_order;
        ELSE
            UPDATE {{switches.table}}
            SET {{switches.display_order}} = CASE WHEN {{switches.id}} = p_switch_id THEN p_display_order ELSE {{switches.display_order}} + 1 END
            WHERE {{switches.switch_panel_id}} = v_switch_panel_id AND {{switches.display_order}} >= p_display_order AND {{switches.display_order}} <= v_current_order;
        END IF;
    END IF;

    RETURN QUERY
    SELECT b.{{switches.id}}, b.{{switches.name}}, b.{{switches.switch_panel_id}}, b.{{switches.space_group_id}}, b.{{switches.display_order}}, b.{{switches.attributes}}
    FROM {{switches.table}} b
    WHERE b.{{switches.switch_panel_id}} = v_switch_panel_id
    ORDER BY b.{{switches.display_order}};
END;
$$;

DROP FUNCTION IF EXISTS {{fn_add_switch_upstream.func}}(TEXT, INT, INT, JSON);

CREATE OR REPLACE FUNCTION {{fn_add_switch_upstream.func}}(
    p_user_email TEXT,
    p_switch_id INT,
    p_upstream_switch_id INT,
    p_attributes JSON
) RETURNS TABLE (
    {{fn_add_switch_upstream.o_upstream_switch_id}} INT,
    {{fn_add_switch_upstream.o_downstream_switch_id}} INT,
    {{fn_add_switch_upstream.o_attributes}} JSON
) LANGUAGE plpgsql AS $$
DECLARE
    v_space_group_id INT;
    v_has_cycle BOOLEAN;
BEGIN
    SELECT b.{{switches.space_group_id}}
    INTO v_space_group_id
    FROM {{switches.table}} b
    JOIN {{space_groups.table}} sg ON sg.{{space_groups.id}} = b.{{switches.space_group_id}}
    JOIN {{space_group_owners.table}} so ON so.{{space_group_owners.space_group_id}} = sg.{{space_groups.id}}
    WHERE b.{{switches.id}} = p_switch_id AND so.{{space_group_owners.user_email}} = p_user_email
    FOR UPDATE OF sg;

    IF v_space_group_id IS NULL THEN
        RETURN;
    END IF;

    IF p_upstream_switch_id = p_switch_id THEN
        RAISE EXCEPTION 'switch % cannot reference itself as upstream', p_switch_id
            USING ERRCODE = '{{fn_add_switch_upstream.err_self_reference}}';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM {{switches.table}}
        WHERE {{switches.id}} = p_upstream_switch_id AND {{switches.space_group_id}} = v_space_group_id
    ) THEN
        RETURN;
    END IF;

    WITH RECURSIVE upstream_chain(id) AS (
        SELECT p_upstream_switch_id

        UNION

        SELECT l.{{switch_links.upstream_switch_id}}
        FROM {{switch_links.table}} l
        JOIN upstream_chain c ON l.{{switch_links.downstream_switch_id}} = c.id
    )
    SELECT EXISTS (SELECT 1 FROM upstream_chain WHERE id = p_switch_id) INTO v_has_cycle;

    IF v_has_cycle THEN
        RAISE EXCEPTION 'upstream switch % would create a cycle for switch %', p_upstream_switch_id, p_switch_id
            USING ERRCODE = '{{fn_add_switch_upstream.err_cycle_detected}}';
    END IF;

    RETURN QUERY
    WITH inserted AS (
        INSERT INTO {{switch_links.table}} ({{switch_links.upstream_switch_id}}, {{switch_links.downstream_switch_id}}, {{switch_links.space_group_id}}, {{switch_links.attributes}})
        VALUES (p_upstream_switch_id, p_switch_id, v_space_group_id, p_attributes)
        RETURNING {{switch_links.upstream_switch_id}}, {{switch_links.downstream_switch_id}}, {{switch_links.attributes}}
    )
    SELECT * FROM inserted;
END;
$$;

DROP FUNCTION IF EXISTS {{fn_add_device_switch.func}}(TEXT, INT, INT);

CREATE OR REPLACE FUNCTION {{fn_add_device_switch.func}}(
    p_user_email TEXT,
    p_device_id INT,
    p_switch_id INT,
    p_attributes JSON
) RETURNS TABLE (
    {{fn_add_device_switch.o_result}} INT
) LANGUAGE plpgsql AS $$
DECLARE
    v_space_group_id INT;
BEGIN
    SELECT s.{{spaces.space_group_id}}
    INTO v_space_group_id
    FROM {{devices.table}} d
    JOIN {{spaces.table}} s ON s.{{spaces.id}} = d.{{devices.space_id}}
    JOIN {{space_group_owners.table}} so ON so.{{space_group_owners.space_group_id}} = s.{{spaces.space_group_id}}
    WHERE d.{{devices.id}} = p_device_id AND so.{{space_group_owners.user_email}} = p_user_email;

    IF v_space_group_id IS NULL THEN
        RETURN QUERY SELECT 1;
        RETURN;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM {{switches.table}}
        WHERE {{switches.id}} = p_switch_id AND {{switches.space_group_id}} = v_space_group_id
    ) THEN
        RETURN QUERY SELECT 2;
        RETURN;
    END IF;

    IF EXISTS (
        SELECT 1 FROM {{device_switches.table}}
        WHERE {{device_switches.device_id}} = p_device_id AND {{device_switches.switch_id}} = p_switch_id
    ) THEN
        RETURN QUERY SELECT 3;
        RETURN;
    END IF;

    INSERT INTO {{device_switches.table}} ({{device_switches.device_id}}, {{device_switches.switch_id}}, {{device_switches.attributes}})
    VALUES (p_device_id, p_switch_id, p_attributes);

    RETURN QUERY SELECT 0;
END;
$$;
