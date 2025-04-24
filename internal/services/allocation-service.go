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
		Programs: make([]models.ProgramAllocation, len(request.Programs)),
	}

	for i, program := range request.Programs {
		// 4.1 Create a new program allocation
		response.Programs[i] = models.ProgramAllocation{
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
		for i, room := range rooms {
			if room.Type == repository.RoomTypeClassroom {
				if room.Adapted {
					response.Programs[i].Adapted = append(response.Programs[i].Adapted, room.Name)
				} else {
					response.Programs[i].Classrooms = append(response.Programs[i].Classrooms, room.Name)
				}
			} else {
				response.Programs[i].Laboratories = append(response.Programs[i].Laboratories, room.Name)
			}
		}
	}

	// 5. Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return response, nil
}
