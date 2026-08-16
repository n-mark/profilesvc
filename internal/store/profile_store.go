package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"profile-service/internal/models"
)

var ErrProfileNotFound = errors.New("profile not found")

type ProfileStore struct {
	db *pgxpool.Pool
}

func NewProfileStore(db *pgxpool.Pool) *ProfileStore {
	return &ProfileStore{db: db}
}

func (s *ProfileStore) Create(ctx context.Context, profile models.Profile) (models.Profile, error) {
	row := s.db.QueryRow(ctx,
		`INSERT INTO profile (userid, name, surname, date_of_birth, gender, city, bio, interests)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING profile_id`,
		profile.OwnerID, profile.Name, profile.Surname, profile.DateOfBirth,
		string(profile.Gender), profile.City, profile.Bio, profile.Interests,
	)

	if err := row.Scan(&profile.ID); err != nil {
		return models.Profile{}, err
	}

	return profile, nil
}

func (s *ProfileStore) Update(ctx context.Context, profile models.Profile) (models.Profile, error) {
	tag, err := s.db.Exec(ctx,
		`UPDATE profile
		 SET name=$1, surname=$2, date_of_birth=$3, gender=$4, city=$5, bio=$6, interests=$7
		 WHERE profile_id=$8 AND userid=$9`,
		profile.Name, profile.Surname, profile.DateOfBirth,
		string(profile.Gender), profile.City, profile.Bio, profile.Interests,
		profile.ID, profile.OwnerID,
	)
	if err != nil {
		return models.Profile{}, err
	}
	if tag.RowsAffected() == 0 {
		return models.Profile{}, ErrProfileNotFound
	}

	return profile, nil
}

func (s *ProfileStore) UpdateStatus(ctx context.Context, ownerId int64, status string) (bool, error) {
	tag, err := s.db.Exec(ctx,
		`UPDATE profile
		 SET status=$1
		 WHERE userid=$2`,
		status, ownerId)

	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		return false, ErrProfileNotFound
	}

	return true, nil
}

func (s *ProfileStore) GetByOwnerID(ctx context.Context, ownerID int64) (models.Profile, error) {
	var p models.Profile
	var gender string
	err := s.db.QueryRow(ctx,
		`SELECT profile_id, userid, name, surname, date_of_birth, gender, city, bio, interests
		 FROM profile WHERE userid=$1`,
		ownerID,
	).Scan(&p.ID, &p.OwnerID, &p.Name, &p.Surname, &p.DateOfBirth, &gender, &p.City, &p.Bio, &p.Interests)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Profile{}, ErrProfileNotFound
		}
		return models.Profile{}, err
	}

	p.Gender = models.Gender(gender)
	return p, nil
}

// Upsert creates a profile if absent or updates editable fields for the existing owner.
func (s *ProfileStore) Upsert(ctx context.Context, profile models.Profile) (models.Profile, error) {
	row := s.db.QueryRow(ctx,
		`INSERT INTO profile (userid, name, surname, date_of_birth, gender, city, bio, interests)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (userid) DO UPDATE SET
		   name = EXCLUDED.name,
		   surname = EXCLUDED.surname,
		   date_of_birth = EXCLUDED.date_of_birth,
		   gender = EXCLUDED.gender,
		   city = EXCLUDED.city,
		   bio = EXCLUDED.bio,
		   interests = EXCLUDED.interests
		 RETURNING profile_id`,
		profile.OwnerID, profile.Name, profile.Surname, profile.DateOfBirth,
		string(profile.Gender), profile.City, profile.Bio, profile.Interests,
	)

	if err := row.Scan(&profile.ID); err != nil {
		return models.Profile{}, err
	}
	return profile, nil
}

func (s *ProfileStore) GetByID(ctx context.Context, id int64) (models.Profile, error) {
	var p models.Profile
	var gender string
	err := s.db.QueryRow(ctx,
		`SELECT profile_id, userid, name, surname, date_of_birth, gender, city, bio, interests
		 FROM profile WHERE profile_id=$1`,
		id,
	).Scan(&p.ID, &p.OwnerID, &p.Name, &p.Surname, &p.DateOfBirth, &gender, &p.City, &p.Bio, &p.Interests)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Profile{}, ErrProfileNotFound
		}
		return models.Profile{}, err
	}

	p.Gender = models.Gender(gender)
	return p, nil
}

func (s *ProfileStore) List(ctx context.Context, query models.QueryDTO) ([]models.Profile, error) {
	sql := `SELECT profile_id, userid, name, surname, date_of_birth, gender, city, bio, interests
	        FROM profile WHERE 1=1`
	args := []any{}
	argIdx := 1

	if query.Gender != "" {
		sql += fmt.Sprintf(" AND gender=$%d", argIdx)
		args = append(args, string(query.Gender))
		argIdx++
	}
	if query.City != "" {
		sql += fmt.Sprintf(" AND city=$%d", argIdx)
		args = append(args, query.City)
		argIdx++
	}
	if query.Query != "" {
		sql += fmt.Sprintf(" AND (name ILIKE $%d OR surname ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+query.Query+"%")
		argIdx++
	}
	if query.AgeFrom > 0 {
		sql += fmt.Sprintf(" AND date_of_birth <= (now() - interval '%d years')", query.AgeFrom)
	}
	if query.AgeTo > 0 {
		sql += fmt.Sprintf(" AND date_of_birth >= (now() - interval '%d years')", query.AgeTo)
	}
	if query.Count > 0 {
		sql += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, query.Count)
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []models.Profile
	for rows.Next() {
		var p models.Profile
		var gender string
		if err := rows.Scan(&p.ID, &p.OwnerID, &p.Name, &p.Surname, &p.DateOfBirth, &gender, &p.City, &p.Bio, &p.Interests); err != nil {
			return nil, err
		}
		p.Gender = models.Gender(gender)
		profiles = append(profiles, p)
	}

	return profiles, rows.Err()
}
