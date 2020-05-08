package tokens

import "testing"

func TestGetRolesFromPayload(t *testing.T) {
	t.Run("with empty payload", func(t *testing.T) {
		p := map[string]string{}
		roles := GetRolesFromPayload(p)
		if len(roles) > 0 {
			t.Errorf("something went wrong: %s", roles)
		}
	})

	t.Run("with missing role field", func(t *testing.T) {
		p := map[string]string{
			"test": "testRole",
		}
		roles := GetRolesFromPayload(p)
		if len(roles) > 0 {
			t.Errorf("something went wrong: %s", roles)
		}
	})
	t.Run("with one role", func(t *testing.T) {
		p := map[string]string{
			"roles": "testRole1",
		}
		roles := GetRolesFromPayload(p)
		if len(roles) != 1 && roles[0] != "testRole1" {
			t.Errorf("something went wrong: %s", roles)
		}
	})

	t.Run("with multiple roles", func(t *testing.T) {
		p := map[string]string{
			"roles": "testRole1,testRole2,testRole3",
		}
		roles := GetRolesFromPayload(p)
		if len(roles) != 3 && roles[0] != "testRole1" {
			t.Errorf("something went wrong: %s", roles)
		}
	})
}
