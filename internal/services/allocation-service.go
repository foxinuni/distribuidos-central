package services

import (
	"context"

	"github.com/foxinuni/distribuidos-central/internal/models"
	"github.com/foxinuni/distribuidos-central/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AllocationService interface {
	Allocate(ctx context.Context, request *models.AllocateRequest) (*models.AllocateResponse, error)
}

type SqlcAllocationService struct {
	pool *pgxpool.Pool
}

func NewSqlcAllocationService(pool *pgxpool.Pool) *SqlcAllocationService {
	return &SqlcAllocationService{
		pool: pool,
	}
}

func (s *SqlcAllocationService) Allocate(ctx context.Context, request *models.AllocateRequest) (*models.AllocateResponse, error) {
	// 1. Create a transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Create new querier for transaction
	querier := repository.New(tx)

	// 3. Allocate rooms
	for _, program := range request.Programs {
		if err := querier.AllocateClassrooms(ctx, repository.AllocateClassroomsParams{
			Semester: request.Semester,
			Faculty:  request.Faculty,
			Program:  program.Name,
			Count:    int32(program.Classrooms),
		}); err != nil {
			return nil, err
		}

		if err := querier.AllocateLaboratories(ctx, repository.AllocateLaboratoriesParams{
			Semester: request.Semester,
			Faculty:  request.Faculty,
			Program:  program.Name,
			Count:    int32(program.Laboratories),
		}); err != nil {
			return nil, err
		}
	}

	// 4. Get the allocated rooms
	response := &models.AllocateResponse{
		Semester: request.Semester,
		Faculty:  request.Faculty,
		Programs: []models.ProgramAllocation{},
	}

	for _, program := range request.Programs {
		// 4.1 Create a new program allocation
		program := models.ProgramAllocation{
			Name:         program.Name,
			Classrooms:   []string{},
			Laboratories: []string{},
			Adapted:      []string{},
		}

		// 4.2 Get the allocated rooms
		rooms, err := querier.GetRoomsByFacultyProgramSemester(ctx, repository.GetRoomsByFacultyProgramSemesterParams{
			Semester: request.Semester,
			Faculty:  request.Faculty,
			Program:  program.Name,
		})
		if err != nil {
			return nil, err
		}

		// 4.3 Add the allocated rooms to the response
		for _, room := range rooms {
			if room.Type == repository.RoomTypeClassroom {
				if room.Adapted {
					program.Adapted = append(program.Adapted, room.Name)
				} else {
					program.Classrooms = append(program.Classrooms, room.Name)
				}
			} else {
				program.Laboratories = append(program.Laboratories, room.Name)
			}
		}

		// 4.4 Add the program allocation to the response
		response.Programs = append(response.Programs, program)
	}

	// 5. Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return response, nil
}
