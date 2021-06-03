package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/influenzanet/user-management-service/pkg/models"

	"github.com/influenzanet/go-utils/pkg/constants"

	"github.com/influenzanet/user-management-service/pkg/pwhash"
	"github.com/influenzanet/user-management-service/pkg/utils"

	"github.com/influenzanet/user-management-service/internal/config"
	"github.com/influenzanet/user-management-service/pkg/dbs/userdb"
)

const userManagementTimerEventFrequency = 90 * 60 // seconds

type UserRequest struct {
	email      string `yaml:"email"`
	password   string `yaml:"password"`
	instanceID string `yaml:"instance"`
}

func terminate(message string) {
	fmt.Println(message)
	os.Exit(1)
}

func reqfromCLI() (UserRequest, error) {
	req := UserRequest{}
	emailF := flag.String("email", "", "email")
	passwordF := flag.String("password", "", "password")
	instanceF := flag.String("instance", "", "instance")

	flag.Parse()

	req.email = *emailF
	req.password = *passwordF
	req.instanceID = *instanceF

	if req.email == "" {
		terminate("email must be provided")
	}

	if req.password == "" {
		terminate("password must be provided")
	}

	if req.instanceID == "" {
		terminate("instance must be provided")
	}

	return req, nil
}

func main() {

	req, err := reqfromCLI()
	if err != nil {
		terminate(err.Error())
	}

	conf := config.InitConfig()

	userDBService := userdb.NewUserDBService(conf.UserDBConfig)

	req.email = utils.SanitizeEmail(req.email)

	if !utils.CheckEmailFormat(req.email) {
		terminate("account id not a valid email")
	}
	if !utils.CheckPasswordFormat(req.password) {
		terminate("password too weak")
	}

	password, err := pwhash.HashPassword(req.password)
	if err != nil {
		terminate(err.Error())
	}

	newUser := models.User{
		Account: models.Account{
			Type:               "email",
			AccountID:          req.email,
			AccountConfirmedAt: time.Now().Unix(), // not confirmed yet
			Password:           password,
			PreferredLanguage:  "fr",
		},
		Roles: []string{constants.USER_ROLE_ADMIN},
		Profiles: []models.Profile{
			{
				ID:                 primitive.NewObjectID(),
				Alias:              utils.BlurEmailAddress(req.email),
				AvatarID:           "default",
				ConsentConfirmedAt: time.Now().Unix(),
				MainProfile:        true,
			},
		},
		Timestamps: models.Timestamps{
			CreatedAt: time.Now().Unix(),
		},
	}
	newUser.AddNewEmail(req.email, false)
	newUser.ContactPreferences.SubscribedToNewsletter = false
	newUser.ContactPreferences.SendNewsletterTo = []string{newUser.ContactInfos[0].ID.Hex()}
	newUser.ContactPreferences.SubscribedToWeekly = false
	newUser.ContactPreferences.ReceiveWeeklyMessageDayOfWeek = int32(rand.Intn(7))

	instanceID := req.instanceID
	id, err := userDBService.AddUser(instanceID, newUser)
	if err != nil {
		terminate(err.Error())
	}
	newUser.ID, _ = primitive.ObjectIDFromHex(id)

	fmt.Println("User created with id : %s", id)
}
