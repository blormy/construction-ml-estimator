package repository

import (
	"constructmaterialai/internal/db"
	"constructmaterialai/internal/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// сохранение проекта и связанных единиц оборудования через ассоциацию
func CreateProject(project *models.Project) error {
	return db.DB.Create(project).Error
}

// обновление проекта и оборудования с проверкой версии (optimistic locking)
func UpdateProject(project *models.Project, oldUpdatedAt time.Time) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		// проверка версии
		var currentProject models.Project
		if err := tx.Where("id = ?", project.ID).First(&currentProject).Error; err != nil {
			return err
		}

		if currentProject.UpdatedAt.After(oldUpdatedAt) && !currentProject.UpdatedAt.Equal(oldUpdatedAt) {
			return fmt.Errorf("CONFLICT_VERSION") // маркер конфликта версий
		}

		// удаление старого оборудования
		if err := tx.Where("project_id = ?", project.ID).Delete(&models.EquipmentItem{}).Error; err != nil {
			return err
		}

		// обновление проекта
		if err := tx.Save(project).Error; err != nil {
			return err
		}

		return nil
	})
}

// получение личных и командных проектов пользователя
func GetProjectsByUser(userID uint) ([]models.Project, error) {
	teamIDs, err := GetUserTeamIDs(userID)
	if err != nil {
		return nil, err
	}

	var projects []models.Project
	query := db.DB.Preload("EquipmentItems").Order("created_at DESC")

	if len(teamIDs) > 0 {
		query = query.Where(
			"(user_id = ?) OR (team_id IN (?))",
			userID, teamIDs,
		)
	} else {
		query = query.Where("user_id = ?", userID)
	}

	err = query.Find(&projects).Error
	return projects, err
}

// получение только личных проектов пользователя
func GetPersonalProjects(userID uint) ([]models.Project, error) {
	var projects []models.Project
	err := db.DB.
		Where("user_id = ? AND (team_id IS NULL OR team_id = 0)", userID).
		Order("created_at DESC").
		Find(&projects).Error
	return projects, err
}

// получение проектов конкретной команды
func GetProjectsByTeam(teamID uint) ([]models.Project, error) {
	var projects []models.Project
	err := db.DB.Preload("EquipmentItems").
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		Find(&projects).Error
	return projects, err
}

// получение проекта по id с предзагрузкой оборудования и проверкой доступа
func GetProjectByID(projectID, userID uint) (*models.Project, error) {
	teamIDs, err := GetUserTeamIDs(userID)
	if err != nil {
		return nil, err
	}

	var project models.Project
	query := db.DB.Preload("EquipmentItems")

	if len(teamIDs) > 0 {
		query = query.Where(
			"id = ? AND ((user_id = ?) OR (team_id IN (?)))",
			projectID, userID, teamIDs,
		)
	} else {
		query = query.Where("id = ? AND user_id = ?", projectID, userID)
	}

	if err := query.First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// перенос проекта в команду или обратно в личные
func MoveProjectToTeam(projectID, ownerID uint, teamID *uint) error {
	result := db.DB.Model(&models.Project{}).
		Where("id = ? AND user_id = ?", projectID, ownerID).
		Update("team_id", teamID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("проект не найден или вы не являетесь владельцем")
	}
	return nil
}

// удаление проекта владельцем
func DeleteProject(projectID, userID uint) error {
	result := db.DB.
		Where("id = ? AND user_id = ?", projectID, userID).
		Delete(&models.Project{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("проект не найден или вы не являетесь владельцем")
	}
	return nil
}
