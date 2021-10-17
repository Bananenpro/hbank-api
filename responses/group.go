package responses

import "github.com/Bananenpro/hbank-api/models"

type CreateGroupSuccess struct {
	Base
	Id string `json:"id"`
}

type Group struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	GroupPictureId string `json:"group_picture_id"`
}

func NewGroups(groups []models.Group) interface{} {
	groupDTOs := make([]Group, len(groups))
	for i, g := range groups {
		groupDTOs[i].Id = g.Id.String()
		groupDTOs[i].Name = g.Name
		groupDTOs[i].Description = g.Description
		groupDTOs[i].GroupPictureId = g.GroupPictureId.String()
	}

	type groupsResp struct {
		Base
		Groups []Group `json:"groups"`
	}

	return groupsResp{
		Base: Base{
			Success: true,
		},
		Groups: groupDTOs,
	}
}

func NewGroup(group *models.Group) interface{} {
	type groupResp struct {
		Base
		Group
	}

	return groupResp{
		Base: Base{
			Success: true,
		},
		Group: Group{
			Id:             group.Id.String(),
			Name:           group.Name,
			Description:    group.Description,
			GroupPictureId: group.GroupPictureId.String(),
		},
	}
}
