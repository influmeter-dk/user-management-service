package timer_event

import (
	"time"

	"github.com/coneno/logger"
)

// CleanUpUnverifiedUsers handles the deletion of unverified accounts after a threshold delay
func (s *UserManagementTimerService) CleanUpUnverifiedUsers() {
	logger.Debug.Println("Starting clean up job for unverified users:")
	instances, err := s.globalDBService.GetAllInstances()
	if err != nil {
		logger.Error.Printf("unexpected error: %s", err.Error())
	}
	deleteUnverifiedUsersAfter := s.CleanUpTimeThreshold
	for _, instance := range instances {
		count, err := s.userDBService.DeleteUnverfiedUsers(instance.InstanceID, time.Now().Unix()-deleteUnverifiedUsersAfter)
		if err != nil {
			logger.Error.Printf("unexpected error: %s", err.Error())
			continue
		}
		if count > 0 {
			logger.Info.Printf("%s: removed %d unverified accounts", instance.InstanceID, count)
		} else {
			logger.Debug.Printf("%s: removed %d unverified accounts", instance.InstanceID, count)
		}

	}
}
