-- name: GenerateRooms :exec
CALL generate_rooms($1, $2);

-- name: LockRooms :exec
CALL lock_rooms($1, $2);

-- name: AllocateClassrooms :exec
CALL allocate_classrooms($1, $2, $3, $4);

-- name: AllocateLaboratories :exec
CALL allocate_laboratories($1, $2, $3, $4);

-- name: GetRoomsByFacultyProgramSemester :many
SELECT r.id, r.name, r.type, ra.adapted
FROM rooms r
JOIN room_allocations ra
    ON r.id = ra.room_id
WHERE ra.faculty = $1
    AND ra.program = $2
    AND ra.semester = $3;
