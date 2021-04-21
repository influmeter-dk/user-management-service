package timer_event

import (
	"time"

	"github.com/influenzanet/user-management-service/pkg/dbs/globaldb"
	"github.com/influenzanet/user-management-service/pkg/dbs/userdb"
	"github.com/influenzanet/user-management-service/pkg/models"
)

// UserManagementTimerService handles background times for user management (cleanup for example).
type UserManagementTimerService struct {
	globalDBService      *globaldb.GlobalDBService
	userDBService        *userdb.UserDBService
	clients              *models.APIClients
	TimerEventFrequency  int64 // how often the timer event should be performed (only from one instance of the service) - seconds
	CleanUpTimeThreshold int64 // if user account not verified, remove user after this many seconds
}

func NewUserManagmentTimerService(
	frequency int64,
	globalDBService *globaldb.GlobalDBService,
	userDBService *userdb.UserDBService,
	clients *models.APIClients,
	cleanUpTimeThreshold int64,
) *UserManagementTimerService {
	return &UserManagementTimerService{
		globalDBService:      globalDBService,
		userDBService:        userDBService,
		TimerEventFrequency:  frequency,
		clients:              clients,
		CleanUpTimeThreshold: cleanUpTimeThreshold,
	}
}

func (s *UserManagementTimerService) Run() {
	go s.startTimerThread(s.TimerEventFrequency)
}

func (s *UserManagementTimerService) startTimerThread(timeCheckInterval int64) {
	// TODO: turn off gracefully
	for {
		<-time.After(time.Duration(timeCheckInterval) * time.Second)
		go s.CleanUpUnverifiedUsers()
	}
}
