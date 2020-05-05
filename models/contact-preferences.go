package models

import api "github.com/influenzanet/user-management-service/api"

// ContactPreferences defines how to reach out to the user for what purpose
type ContactPreferences struct {
	SubscribedToNewletter bool     `bson:"subscribedToNewletter"`
	SendNewsletterTo      []string `bson:"sendNewsletterTo"`
}

func ContactPreferencesFromAPI(obj *api.ContactPreferences) ContactPreferences {
	if obj == nil {
		return ContactPreferences{}
	}

	res := ContactPreferences{
		SubscribedToNewletter: obj.SubscribedToNewletter,
		SendNewsletterTo:      obj.SendNewsletterTo,
	}
	return res
}

// ToAPI converts a person from DB format into the API format
func (obj ContactPreferences) ToAPI() *api.ContactPreferences {
	return &api.ContactPreferences{
		SubscribedToNewletter: obj.SubscribedToNewletter,
		SendNewsletterTo:      obj.SendNewsletterTo,
	}
}
