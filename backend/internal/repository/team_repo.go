package repository

import (
	"constructmaterialai/internal/db"
	"constructmaterialai/internal/models"
	"fmt"

	"gorm.io/gorm"
)

// участник команды с email (результат join)
type TeamMemberInfo struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// создание новой команды с добавлением владельца как участника с ролью owner
func CreateTeam(team *models.Team) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(team).Error; err != nil {
			return err
		}
		// владелец автоматически становится участником команды
		member := models.TeamMember{
			TeamID: team.ID,
			UserID: team.OwnerID,
			Role:   "owner",
		}
		return tx.Create(&member).Error
	})
}

// получение команд, где пользователь является владельцем или участником
func GetTeamsByUser(userID uint) ([]models.Team, error) {
	var teams []models.Team
	err := db.DB.
		Where("id IN (SELECT team_id FROM team_members WHERE user_id = ?)", userID).
		Order("created_at DESC").
		Find(&teams).Error
	return teams, err
}

// получение команды по id
func GetTeamByID(teamID uint) (*models.Team, error) {
	var team models.Team
	if err := db.DB.First(&team, teamID).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

// проверка, является ли пользователь участником команды
func IsTeamMember(teamID, userID uint) bool {
	var count int64
	db.DB.Model(&models.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Count(&count)
	return count > 0
}

// проверка, является ли пользователь владельцем команды
func IsTeamOwner(teamID, userID uint) bool {
	var count int64
	db.DB.Model(&models.Team{}).
		Where("id = ? AND owner_id = ?", teamID, userID).
		Count(&count)
	return count > 0
}

// добавление пользователя в команду по id
func AddTeamMember(teamID, userID uint, role string) error {
	if role == "" {
		role = "member"
	}
	member := models.TeamMember{
		TeamID: teamID,
		UserID: userID,
		Role:   role,
	}
	return db.DB.Create(&member).Error
}

// поиск пользователя по email
func FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// получение списка участников команды с email
func GetTeamMembers(teamID uint) ([]TeamMemberInfo, error) {
	var members []TeamMemberInfo
	err := db.DB.
		Table("team_members").
		Select("team_members.user_id, users.email, team_members.role").
		Joins("JOIN users ON users.id = team_members.user_id").
		Where("team_members.team_id = ?", teamID).
		Scan(&members).Error
	return members, err
}

// удаление участника из команды
func RemoveTeamMember(teamID, userID uint) error {
	return db.DB.
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&models.TeamMember{}).Error
}

// удаление команды и перевод связанных проектов в личные проекты владельца
func DeleteTeam(teamID, ownerID uint) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		// проверка, что инициатор является владельцем
		var team models.Team
		if err := tx.Where("id = ? AND owner_id = ?", teamID, ownerID).First(&team).Error; err != nil {
			return fmt.Errorf("команда не найдена или вы не являетесь владельцем")
		}

		// перенос проектов команды в личные проекты владельца
		if err := tx.Model(&models.Project{}).
			Where("team_id = ?", teamID).
			Updates(map[string]interface{}{"team_id": nil}).Error; err != nil {
			return err
		}

		// удаление всех участников
		if err := tx.Where("team_id = ?", teamID).Delete(&models.TeamMember{}).Error; err != nil {
			return err
		}

		// удаление команды
		return tx.Delete(&team).Error
	})
}

// получение id всех команд пользователя
func GetUserTeamIDs(userID uint) ([]uint, error) {
	var teamIDs []uint
	err := db.DB.Model(&models.TeamMember{}).
		Where("user_id = ?", userID).
		Pluck("team_id", &teamIDs).Error
	return teamIDs, err
}
