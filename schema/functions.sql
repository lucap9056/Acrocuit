DROP FUNCTION IF EXISTS {{fn_set_space.func}}(INT, INT, VARCHAR, INT);

CREATE OR REPLACE FUNCTION {{fn_set_space.func}}(
    p_user_id INT,
    p_space_id INT,
    p_name VARCHAR(255) DEFAULT NULL,
    p_display_order INT DEFAULT NULL
) RETURNS TABLE (
    {{fn_set_space.o_id}} INT,
    {{fn_set_space.o_name}} VARCHAR(255),
    {{fn_set_space.o_space_group_id}} INT,
    {{fn_set_space.o_display_order}} INT,
    {{fn_set_space.o_background_image_updated_at}} TIMESTAMPTZ
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
    WHERE s.{{spaces.id}} = p_space_id AND so.{{space_group_owners.user_id}} = p_user_id
    FOR UPDATE OF sg;

    IF v_space_group_id IS NULL THEN
        RETURN;
    END IF;

    IF p_display_order IS NOT NULL AND p_display_order <> v_current_order THEN
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

    IF p_name IS NOT NULL THEN
        UPDATE {{spaces.table}} SET {{spaces.name}} = p_name WHERE {{spaces.id}} = p_space_id;
    END IF;

    RETURN QUERY
    SELECT s.{{spaces.id}}, s.{{spaces.name}}, s.{{spaces.space_group_id}}, s.{{spaces.display_order}}, s.{{spaces.background_image_updated_at}}
    FROM {{spaces.table}} s
    WHERE s.{{spaces.id}} = p_space_id;
END;
$$;

DROP FUNCTION IF EXISTS {{fn_set_breaker.func}}(INT, INT, VARCHAR, INT);

CREATE OR REPLACE FUNCTION {{fn_set_breaker.func}}(
    p_user_id INT,
    p_breaker_id INT,
    p_name VARCHAR(255) DEFAULT NULL,
    p_display_order INT DEFAULT NULL
) RETURNS TABLE (
    {{fn_set_breaker.o_id}} INT,
    {{fn_set_breaker.o_name}} VARCHAR(255),
    {{fn_set_breaker.o_breaker_group_id}} INT,
    {{fn_set_breaker.o_space_group_id}} INT,
    {{fn_set_breaker.o_display_order}} INT,
    {{fn_set_breaker.o_upstream_breaker_id}} INT
) LANGUAGE plpgsql AS $$
DECLARE
    v_breaker_group_id INT;
    v_current_order INT;
    v_delta INT;
BEGIN
    SELECT b.{{breakers.breaker_group_id}}, b.{{breakers.display_order}}
    INTO v_breaker_group_id, v_current_order
    FROM {{breakers.table}} b
    JOIN {{breaker_groups.table}} bg ON bg.{{breaker_groups.id}} = b.{{breakers.breaker_group_id}}
    JOIN {{space_group_owners.table}} so ON so.{{space_group_owners.space_group_id}} = bg.{{breaker_groups.space_group_id}}
    WHERE b.{{breakers.id}} = p_breaker_id AND so.{{space_group_owners.user_id}} = p_user_id
    FOR UPDATE OF bg;

    IF v_breaker_group_id IS NULL THEN
        RETURN;
    END IF;

    IF p_display_order IS NOT NULL AND p_display_order <> v_current_order THEN
        v_delta := p_display_order - v_current_order;

        IF ABS(v_delta) = 1 THEN
            UPDATE {{breakers.table}}
            SET {{breakers.display_order}} = CASE WHEN {{breakers.id}} = p_breaker_id THEN p_display_order ELSE v_current_order END
            WHERE {{breakers.breaker_group_id}} = v_breaker_group_id AND ({{breakers.id}} = p_breaker_id OR {{breakers.display_order}} = p_display_order);
        ELSIF v_delta > 0 THEN
            UPDATE {{breakers.table}}
            SET {{breakers.display_order}} = CASE WHEN {{breakers.id}} = p_breaker_id THEN p_display_order ELSE {{breakers.display_order}} - 1 END
            WHERE {{breakers.breaker_group_id}} = v_breaker_group_id AND {{breakers.display_order}} >= v_current_order AND {{breakers.display_order}} <= p_display_order;
        ELSE
            UPDATE {{breakers.table}}
            SET {{breakers.display_order}} = CASE WHEN {{breakers.id}} = p_breaker_id THEN p_display_order ELSE {{breakers.display_order}} + 1 END
            WHERE {{breakers.breaker_group_id}} = v_breaker_group_id AND {{breakers.display_order}} >= p_display_order AND {{breakers.display_order}} <= v_current_order;
        END IF;
    END IF;

    IF p_name IS NOT NULL THEN
        UPDATE {{breakers.table}} SET {{breakers.name}} = p_name WHERE {{breakers.id}} = p_breaker_id;
    END IF;

    RETURN QUERY
    SELECT b.{{breakers.id}}, b.{{breakers.name}}, b.{{breakers.breaker_group_id}}, b.{{breakers.space_group_id}}, b.{{breakers.display_order}}, b.{{breakers.upstream_breaker_id}}
    FROM {{breakers.table}} b
    WHERE b.{{breakers.id}} = p_breaker_id;
END;
$$;

DROP FUNCTION IF EXISTS {{fn_set_breaker_upstream.func}}(INT, INT, INT);

CREATE OR REPLACE FUNCTION {{fn_set_breaker_upstream.func}}(
    p_user_id INT,
    p_breaker_id INT,
    p_upstream_breaker_id INT DEFAULT NULL
) RETURNS TABLE (
    {{fn_set_breaker_upstream.o_id}} INT,
    {{fn_set_breaker_upstream.o_name}} VARCHAR(255),
    {{fn_set_breaker_upstream.o_breaker_group_id}} INT,
    {{fn_set_breaker_upstream.o_space_group_id}} INT,
    {{fn_set_breaker_upstream.o_display_order}} INT,
    {{fn_set_breaker_upstream.o_upstream_breaker_id}} INT
) LANGUAGE plpgsql AS $$
DECLARE
    v_space_group_id INT;
    v_has_cycle BOOLEAN;
BEGIN
    SELECT b.{{breakers.space_group_id}}
    INTO v_space_group_id
    FROM {{breakers.table}} b
    JOIN {{breaker_groups.table}} bg ON bg.{{breaker_groups.id}} = b.{{breakers.breaker_group_id}}
    JOIN {{space_groups.table}} sg ON sg.{{space_groups.id}} = bg.{{breaker_groups.space_group_id}}
    JOIN {{space_group_owners.table}} so ON so.{{space_group_owners.space_group_id}} = sg.{{space_groups.id}}
    WHERE b.{{breakers.id}} = p_breaker_id AND so.{{space_group_owners.user_id}} = p_user_id
    FOR UPDATE OF sg;

    IF v_space_group_id IS NULL THEN
        RETURN;
    END IF;

    IF p_upstream_breaker_id IS NOT NULL THEN
        IF p_upstream_breaker_id = p_breaker_id THEN
            RAISE EXCEPTION 'breaker % cannot reference itself as upstream', p_breaker_id
                USING ERRCODE = '{{fn_set_breaker_upstream.err_self_reference}}';
        END IF;

        IF NOT EXISTS (
            SELECT 1 FROM {{breakers.table}}
            WHERE {{breakers.id}} = p_upstream_breaker_id AND {{breakers.space_group_id}} = v_space_group_id
        ) THEN
            RETURN;
        END IF;

        WITH RECURSIVE upstream_chain AS (
            SELECT b.{{breakers.id}}, b.{{breakers.upstream_breaker_id}}
            FROM {{breakers.table}} b
            WHERE b.{{breakers.id}} = p_upstream_breaker_id

            UNION ALL

            SELECT p.{{breakers.id}}, p.{{breakers.upstream_breaker_id}}
            FROM {{breakers.table}} p
            JOIN upstream_chain c ON p.{{breakers.id}} = c.{{breakers.upstream_breaker_id}}
        )
        SELECT EXISTS (SELECT 1 FROM upstream_chain WHERE {{breakers.id}} = p_breaker_id) INTO v_has_cycle;

        IF v_has_cycle THEN
            RAISE EXCEPTION 'upstream breaker % would create a cycle for breaker %', p_upstream_breaker_id, p_breaker_id
                USING ERRCODE = '{{fn_set_breaker_upstream.err_cycle_detected}}';
        END IF;
    END IF;

    UPDATE {{breakers.table}} SET {{breakers.upstream_breaker_id}} = p_upstream_breaker_id WHERE {{breakers.id}} = p_breaker_id;

    RETURN QUERY
    SELECT b.{{breakers.id}}, b.{{breakers.name}}, b.{{breakers.breaker_group_id}}, b.{{breakers.space_group_id}}, b.{{breakers.display_order}}, b.{{breakers.upstream_breaker_id}}
    FROM {{breakers.table}} b
    WHERE b.{{breakers.id}} = p_breaker_id;
END;
$$;

CREATE OR REPLACE FUNCTION {{fn_add_device_breaker.func}}(
    p_user_id INT,
    p_device_id INT,
    p_breaker_id INT
) RETURNS TABLE (
    {{fn_add_device_breaker.o_result}} INT
) LANGUAGE plpgsql AS $$
DECLARE
    v_space_group_id INT;
BEGIN
    SELECT s.{{spaces.space_group_id}}
    INTO v_space_group_id
    FROM {{devices.table}} d
    JOIN {{spaces.table}} s ON s.{{spaces.id}} = d.{{devices.space_id}}
    JOIN {{space_group_owners.table}} so ON so.{{space_group_owners.space_group_id}} = s.{{spaces.space_group_id}}
    WHERE d.{{devices.id}} = p_device_id AND so.{{space_group_owners.user_id}} = p_user_id;

    IF v_space_group_id IS NULL THEN
        RETURN QUERY SELECT 1;
        RETURN;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM {{breakers.table}}
        WHERE {{breakers.id}} = p_breaker_id AND {{breakers.space_group_id}} = v_space_group_id
    ) THEN
        RETURN QUERY SELECT 2;
        RETURN;
    END IF;

    IF EXISTS (
        SELECT 1 FROM {{device_breakers.table}}
        WHERE {{device_breakers.device_id}} = p_device_id AND {{device_breakers.breaker_id}} = p_breaker_id
    ) THEN
        RETURN QUERY SELECT 3;
        RETURN;
    END IF;

    INSERT INTO {{device_breakers.table}} ({{device_breakers.device_id}}, {{device_breakers.breaker_id}})
    VALUES (p_device_id, p_breaker_id);

    RETURN QUERY SELECT 0;
END;
$$;
