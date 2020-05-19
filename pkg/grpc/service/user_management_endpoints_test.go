package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/models"
	messageMock "github.com/influenzanet/user-management-service/test/mocks/messaging_service"
)

func TestCreateUserEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockMessagingClient := messageMock.NewMockMessagingServiceApiClient(mockCtrl)

	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenMinimumAgeMin:  time.Second * 1,
			TokenExpiryInterval: time.Second * 2,
		},
		clients: &models.APIClients{
			MessagingService: mockMessagingClient,
		},
	}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.CreateUser(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.CreateUserReq{}
		_, err := s.CreateUser(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non admin user", func(t *testing.T) {
		req := &api.CreateUserReq{
			Token: &api.TokenInfos{
				Id:         "testuserid",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles": "PARTICIPANT",
				},
			},
			AccountId:       "test_created_user@email.test",
			InitialPassword: "initpw",
		}
		_, err := s.CreateUser(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "permission denied")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with valid arguments", func(t *testing.T) {
		mockMessagingClient.EXPECT().SendInstantEmail(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, nil)

		req := &api.CreateUserReq{
			Token: &api.TokenInfos{
				Id:         "testuserid",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles": "PARTICIPANT,ADMIN",
				},
			},
			AccountId:         "test_created_user@email.test",
			InitialPassword:   "initpw543",
			PreferredLanguage: "en",
			Roles:             []string{"PARTICIPANT", "ADMIN"},
		}
		resp, err := s.CreateUser(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.Account.AccountId != req.AccountId {
			t.Errorf("unexpected response: %s", resp)
			return
		}
	})

	t.Run("with already existing user", func(t *testing.T) {
		req := &api.CreateUserReq{
			Token: &api.TokenInfos{
				Id:         "testuserid",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles": "PARTICIPANT,ADMIN",
				},
			},
			AccountId:       "test_created_user@email.test",
			InitialPassword: "initpwi7867-k",
		}
		_, err := s.CreateUser(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "user already exists")
		if !ok {
			t.Error(msg)
		}
	})

}

func TestAddRoleForUserEndpoint(t *testing.T) {
	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenMinimumAgeMin:  time.Second * 1,
			TokenExpiryInterval: time.Second * 2,
		},
	}

	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_add_role@test.com",
			},
			Roles: []string{"PARTICIPANT"},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.AddRoleForUser(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.RoleMsg{}
		_, err := s.AddRoleForUser(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non admin user", func(t *testing.T) {
		req := &api.RoleMsg{
			Token: &api.TokenInfos{
				Id:         "testuserid",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles": "PARTICIPANT",
				},
			},
			AccountId: testUsers[0].Account.AccountID,
			Role:      "ADMIN",
		}
		_, err := s.AddRoleForUser(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "permission denied")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with valid arguments", func(t *testing.T) {
		req := &api.RoleMsg{
			Token: &api.TokenInfos{
				Id:         "testuserid",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles": "PARTICIPANT,ADMIN",
				},
			},
			AccountId: testUsers[0].Account.AccountID,
			Role:      "ADMIN",
		}
		resp, err := s.AddRoleForUser(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Roles) != 2 || resp.Roles[1] != "ADMIN" {
			t.Errorf("unexpected response: %s", resp)
			return
		}
	})

	t.Run("with already added role", func(t *testing.T) {
		req := &api.RoleMsg{
			Token: &api.TokenInfos{
				Id:         "testuserid",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles": "PARTICIPANT,ADMIN",
				},
			},
			AccountId: testUsers[0].Account.AccountID,
			Role:      "ADMIN",
		}
		_, err := s.AddRoleForUser(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "role already added")
		if !ok {
			t.Error(msg)
		}
	})

}

func TestRemoveRoleForUserEndpoint(t *testing.T) {
	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenMinimumAgeMin:  time.Second * 1,
			TokenExpiryInterval: time.Second * 2,
		},
	}

	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_remove_role@test.com",
			},
			Roles: []string{"PARTICIPANT", "RESEARCHER"},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.RemoveRoleForUser(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.RoleMsg{}
		_, err := s.RemoveRoleForUser(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non admin user", func(t *testing.T) {
		req := &api.RoleMsg{
			Token: &api.TokenInfos{
				Id:         "testuserid",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles": "PARTICIPANT",
				},
			},
			AccountId: testUsers[0].Account.AccountID,
			Role:      "RESEARCHER",
		}
		_, err := s.RemoveRoleForUser(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "permission denied")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with valid arguments", func(t *testing.T) {
		req := &api.RoleMsg{
			Token: &api.TokenInfos{
				Id:         "testuserid",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles": "PARTICIPANT,ADMIN",
				},
			},
			AccountId: testUsers[0].Account.AccountID,
			Role:      "RESEARCHER",
		}
		resp, err := s.RemoveRoleForUser(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Roles) != 1 && resp.Roles[0] != "PARTICIPANT" {
			t.Errorf("unexpected response: %s", resp)
			return
		}
	})

	t.Run("with already non existing role", func(t *testing.T) {
		req := &api.RoleMsg{
			Token: &api.TokenInfos{
				Id:         "testuserid",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles": "PARTICIPANT,ADMIN",
				},
			},
			AccountId: testUsers[0].Account.AccountID,
			Role:      "RESEARCHER",
		}
		_, err := s.RemoveRoleForUser(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "role not found")
		if !ok {
			t.Error(msg)
		}
	})
}

func TestFindNonParticipantUsersEndpoint(t *testing.T) {
	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenMinimumAgeMin:  time.Second * 1,
			TokenExpiryInterval: time.Second * 2,
		},
	}

	_, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_findingusers_1@test.com",
			},
			Roles: []string{"PARTICIPANT", "RESEARCHER"},
		},
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_findingusers_2@test.com",
			},
			Roles: []string{"PARTICIPANT", "ADMIN"},
		},
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_findingusers_3@test.com",
			},
			Roles: []string{"PARTICIPANT"},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.FindNonParticipantUsers(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.FindNonParticipantUsersMsg{}
		_, err := s.FindNonParticipantUsers(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non admin user", func(t *testing.T) {
		req := &api.FindNonParticipantUsersMsg{
			Token: &api.TokenInfos{
				Id:         "testuserid",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles": "PARTICIPANT",
				},
			},
		}
		_, err := s.FindNonParticipantUsers(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "permission denied")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with valid arguments", func(t *testing.T) {
		req := &api.FindNonParticipantUsersMsg{
			Token: &api.TokenInfos{
				Id:         "testuserid",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles": "PARTICIPANT,ADMIN",
				},
			},
		}
		resp, err := s.FindNonParticipantUsers(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		for _, u := range resp.Users {
			if len(u.Roles) == 1 && u.Roles[0] == "PARTICIPANT" {
				t.Errorf("unexpected user: %s", u)
			}
		}
	})
}

func TestStreamUsersEndpoint(t *testing.T) {
	// create users

	// call with emtpy req
	// with missing req
	// with valid req
	t.Error("test unimplemented")
}
