package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

// Habit represents the habit data
type Habit struct {
	ID     uint   `gorm:"primaryKey"`
	Name   string `gorm:"not null"`
	Streak int    `gorm:"default:0"`
}

// calculateProgress calculates the streak progress as a percentage
func calculateProgress(streak int) int {
	goal := 10 // Set a streak goal of 10
	if streak > goal {
		return 100
	}
	return (streak * 100) / goal
}

// InitDB initializes the SQLite database
func InitDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("db/habits.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the Habit schema
	err = db.AutoMigrate(&Habit{})
	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}

	log.Println("Database connection initialized successfully")
}

func main() {
	// Initialize the database
	InitDB()

	// Set up Gin router
	r := gin.Default()

	// Register custom template functions
	r.SetFuncMap(template.FuncMap{
		"calculateProgress": calculateProgress,
	})

	// Load templates and static files
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	// Routes
	r.GET("/", showHabits)
	r.POST("/add", addHabit)
	r.POST("/delete_all", deleteAllHabits)
	r.GET("/mark_done/:id", markHabitDone)

	// Start the server
	r.Run(":8080")
}

// showHabits renders the main page with all habits
func showHabits(c *gin.Context) {
	var habits []Habit
	db.Find(&habits)
	c.HTML(http.StatusOK, "index.html", gin.H{"habits": habits})
}

// addHabit adds a new habit to the database
func addHabit(c *gin.Context) {
	name := c.PostForm("name")
	if name != "" {
		db.Create(&Habit{Name: name})
	}
	c.Redirect(http.StatusFound, "/")
}

// markHabitDone increments the streak of a habit
func markHabitDone(c *gin.Context) {
	id := c.Param("id")
	var habit Habit

	// Check if the habit exists
	if err := db.First(&habit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
		return
	}

	// Increment the streak
	habit.Streak++
	db.Save(&habit)

	// Redirect back to the main page
	c.Redirect(http.StatusFound, "/")
}

// deleteAllHabits deletes all habits from the database
func deleteAllHabits(c *gin.Context) {
	// Delete all habits from the database
	if err := db.Exec("DELETE FROM habits").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete habits"})
		return
	}

	// Reset the auto-increment counter for SQLite
	db.Exec("DELETE FROM sqlite_sequence WHERE name='habits'")

	c.Redirect(http.StatusFound, "/")
}
