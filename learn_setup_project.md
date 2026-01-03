# 1. Create the main project folder
mkdir CabBookingService

# 2. Go into the new folder
cd CabBookingService

# 3. Create the command folder for your 'main' package (-p is creating parents also)
mkdir -p cmd/api

# 4. Create the 'internal' folders for all your logic
mkdir -p internal/db/migrations
mkdir -p internal/config
mkdir -p internal/models
mkdir -p internal/repositories
mkdir -p internal/services
mkdir -p internal/controllers



Setup Git
# 1. Initialize a new Git repository
git init

# 2. Add all your new files to be tracked
git add .

# 3. Make your very first commit
git commit -m "Initial commit: project structure, Docker, and Task setup"

# --------------
brew install go-task
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest



- task db-up
- task migrate-up
- task run
- task migrate-new -- create_bookings_table
