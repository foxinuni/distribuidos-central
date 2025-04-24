-- Drop triggers first
DROP TRIGGER IF EXISTS enforce_adapted_logic ON room_allocations;
DROP FUNCTION IF EXISTS check_adapted_consistency;

-- Drop stored procedures
DROP FUNCTION IF EXISTS allocate_laboratories(TEXT, TEXT, TEXT, INT);
DROP FUNCTION IF EXISTS allocate_classrooms(TEXT, TEXT, TEXT, INT);
DROP FUNCTION IF EXISTS lock_rooms(TEXT, TEXT);
DROP FUNCTION IF EXISTS generate_rooms(INT, INT);

-- Drop tables
DROP TABLE IF EXISTS room_allocations;
DROP TABLE IF EXISTS rooms;

-- Drop ENUM types
DROP TYPE IF EXISTS room_state;
DROP TYPE IF EXISTS room_type;
