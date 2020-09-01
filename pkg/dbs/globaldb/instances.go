package globaldb

import (
	"github.com/influenzanet/go-utils/pkg/global_types"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbService *GlobalDBService) GetAllInstances() ([]global_types.Instance, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	cur, err := dbService.collectionRefInstances().Find(
		ctx,
		filter,
	)

	if err != nil {
		return []global_types.Instance{}, err
	}
	defer cur.Close(ctx)

	instances := []global_types.Instance{}
	for cur.Next(ctx) {
		var result global_types.Instance
		err := cur.Decode(&result)
		if err != nil {
			return instances, err
		}

		instances = append(instances, result)
	}
	if err := cur.Err(); err != nil {
		return instances, err
	}

	return instances, nil
}
