-- Define ENUM types
CREATE TYPE room_type AS ENUM ('classroom', 'laboratory');
CREATE TYPE room_state AS ENUM ('awaiting', 'locked');

-- Create rooms table
CREATE TABLE IF NOT EXISTS rooms (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type room_type NOT NULL
);

-- Create room_allocations table
CREATE TABLE IF NOT EXISTS room_allocations (
    id SERIAL PRIMARY KEY,
    state room_state NOT NULL,
    room_id INTEGER NOT NULL REFERENCES rooms(id),
    semester TEXT NOT NULL,
    faculty TEXT NOT NULL,
    program TEXT NOT NULL,
    adapted BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (semester, room_id)
);

-- Trigger function to enforce adapted logic
CREATE OR REPLACE FUNCTION check_adapted_consistency() RETURNS TRIGGER AS $$
DECLARE
    room_type_val room_type;
BEGIN
    SELECT type INTO room_type_val FROM rooms WHERE id = NEW.room_id;

    IF room_type_val = 'laboratory' AND NEW.adapted THEN
        RAISE EXCEPTION 'Laboratory rooms cannot be adapted.';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create the trigger
CREATE TRIGGER enforce_adapted_logic
BEFORE INSERT OR UPDATE ON room_allocations
FOR EACH ROW
EXECUTE FUNCTION check_adapted_consistency();

-- Function to generate rooms
CREATE OR REPLACE FUNCTION generate_rooms(normal_rooms INT, laboratories INT) RETURNS VOID AS $$
DECLARE
    name TEXT;
BEGIN
    TRUNCATE rooms CASCADE;

    FOR i IN 1..normal_rooms LOOP
        SELECT CONCAT('AUL-', LPAD(i::TEXT, 3, '0')) INTO name;
        INSERT INTO rooms (name, type) VALUES (name, 'classroom');
    END LOOP;

    FOR i IN 1..laboratories LOOP
        SELECT CONCAT('LAB-', LPAD(i::TEXT, 3, '0')) INTO name;
        INSERT INTO rooms (name, type) VALUES (name, 'laboratory');
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to lock rooms
CREATE OR REPLACE FUNCTION lock_rooms(faculty_name TEXT, semester_name TEXT) RETURNS VOID AS $$
BEGIN
    UPDATE room_allocations
    SET state = 'locked'
    WHERE faculty = faculty_name AND semester = semester_name;
END;
$$ LANGUAGE plpgsql;

-- Function to allocate classrooms
CREATE OR REPLACE FUNCTION allocate_classrooms(semester TEXT, faculty TEXT, program TEXT, count INT) RETURNS VOID AS $$
DECLARE
    room RECORD;
BEGIN
    FOR i IN 1..count LOOP
        SELECT r INTO room
        FROM rooms r
        WHERE r.type = 'classroom'
        AND NOT EXISTS (
            SELECT 1
            FROM room_allocations ra
            WHERE ra.room_id = r.id
            AND ra.semester = semester
            AND ra.faculty = faculty
        )
        ORDER BY r.name ASC
        LIMIT 1;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'No classrooms available for allocation';
        END IF;

        INSERT INTO room_allocations (state, room_id, semester, faculty, program)
        VALUES ('awaiting', room.id, semester, faculty, program);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to allocate laboratories
CREATE OR REPLACE FUNCTION allocate_laboratories(semester TEXT, faculty TEXT, program TEXT, count INT) RETURNS VOID AS $$
DECLARE
    room RECORD;
BEGIN
    FOR i IN 1..count LOOP
        SELECT r INTO room
        FROM rooms r
        WHERE r.type IN ('classroom', 'laboratory')
        AND NOT EXISTS (
            SELECT 1
            FROM room_allocations ra
            WHERE ra.room_id = r.id
            AND ra.semester = semester
            AND ra.faculty = faculty
        )
        ORDER BY r.type DESC, r.name ASC
        LIMIT 1;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'No classrooms or laboratories available for allocation';
        END IF;

        INSERT INTO room_allocations (state, room_id, semester, faculty, program, adapted)
        VALUES (
            'awaiting',
            room.id,
            semester,
            faculty,
            program,
            (room.type = 'classroom')::BOOLEAN
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;
