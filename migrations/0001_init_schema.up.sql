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
CREATE OR REPLACE FUNCTION allocate_classrooms(
    _semester TEXT,
    _faculty TEXT,
    _program TEXT,
    _count INT
)
RETURNS VOID AS $$
DECLARE
    available_room RECORD;
    allocated_count INT := 0;
BEGIN
    -- Loop through available classroom-type rooms and allocate them
    FOR available_room IN
        SELECT r.id
        FROM rooms r
        WHERE r.type = 'classroom'
          AND r.id NOT IN (
              SELECT room_id
              FROM room_allocations
              WHERE semester = _semester
          )
        ORDER BY r.id
        LIMIT _count
    LOOP
        -- Insert allocation
        INSERT INTO room_allocations (
            state, room_id, semester, faculty, program
        ) VALUES (
            'awaiting', available_room.id, _semester, _faculty, _program
        );

        allocated_count := allocated_count + 1;
    END LOOP;

    -- If not enough rooms were allocated, raise exception
    IF allocated_count < _count THEN
        RAISE EXCEPTION 'Not enough available classroom rooms to allocate (% out of %)', allocated_count, _count;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function to allocate laboratories
CREATE OR REPLACE FUNCTION allocate_laboratories(
    _semester TEXT,
    _faculty TEXT,
    _program TEXT,
    _count INT
)
RETURNS VOID AS $$
DECLARE
    allocated_count INT := 0;
    lab_room RECORD;
    classroom_room RECORD;
BEGIN
    -- Step 1: Try to allocate as many laboratory-type rooms as available
    FOR lab_room IN
        SELECT r.id
        FROM rooms r
        WHERE r.type = 'laboratory'
          AND r.id NOT IN (
              SELECT room_id FROM room_allocations WHERE semester = _semester
          )
        ORDER BY r.id
        LIMIT _count
    LOOP
        INSERT INTO room_allocations (
            state, room_id, semester, faculty, program, adapted
        ) VALUES (
            'awaiting', lab_room.id, _semester, _faculty, _program, FALSE
        );
        allocated_count := allocated_count + 1;
    END LOOP;

    -- Step 2: If not enough, allocate classroom-type rooms as adapted labs
    IF allocated_count < _count THEN
        FOR classroom_room IN
            SELECT r.id
            FROM rooms r
            WHERE r.type = 'classroom'
              AND r.id NOT IN (
                  SELECT room_id FROM room_allocations WHERE semester = _semester
              )
            ORDER BY r.id
            LIMIT (_count - allocated_count)
        LOOP
            INSERT INTO room_allocations (
                state, room_id, semester, faculty, program, adapted
            ) VALUES (
                'awaiting', classroom_room.id, _semester, _faculty, _program, TRUE
            );
            allocated_count := allocated_count + 1;
        END LOOP;
    END IF;

    -- Step 3: If still not enough rooms, raise exception
    IF allocated_count < _count THEN
        RAISE EXCEPTION 'Not enough rooms to allocate for labs (% out of %)', allocated_count, _count;
    END IF;
END;
$$ LANGUAGE plpgsql;
